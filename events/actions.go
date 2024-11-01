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

var _chan_SafetyFailure <-chan bool

func AssertSafety(id string, fn TestConditionFunction, timeAllowed time.Duration) {
	_safetyAsserts = append(_safetyAsserts, SafetyAssert{
		Condition:   fn,
		AllowedTime: timeAllowed,
		C:           _chan_SafetyFailure,
		assert: 0,
	})
}

func AssertUntil(id string, fn TestConditionFunction, timeout time.Duration) chan bool {
	c := make(chan bool)
	_untilAsserts = append(_untilAsserts, WaitFor{
		Condition: fn,
		Timeout:   timeout,
		C:         c,
	})
	return c
}

// On the arrival of a new trigger, check the loaded events and see if
// any of them are listening on the current trigger. If yes,
func pollEvents(triggerType Trigger, triggerParams interface{}) {
	fmt.Println("Polling events of type", triggerType, "params: ", triggerParams)
	for i := range _safetyAsserts {
		event := &_safetyAsserts[i]
		if event.IsAsserted() && event.Condition(_elevatorStates) {
			event.Abort()
			fmt.Println("safety assert aborted.")
		} else if !event.Condition(_elevatorStates) {
			fmt.Println(i, "safety assert failed!")
			event.Assert()
		}
	}

	for i, event := range _untilAsserts {
		if event.Condition(_elevatorStates) {
			fmt.Println("until assert happened!")
			event.Trigger()
			_untilAsserts = append(_untilAsserts[:i], _untilAsserts[i+1:]...)
		}
	}
}

func EventListener(
	simulatedElevators []fatelevator.SimulatedElevator,
	chan_SafetyFailure <-chan bool) {
	//Initialize bindings between event listeners, and setup our perspective of the elevator states.
	_simulatedElevators = simulatedElevators
	_chan_SafetyFailure = chan_SafetyFailure
	for i := range _simulatedElevators {
		_elevatorStates = append(_elevatorStates, InitElevatorState(elevio.N_FLOORS))
		go listenToElevators(i, &_simulatedElevators[i])
	}
}

func listenToElevators(elevatorId int, simulatedElevator *fatelevator.SimulatedElevator) {
	//Process signals from simulated elevators.
	//In response, poll active events for triggers, and update the local state.
	for {
		select {
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
		}
	}
}
