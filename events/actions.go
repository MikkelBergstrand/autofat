package events

import (
	"autofat/elevio"
	"autofat/fatelevator"
	"fmt"
	"time"
)

var _asserts map[string]AssertObj
var _awaits map[string]AwaitObj

var _elevatorStates []ElevatorState
var _testId string

var _chan_Kill chan bool
var _pollAgain chan TriggerMessage

func Assert(id string, fn TestConditionFunction, timeAllowed time.Duration) {
	_asserts[id] = AssertObj{
		Condition:   fn,
		AllowedTime: timeAllowed,
		assert:      0,
		Data: EventMetadata{
			Id:     id,
			TestId: _testId,
		},
	}
}

func Disassert(id string) {
	delete(_asserts, id)
}

func Await(id string, fn TestConditionFunction, timeout time.Duration) error {
	wait_for := AwaitObj{
		Data: EventMetadata{
			Id:     id,
			TestId: _testId,
		},
		Condition:     fn,
		Timeout:       timeout,
		chan_internal: make(chan EventMetadata),
	}

	//Check if immediately true
	if wait_for.Condition(_elevatorStates) {
		go func() {
			wait_for.C_OK <- true
		}()
		fmt.Println("AssertUntil: ", id, "true upon added to system.")
		return nil
	}

	_awaits[id] = wait_for
	go AwaitWatchdog(id)

	fmt.Println("AssertUntil: ", id, "Added to system")

	data := <-wait_for.chan_internal
	fmt.Println("Await event was heard from: ", data)
	if data.Timeout {
		return fmt.Errorf("%s timeout", data.Id)
	}
	return nil
}

// On the arrival of a new trigger, check the loaded events and see if
// any of them are listening on the current trigger. If yes,
func pollEvents(triggerType Trigger, triggerParams interface{}) {
	fmt.Println("Polling events of type", triggerType, "params: ", triggerParams, "u: ", len(_awaits))
	for i := range _asserts {
		if _asserts[i].IsAsserted() && _asserts[i].Condition(_elevatorStates) {
			_asserts[i] = _asserts[i].Abort()
		} else if !_asserts[i].Condition(_elevatorStates) {
			_asserts[i] = _asserts[i].Assert()

			go func() {
				assert_at_beginning := _asserts[i].assert
				timer := time.NewTimer(_asserts[i].AllowedTime)
				<-timer.C
				if _asserts[i].assert == assert_at_beginning {
					fmt.Println("Safety assert ", _asserts[i].Data, "failed!")
					//Fail all waiting awaits
					for id, await := range _awaits {
						delete(_awaits, id)
						await.chan_internal <- EventMetadata{
							Timeout: true,
							Id:      await.Data.Id,
							TestId:  await.Data.TestId,
						}
					}
				}
			}()
		}
	}

	for i := range _awaits {
		if _awaits[i].Condition(_elevatorStates) {
			_awaits[i] = _awaits[i].Trigger()
			delete(_awaits, i)
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
) {
	_chan_Kill = make(chan bool)
	_testId = testId

	_elevatorStates = make([]ElevatorState, 0)
	_asserts = make(map[string]AssertObj)
	_awaits = make(map[string]AwaitObj)

	//First time init
	for i := range fatelevator.Count() {
		_elevatorStates = append(_elevatorStates, InitElevatorState(elevio.N_FLOORS))
		go listenToElevators(i, fatelevator.Get(i))
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
				_elevatorStates[elevatorId].HallDownLights[order_light.Floor] = order_light.Value
			case elevio.BT_HallUp:
				_elevatorStates[elevatorId].HallUpLights[order_light.Floor] = order_light.Value
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
		case new_dir := <-simulatedElevator.Chan_Direction:
			_elevatorStates[elevatorId].Direction = new_dir
			_pollAgain <- TriggerMessage{
				Type:   TRIGGER_DIRECTION,
				Params: fmt.Sprintf("Elevator %d, dir %s", elevatorId, new_dir.String()),
			}
		case <-simulatedElevator.Chan_Outofbounds:
			//Fail instantly when elevator reaches out of bounds
			fmt.Println("Out of bounds detected for elevator", elevatorId)
		}
	}
}

func Kill() {
	for range _elevatorStates {
		_chan_Kill <- true
	}

	for id, v := range _awaits {
		_awaits[id] = v.Delete()
		delete(_awaits, id)
	}
}
