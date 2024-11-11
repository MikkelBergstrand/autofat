package tests

import (
	"autofat/events"
	"fmt"
	"time"
)

func TestFloorLamp() {
	fmt.Println("Wait for initial state")
	<-events.AssertUntil("init", func(es []events.ElevatorState) bool { return es[0].Floor != -1 && !es[0].DoorOpen }, time.Second*1)
	fmt.Println("Initial state succeeded.")
	events.AssertSafety("floor_light_correct", func(es []events.ElevatorState) bool { return es[0].FloorLamp == es[0].Floor }, time.Millisecond*200)
}
