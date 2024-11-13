package events

import (
	"autofat/elevio"
	"autofat/fatelevator"
	"fmt"
	"time"
)

var _safetyAsserts []SafetyAssert
var _untilAsserts []WaitFor

var _simulatedElevators []fatelevator.SimulatedElevator
var _elevatorStates []ElevatorState

var _chan_SafetyFailure chan<- bool
var _chan_Kill chan bool

func AssertSafety(id string, fn TestConditionFunction, timeAllowed time.Duration) {
	_safetyAsserts = append(_safetyAsserts, SafetyAssert{
		Condition:   fn,
		AllowedTime: timeAllowed,
		C:           _chan_SafetyFailure,
		assert:      0,
	})
}

func AssertUntil(id string, fn TestConditionFunction, timeout time.Duration) chan bool {
	chan_done := make(chan bool)

	wait_for := WaitFor{
		ID:          id,
		Condition:   fn,
		Timeout:     timeout,
		C:           chan_done,
		Chan_Result: _chan_SafetyFailure,
	}

	_untilAsserts = append(_untilAsserts, wait_for)

	go _untilAsserts[len(_untilAsserts)-1].Watchdog()

	fmt.Println("AssertUntil: ", id, "Added to system")
	return chan_done
}

// On the arrival of a new trigger, check the loaded events and see if
// any of them are listening on the current trigger. If yes,
func pollEvents(triggerType Trigger, triggerParams interface{}) {
	fmt.Println("Polling events of type", triggerType, "params: ", triggerParams, "u: ", len(_untilAsserts))
	for i := range _safetyAsserts {
		event := &_safetyAsserts[i]
		if event.IsAsserted() && event.Condition(_elevatorStates) {
			event.Abort()
		} else if !event.Condition(_elevatorStates) {
			event.Assert()
		}
	}

	for i := range _untilAsserts {
		if _untilAsserts[i].Condition(_elevatorStates) {
			_untilAsserts[i].Trigger()
			_untilAsserts = append(_untilAsserts[:i], _untilAsserts[i+1:]...)
		}
	}
}

func EventListener(
	simulatedElevators []fatelevator.SimulatedElevator,
	chan_SafetyFailure chan<- bool) {
	//Initialize bindings between event listeners, and setup our perspective of the elevator states.
	_simulatedElevators = simulatedElevators
	_chan_SafetyFailure = chan_SafetyFailure
	_chan_Kill = make(chan bool)

	_elevatorStates = make([]ElevatorState, 0)
	_safetyAsserts = make([]SafetyAssert, 0)
	_untilAsserts = make([]WaitFor, 0)

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
			pollEvents(TRIGGER_ARRIVE_FLOOR, Floor{
				Floor:    new_floor,
				Elevator: elevatorId,
			})
		case new_floor_light := <-simulatedElevator.Chan_FloorLight:
			_elevatorStates[elevatorId].FloorLamp = new_floor_light
			pollEvents(TRIGGER_FLOOR_LIGHT, Floor{
				Floor:    new_floor_light,
				Elevator: elevatorId,
			})
		case door_state := <-simulatedElevator.Chan_Door:
			_elevatorStates[elevatorId].DoorOpen = door_state
			if door_state {
				pollEvents(TRIGGER_DOOR_OPEN, nil)
			} else {
				pollEvents(TRIGGER_DOOR_CLOSE, nil)
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
			pollEvents(TRIGGER_ORDER_LIGHT, order_light)
		case obstruction := <-simulatedElevator.Chan_Obstruction:
			{
				_elevatorStates[elevatorId].Obstruction = obstruction
				pollEvents(TRIGGER_OBSTRUCTION, obstruction)
			}
		case <-simulatedElevator.Chan_Outofbounds:
			//Fail instantly when elevator reaches out of bounds
			fmt.Println("Out of bounds detected for elevator", elevatorId)
			_chan_SafetyFailure <- false
		}

	}
}

func Kill() {
	//Send a kill signal for each simulated elevator to kill them all.
	for _ = range _simulatedElevators {
		_chan_Kill <- true
	}
}
