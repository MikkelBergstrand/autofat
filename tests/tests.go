package tests

import (
	"autofat/elevio"
	"autofat/events"
	"fmt"
	"time"
)

// Helper function. We want to evaluate a condition that should be true for all
// elevators in the system.
func assertAll(test func(e events.ElevatorState) bool) func([]events.ElevatorState) bool {
	return func(es []events.ElevatorState) bool {
		for i := range es {
			if !test(es[i]) {
				return false
			}
		}
		return true
	}
}

func waitForInit() {
	fmt.Println("Wait for initial state")

	<-events.AssertUntil("init", assertAll(func(e events.ElevatorState) bool {
		return e.Floor != -1 && !e.DoorOpen
	}), time.Second*10)

	fmt.Println("Initial state succeeded.")
}

func TestFloorLamp() {
	waitForInit()

	events.MakeOrder(0, elevio.BT_Cab, 3)
	events.AssertSafety("floor_light_correct", func(es []events.ElevatorState) bool { return es[0].FloorLamp == es[0].Floor }, time.Second*10)

	<-events.AssertUntil("reached_dest", func(es []events.ElevatorState) bool { return es[0].Floor == 3 && es[0].FloorLamp == 3 }, time.Second*15)
	time.Sleep(1 * time.Second)

}

func TestInitBetweenFloors() {
	waitForInit()
	time.Sleep(1 * time.Second)
}
