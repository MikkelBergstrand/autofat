package tests

import (
	"autofat/elevio"
	"autofat/events"
	"autofat/studentprogram"
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

func processCabOrder(elevator int, floor int, done chan bool) {
	<-events.AssertUntil(fmt.Sprintf("process_order_%d_%d_open_door", elevator, floor), func(es []events.ElevatorState) bool {
		return es[elevator].DoorOpen && es[elevator].Floor == floor && es[elevator].Direction == elevio.MD_Stop
	}, time.Second*30)
	<-events.AssertUntil(fmt.Sprintf("process_order_%d_%d_shut_light", elevator, floor), func(es []events.ElevatorState) bool {
		return !es[elevator].CabLights[floor]
	}, time.Second*30)
	<-events.AssertUntil(fmt.Sprintf("process_order_%d_%d_door_not_open", elevator, floor), func(es []events.ElevatorState) bool {
		return !es[elevator].DoorOpen && false
	}, time.Second*30)

	done <- true
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

func TestCabBackup() {
	waitForInit()

	events.MakeOrder(0, elevio.BT_Cab, 3)
	events.MakeOrder(0, elevio.BT_Cab, 2)

	<-events.AssertUntil("cab_order_confirm", func(es []events.ElevatorState) bool {
		return es[0].CabLights[3] && es[0].CabLights[2]
	}, time.Second*1)
	time.Sleep(500 * time.Millisecond)

	studentprogram.KillProgram(0)

	time.Sleep(4 * time.Second)
	studentprogram.StartProgram(0)

	<-events.AssertUntil("cab_orders_restored", func(es []events.ElevatorState) bool {
		return es[0].CabLights[2] && es[0].CabLights[3]
	}, time.Second*10)

	done := make(chan bool)
	go processCabOrder(0, 2, done)
	go processCabOrder(0, 3, done)
	processed := 0

Wait_For_Processed:
	for {
		select {
		case <-done:
			processed += 1
			if processed == 2 {
				break Wait_For_Processed
			}
		}
	}
}
