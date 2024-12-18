package statemanager

import (
	"autofat/elevio"
	"autofat/simulator"
	"autofat/studentprogram"
	"fmt"
	"time"
)

var _asserts map[string]t_assert
var _awaits map[string]t_await

var _elevatorStates []ElevatorState
var _testId string

var _chan_Kill chan bool
var _chan_Terminated chan bool
var _pollAgain chan triggerMessage

func Assert(id string, fn TestConditionFunction, timeAllowed time.Duration) {
	_asserts[id] = t_assert{
		Condition:   fn,
		AllowedTime: timeAllowed,
		assert:      0,
		Data: t_eventData{
			Id:     id,
			TestId: _testId,
		},
	}
}

func Disassert(id string) {
	delete(_asserts, id)
}

func Await(id string, fn TestConditionFunction, timeout time.Duration) error {
	wait_for := t_await{
		Data: t_eventData{
			Id:     id,
			TestId: _testId,
		},
		Condition:     fn,
		Timeout:       timeout,
		chan_internal: make(chan t_eventData),
	}

	//Check if immediately true
	if wait_for.Condition(_elevatorStates) {
		fmt.Println("Await: ", id, "true upon added to system.")
		return nil
	}

	_awaits[id] = wait_for
	go awaitWatchDog(id)

	fmt.Println("Await: ", id, "Added to system")

	data := <-wait_for.chan_internal
	fmt.Println("Await event was heard from: ", data)
	if data.Timeout {
		return fmt.Errorf("%s timeout", data.Id)
	}
	return nil
}

// On the arrival of a new trigger, check the loaded events and see if
// any of them are listening on the current trigger. If yes,
func pollEvents(triggerType trigger, triggerParams interface{}) {
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
						await.chan_internal <- t_eventData{
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
	_pollAgain = make(chan triggerMessage)

	go func() {
		for {
			trigger := <-_pollAgain
			pollEvents(trigger.Type, trigger.Params)
		}
	}()
}

func EventListener(
	testId string,
) {
	_chan_Kill = make(chan bool)
	_chan_Terminated = make(chan bool)
	_testId = testId

	_elevatorStates = make([]ElevatorState, 0)
	_asserts = make(map[string]t_assert)
	_awaits = make(map[string]t_await)

	//First time init
	for i := range simulator.Count() {
		_elevatorStates = append(_elevatorStates, InitElevatorState(elevio.N_FLOORS))
		go listenToElevators(i, simulator.Get(i), studentprogram.Get(i))
	}
}

func listenToElevators(elevatorId int, simulatedElevator *simulator.Simulator, studentProgram studentprogram.StudentProgram) {
	//Process signals from simulated elevators.
	//In response, poll active events for triggers, and update the local state.
	for {
		select {
		case <-_chan_Kill:
			{
				fmt.Println("Killed elevator listener ", elevatorId)
				_chan_Terminated <- true
				return
			}
		case new_floor := <-simulatedElevator.Chan_FloorSensor:
			_elevatorStates[elevatorId].Floor = new_floor
			_pollAgain <- triggerMessage{
				Type:   TRIGGER_ARRIVE_FLOOR,
				Params: fmt.Sprintf("floor=%d", new_floor),
			}
		case new_floor_light := <-simulatedElevator.Chan_FloorLight:
			_elevatorStates[elevatorId].FloorLamp = new_floor_light
			_pollAgain <- triggerMessage{
				Type:   TRIGGER_FLOOR_LIGHT,
				Params: fmt.Sprintf("floor=%d", new_floor_light),
			}
		case door_state := <-simulatedElevator.Chan_Door:
			_elevatorStates[elevatorId].DoorOpen = door_state
			_pollAgain <- triggerMessage{
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
			_pollAgain <- triggerMessage{
				Type:   TRIGGER_ORDER_LIGHT,
				Params: order_light,
			}
		case obstruction := <-simulatedElevator.Chan_Obstruction:
			_elevatorStates[elevatorId].Obstruction = obstruction
			_pollAgain <- triggerMessage{
				Type:   TRIGGER_OBSTRUCTION,
				Params: obstruction,
			}
		case new_dir := <-simulatedElevator.Chan_Direction:
			_elevatorStates[elevatorId].Direction = new_dir
			_pollAgain <- triggerMessage{
				Type:   TRIGGER_DIRECTION,
				Params: fmt.Sprintf("Elevator %d, dir %s", elevatorId, new_dir.String()),
			}
		case <-simulatedElevator.Chan_Outofbounds:
			//Fail instantly when elevator reaches out of bounds
			_elevatorStates[elevatorId].Outofbounds = true
			_pollAgain <- triggerMessage{
				Type:   TRIGGER_OOB,
				Params: elevatorId,
			}
		case <-studentProgram.Chan_Crash:
			_elevatorStates[elevatorId].Status = studentprogram.CRASHED
			_pollAgain <- triggerMessage{
				Type:   TRIGGER_CRASH,
				Params: elevatorId,
			}
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
	<-_chan_Terminated
}
