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

var _chan_SafetyFailure chan<- bool
var _chan_Kill chan bool
var _pollAgain chan TriggerMessage
var _untilTimeoutHandler chan EventMetadata

func AssertSafety(id string, fn TestConditionFunction, timeAllowed time.Duration) {
	_safetyAsserts[id] = SafetyAssert{
		Condition:   fn,
		AllowedTime: timeAllowed,
		C:           _chan_SafetyFailure,
		assert:      0,
	}
	_pollAgain <- TriggerMessage{
		Type:   TRIGGER_SAFETYASSERT,
		Params: _safetyAsserts[id],
	}
}

func AssertUntil(id string, fn TestConditionFunction, timeout time.Duration) chan bool {
	chan_done := make(chan bool)

	wait_for := WaitFor{
		Data: EventMetadata{
			Id:     id,
			TestId: _testId,
		},
		Condition:    fn,
		Timeout:      timeout,
		Chan_OK:      chan_done,
		Chan_Timeout: _untilTimeoutHandler,
	}

	//Check if immediately true
	if wait_for.Condition(_elevatorStates) {
		//Bit hacky.
		go func() {
			chan_done <- true
		}()

		fmt.Println("AssertUntil: ", id, "true upon added to system.")
		return chan_done
	}

	go AwaitWatchdog(id)

	fmt.Println("AssertUntil: ", id, "Added to system")
	_untilAsserts[id] = wait_for

	return chan_done
}

// On the arrival of a new trigger, check the loaded events and see if
// any of them are listening on the current trigger. If yes,
func pollEvents(triggerType Trigger, triggerParams interface{}) {
	fmt.Println("Polling events of type", triggerType, "params: ", triggerParams, "u: ", len(_untilAsserts))
	for i := range _safetyAsserts {
		event := _safetyAsserts[i]
		if event.IsAsserted() && event.Condition(_elevatorStates) {
			_safetyAsserts[i] = event.Abort()
		} else if !event.Condition(_elevatorStates) {
			_safetyAsserts[i] = event.Assert()
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
	chan_SafetyFailure chan<- bool) {
	//Initialize bindings between event listeners, and setup our perspective of the elevator states.
	_simulatedElevators = simulatedElevators
	_chan_SafetyFailure = chan_SafetyFailure
	_chan_Kill = make(chan bool)
	_testId = testId
	_active = true
	_untilTimeoutHandler = make(chan EventMetadata)

	_elevatorStates = make([]ElevatorState, 0)
	_safetyAsserts = make(map[string]SafetyAssert)
	_untilAsserts = make(map[string]WaitFor)

	//First time init
	for i := range _simulatedElevators {
		_elevatorStates = append(_elevatorStates, InitElevatorState(elevio.N_FLOORS))
		go listenToElevators(i, &_simulatedElevators[i])
	}

	//Handle timeouts of Until events
	//We should only handle it once, so subsequent timeout events will be ignored.
	//Furthermore, we ignore timeouts set by past tests by checking the test Id.
	go func() {
		firedOnce := false
		for {
			select {
			case event := <-_untilTimeoutHandler:
				if !firedOnce && _testId == event.TestId {
					firedOnce = true
					_chan_SafetyFailure <- true
				}
			}
		}
	}()
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
			_chan_SafetyFailure <- false
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
