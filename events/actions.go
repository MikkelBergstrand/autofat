package events

import (
	"autofat/fatelevator"
	"fmt"
	"log"
	"time"
)


type ElevatorState struct {
	Floor uint8			
}

var _allEvents map[string]Event
var _loadedEvents map[string]*Event
var _simulatedElevators []fatelevator.SimulatedElevator

func loadEvent(event *Event) {
	fmt.Printf("Event %s (%s) has been loaded \n", event.ID, event.Description)
	_loadedEvents[event.ID] = event

	if event.TriggerType == TRIGGER_INIT  || event.TriggerType == TRIGGER_LOAD {
		//LOAD and INIT both trigger as they are queued.
		triggerEvent(event)
	} else if event.TriggerType == TRIGGER_DELAY{
		// Delay trigger: params are milliseconds to wait. Start a goroutine
		// that does nothing but wait for timer, then trigger. 
		go func ()  {
			timer := time.NewTimer(time.Millisecond * time.Duration(event.TriggerParams.(int)))
			<-timer.C
			triggerEvent(event)
		}()
	}

	switch(event.ActionType) {
	case ACTION_MAKE_ORDER:
		params := event.ActionParams.(Button)
		fmt.Printf("Action execute: Elevator %d making order %d %d \n", params.Elevator, params.Button, params.Floor)
		_simulatedElevators[params.Elevator].Chan_ButtonPresser <- params.ButtonEvent
	}
}

func triggerEvent(event *Event) {
	fmt.Printf("Triggered event %s (%s)\n", event.ID, event.Description)
	//Load cascading events
	for _, eventId := range event.LoadOnTrigger {
		toLoad, ok := _allEvents[eventId]
		if !ok {
			log.Fatalf("Unrecognized event id %s\n", eventId)
		}
		loadEvent(&toLoad)
	}

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
					triggerEvent(event)
				}
			}	
		}		
	}
}

func initEvents(events []Event) {
	_loadedEvents = make(map[string]*Event)
	_allEvents = make(map[string]Event)

	for i := range events 	{
		_allEvents[events[i].ID] = events[i]
	}

	for _, event := range _allEvents {
		if event.TriggerType == TRIGGER_INIT {
			loadEvent(&event)
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
