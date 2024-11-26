package events

import (
	"autofat/elevio"
	"autofat/fatelevator"
	"fmt"
	"time"
)

var _safetyAsserts map[string]SafetyAssert
var _untilAsserts map[string]WaitFor

var _simulatedElevators []fatelevator.SimulatedElevator
var _elevatorStates []ElevatorState
var _testId string
var _active bool

var _chan_Kill chan bool
var _pollAgain chan TriggerMessage

func AssertSafety(id string, fn TestConditionFunction, timeAllowed time.Duration, output chan bool) {
	_safetyAsserts[id] = SafetyAssert{
		Condition:   fn,
		AllowedTime: timeAllowed,
		assert:      0,
		Data: EventMetadata{
			Id:     id,
			TestId: _testId,
		},
	}

}

func AssertUntil(id string, fn TestConditionFunction, timeout time.Duration) (chan bool, chan bool) {
	wait_for := WaitFor{
		Data: EventMetadata{
			Id:     id,
			TestId: _testId,
		},
		Condition:     fn,
		Timeout:       timeout,
		chan_internal: make(chan EventMetadata),
		C_OK:          make(chan bool),
		C_Timeout:     make(chan bool),
	}

	//Check if immediately true
	if wait_for.Condition(_elevatorStates) {
		go func() {
			wait_for.C_OK <- true
		}()
		fmt.Println("AssertUntil: ", id, "true upon added to system.")
		return wait_for.C_OK, wait_for.C_Timeout
	}


	_untilAsserts[id] = wait_for
	go AwaitWatchdog(id)

	fmt.Println("AssertUntil: ", id, "Added to system")

	go func() bool {
		triggered := false
		for {
			select {
			case data := <-wait_for.chan_internal:
				fmt.Println("Await event was heard from: ", data)
				if !triggered && _active && _testId == data.TestId {
					triggered = true
					if data.Timeout {
						wait_for.C_Timeout <- true
					} else {
						wait_for.C_OK <- true
					}
				} else {
					fmt.Println("Subsequent Await event: ", data, " was ignored. It is not active.")
				}
			}
		}
	}()

	return wait_for.C_OK, wait_for.C_Timeout
}

// On the arrival of a new trigger, check the loaded events and see if
// any of them are listening on the current trigger. If yes,
func pollEvents(triggerType Trigger, triggerParams interface{}) {
	fmt.Println("Polling events of type", triggerType, "params: ", triggerParams, "u: ", len(_untilAsserts))
	for i := range _safetyAsserts {
		if _safetyAsserts[i].IsAsserted() && _safetyAsserts[i].Condition(_elevatorStates) {
			_safetyAsserts[i] = _safetyAsserts[i].Abort()
		} else if !_safetyAsserts[i].Condition(_elevatorStates) {
			_safetyAsserts[i] = _safetyAsserts[i].Assert()

			go func() {
				assert_at_beginning := _safetyAsserts[i].assert
				timer := time.NewTimer(_safetyAsserts[i].AllowedTime)
				<-timer.C
				if _safetyAsserts[i].assert == assert_at_beginning {
					fmt.Println("Safety assert ", _safetyAsserts[i].Data, "failed!")
					_safetyAsserts[i].C <- false
				}
			}()
		}
	}

	for i := range _untilAsserts {
		if _untilAsserts[i].Condition(_elevatorStates) {
			_untilAsserts[i] = _untilAsserts[i].Trigger()
			delete(_untilAsserts, i)
		}
	}
}

func Init() {
	_pollAgain = make(chan TriggerMessage)

	go func() {
		for {
			select {
			case trigger := <-_pollAgain:
				pollEvents(trigger.Type, trigger.Params)
			}
		}
	}()
}

func EventListener(
	testId string,
	simulatedElevators []fatelevator.SimulatedElevator,
) {
	//Initialize bindings between event listeners, and setup our perspective of the elevator states.
	_simulatedElevators = simulatedElevators
	_chan_Kill = make(chan bool)
	_testId = testId
	_active = true

	_elevatorStates = make([]ElevatorState, 0)
	_safetyAsserts = make(map[string]SafetyAssert)
	_untilAsserts = make(map[string]WaitFor)

	//First time init
	for i := range _simulatedElevators {
		_elevatorStates = append(_elevatorStates, InitElevatorState(elevio.N_FLOORS))
		go listenToElevators(i, &_simulatedElevators[i])
	}
}

func MakeOrder(elevator int, orderType elevio.ButtonType, floor int) {
	fmt.Println("Making order to elevator", elevator, "type", orderType, "floor: ", floor)

	_simulatedElevators[elevator].Chan_ButtonPresser <- elevio.ButtonEvent{
		Button: orderType,
		Floor:  floor,
	}
}

func listenToElevators(elevatorId int, simulatedElevator *fatelevator.SimulatedElevator) {
	//Process signals from simulated elevators.
	//In response, poll active events for triggers, and update the local state.
	for {
		select {
		case <-_chan_Kill:
			{
				fmt.Println("Killed elevator listener ", elevatorId)
				return
			}
		case new_floor := <-simulatedElevator.Chan_FloorSensor:
			_elevatorStates[elevatorId].Floor = new_floor
			_pollAgain <- TriggerMessage{
				Type: TRIGGER_ARRIVE_FLOOR,
				Params: Floor{
					Floor:    new_floor,
					Elevator: elevatorId,
				},
			}
		case new_floor_light := <-simulatedElevator.Chan_FloorLight:
			_elevatorStates[elevatorId].FloorLamp = new_floor_light
			_pollAgain <- TriggerMessage{
				Type: TRIGGER_FLOOR_LIGHT,
				Params: Floor{
					Floor:    new_floor_light,
					Elevator: elevatorId,
				}}
		case door_state := <-simulatedElevator.Chan_Door:
			_elevatorStates[elevatorId].DoorOpen = door_state
			_pollAgain <- TriggerMessage{
				Type:   TRIGGER_DOOR,
				Params: door_state,
			}
		case order_light := <-simulatedElevator.Chan_OrderLights:
			switch order_light.Button {
			case elevio.BT_Cab:
				_elevatorStates[elevatorId].CabLights[order_light.Floor] = order_light.Value
			case elevio.BT_HallDown:
				_elevatorStates[elevatorId].HallUpLights[order_light.Floor] = order_light.Value
			case elevio.BT_HallUp:
				_elevatorStates[elevatorId].HallDownLights[order_light.Floor] = order_light.Value
			}
			_pollAgain <- TriggerMessage{
				Type:   TRIGGER_ORDER_LIGHT,
				Params: order_light,
			}
		case obstruction := <-simulatedElevator.Chan_Obstruction:
			_elevatorStates[elevatorId].Obstruction = obstruction
			_pollAgain <- TriggerMessage{
				Type:   TRIGGER_OBSTRUCTION,
				Params: obstruction,
			}
		case <-simulatedElevator.Chan_Outofbounds:
			//Fail instantly when elevator reaches out of bounds
			fmt.Println("Out of bounds detected for elevator", elevatorId)
		}
	}
}

func Kill() {
	_active = false
	for id, v := range _untilAsserts {
		_untilAsserts[id] = v.Delete()
		delete(_untilAsserts, id)
	}
	//Send a kill signal for each simulated elevator to kill them all.
	for i := 0; i < len(_simulatedElevators); i++ {
		_chan_Kill <- true
	}
}
