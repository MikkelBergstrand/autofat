package tests

import (
	"autofat/elevio"
	"autofat/events"
)

type TestParams struct {
	InitialFloors []int //The size of the array denotes the number of active elevators to use.
	EventList     []events.Event
}

// Single elevator test.
// Test if lamp lights up when we arrive at a floor.
func TestFloorLamp() TestParams {
	eventList := []events.Event{
		{
			ID:            "init",
			Description:   "Initial event",
			TriggerType:   events.TRIGGER_INIT,
			LoadOnTrigger: []string{"floor_light_0"},
		},
		{
			ID:          "floor_light_0",
			Description: "Await floor 0 light",
			TriggerType: events.TRIGGER_FLOOR_LIGHT,
			TriggerParams: events.Floor{
				Floor:    0,
				Elevator: 0,
			},
			ActionType: events.ACTION_MAKE_ORDER,
			ActionParams: events.Button{
				Elevator: 0,
				ButtonEvent: elevio.ButtonEvent{
					Button: elevio.BT_Cab,
					Floor:  3,
				},
			},
		},
	}

	return TestParams{
		InitialFloors: []int{0},
		EventList:     eventList,
	}
}
