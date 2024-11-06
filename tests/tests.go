package tests

import (
	"autofat/events"
	"autofat/fatelevator"
	"fmt"
	"time"
)

func TestFloorLamp(e []fatelevator.SimulatedElevator) bool {
	timeout := make(chan bool)
	success := make(chan bool)
	go events.EventListener(e, timeout)

	go func() {
		c := events.AssertUntil("init", func(es []events.ElevatorState) bool { return es[0].Floor != -1 && !es[0].DoorOpen }, time.Second*1)

		fmt.Println("Wait for initial state")
		<-c
		fmt.Println("Initial state succeeded.")

		events.AssertSafety("floor_light_correct", func(es []events.ElevatorState) bool { return es[0].FloorLamp == es[0].Floor }, time.Millisecond*200)
		<-success
	}()

	for {
		select {
		case <-success:
			return true
		case <-timeout:
			return false
		}
	}
}
