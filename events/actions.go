package events

import (
	"autofat/fatelevator"
	"fmt"
)


type ElevatorState struct {
	Floor uint8			
}



var _allEvents []Event
var _loadedEvents map[string]*Event
var _simulatedElevators []fatelevator.SimulatedElevator

func loadEvent(event *Event) {
	fmt.Printf("Event %s (%s) has been loaded \n", event.ID, event.Description)
	_loadedEvents[event.ID] = event

	switch(event.ActionType) {
	case ACTION_MAKE_ORDER:
		params := event.ActionParams.(Button)
		fmt.Printf("Action execute: Elevator %d making order %d %d \n", params.Elevator, params.Button, params.Floor)
		_simulatedElevators[params.Elevator].Chan_ButtonPresser <- params.ButtonEvent
	}
}

func startEvent(event *Event) {
}

//On the arrival of a new trigger, check the loaded events and see if
//any of them are listening on the current trigger. If yes, 
func pollEvents(triggerType byte, triggerParams interface{}) {
	fmt.Println("Polling events of type", triggerType, "params: ", triggerParams)
	for _, event := range _loadedEvents {
		if event.ActionType == triggerType {
			switch event.ActionType {
			case ACTION_MAKE_ORDER:
				if (event.ActionParams.(Button).Equals(event.TriggerParams.(Button))) {
					startEvent(event)
				}
			}	
		}		
	}
}

func initEvents(events []Event) {
	_loadedEvents = make(map[string]*Event)

	_allEvents = events
	for i := range _allEvents 	{
		if events[i].TriggerType == TRIGGER_INIT {
			loadEvent(&events[i])
		}
	}
}

func EventListener(
	simulatedElevators []fatelevator.SimulatedElevator, 
	events []Event,
	) {
	
	_simulatedElevators = simulatedElevators
	for i := range _simulatedElevators {
		fmt.Println(&_simulatedElevators[i])
		go listenToElevators(i, &_simulatedElevators[i])	
	}
	
	initEvents(events)	
}

func listenToElevators(elevatorId int, simulatedElevator *fatelevator.SimulatedElevator) {
	for {
		select {
		case new_floor := <-simulatedElevator.Chan_FloorSensor:
			pollEvents(TRIGGER_ARRIVE_FLOOR, Floor {
				Floor: new_floor,
				Elevator: elevatorId,
			})	
			case new_floor_light := <-simulatedElevator.Chan_FloorLight:
				pollEvents(TRIGGER_FLOOR_LIGHT, Floor {
					Floor: new_floor_light,
					Elevator: elevatorId,
				})
		}
	}			
}
