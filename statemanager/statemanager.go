package statemanager

import (
	"autofat/elevio"
	"autofat/simulator"
	"autofat/studentprogram"
	"fmt"
)

const BUFFER_SIZE = 10

type States []ElevatorState
type StateChannel chan States

var _elevatorStates States

var _chan_Kill chan bool
var _chan_Terminated chan bool
var _pollAgain chan triggerMessage

var _chan_addListener chan StateChannel
var _chan_removeListener chan StateChannel

var _stateChannels []StateChannel

// On the arrival of a new trigger, check the loaded events and see if
// any of them are listening on the current trigger. If yes,
func pollEvents(triggerType trigger, triggerParams interface{}) {
	fmt.Println("Polling events of type", triggerType, "params: ", triggerParams)

	for _, stateChan := range _stateChannels {
		fmt.Println("Sending state...")
		stateChan <- _elevatorStates
		fmt.Println("Done")
	}

}

func RegisterStateChannel() StateChannel {
	ret := make(StateChannel, BUFFER_SIZE)
	fmt.Println("Adding listener")
	_chan_addListener <- ret
	return ret
}

func UnregisterStateChannel(ch StateChannel) {
	fmt.Println("Removing new listener")
	_chan_removeListener <- ch
}

func Init() {
	_pollAgain = make(chan triggerMessage)
	_chan_addListener = make(chan StateChannel)
	_chan_removeListener = make(chan StateChannel)
	go func() {
		for {
			select {
			case trigger, more := <-_pollAgain:
				if !more {
					for i := range _stateChannels {
						fmt.Println("Closing state channel", i, "of", len(_stateChannels))
						close(_stateChannels[i])
					}
					return
				}
				pollEvents(trigger.Type, trigger.Params)
			case ch := <-_chan_addListener:
				_stateChannels = append(_stateChannels, ch)
			case ch := <-_chan_removeListener:
				close(ch)
				idx := -1
				for i, ch := range _stateChannels {
					if _stateChannels[i] == ch {
						idx = i
					}
				}
				if idx >= 0 {
					_stateChannels = append(_stateChannels[:idx], _stateChannels[idx+1:]...)
				}
			}
		}
	}()
}

func EventListener(
	testId string,
) {
	_chan_Kill = make(chan bool)
	_chan_Terminated = make(chan bool)

	_elevatorStates = make([]ElevatorState, 0)

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
	close(_pollAgain)

	for i := range _elevatorStates {
		fmt.Println("Closing elev poll channel", i)
		_chan_Kill <- true
	}

	for range _elevatorStates {
		<-_chan_Terminated
	}
}
