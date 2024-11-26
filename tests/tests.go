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

// Helper function which wraps an Await. On timeout, fails the test as well.
func awaitOrFail(awaitId string, cond events.TestConditionFunction, timeout time.Duration, result chan bool) <-chan bool {
	ch_ok, ch_timeout := events.AssertUntil(awaitId, cond, timeout)

	//A timeout means the test fails.
	go func() {
		<-ch_timeout
		result <- false
	}()

	return ch_ok
}

// Waits asynchronously for the set of "threads" (functions) to
// finish running. The functions are responsible for calling an
// outside global timeout channel should they time out.
func awaitAllAsync(functions ...(func())) {
	total := len(functions)
	current := 0
	done := make(chan bool)
	for i := range functions {
		go func() {
			functions[i]()
			done <- true
		}()
	}

	//Wait for all functions above to complete.
	for {
		<-done
		current += 1
		if current >= total {
			return
		}
	}
}

func waitForInit(result chan bool) {
	fmt.Println("Wait for initial state")
	<-awaitOrFail("init", assertAll(func(e events.ElevatorState) bool {
		return e.Floor != -1 && !e.DoorOpen
	}), time.Second*10, result)

	fmt.Println("Initial state succeeded.")
}

func processCabOrder(elevator int, floor int, result chan bool) func() {
	return func() {
		<-awaitOrFail(fmt.Sprintf("process_order_%d_%d_open_door", elevator, floor), func(es []events.ElevatorState) bool {
			return es[elevator].DoorOpen && es[elevator].Floor == floor && es[elevator].Direction == elevio.MD_Stop
		}, time.Second*30, result)
		<-awaitOrFail(fmt.Sprintf("process_order_%d_%d_shut_light", elevator, floor), func(es []events.ElevatorState) bool {
			return !es[elevator].CabLights[floor]
		}, time.Second*30, result)
		<-awaitOrFail(fmt.Sprintf("process_order_%d_%d_door_not_open", elevator, floor), func(es []events.ElevatorState) bool {
			return !es[elevator].DoorOpen
		}, time.Second*30, result)
	}
}

func TestFloorLamp(result chan bool) {
	waitForInit(result)

	events.MakeOrder(0, elevio.BT_Cab, 3)
	events.AssertSafety("floor_light_correct", func(es []events.ElevatorState) bool { return es[0].FloorLamp == es[0].Floor },
		time.Millisecond*500, result)

	<-awaitOrFail("reached_dest", func(es []events.ElevatorState) bool { return es[0].Floor == 3 && es[0].FloorLamp == 3 }, time.Second*15, result)

	time.Sleep(1 * time.Second)

}

func TestInitBetweenFloors(result chan bool) {
	waitForInit(result)
	time.Sleep(1 * time.Second)
}

func TestCabBackup(result chan bool) {
	waitForInit(result)

	events.MakeOrder(0, elevio.BT_Cab, 3)
	events.MakeOrder(0, elevio.BT_Cab, 2)

	<-awaitOrFail("cab_order_confirm", func(es []events.ElevatorState) bool {
		return es[0].CabLights[3] && es[0].CabLights[2]
	}, time.Second*1, result)
	time.Sleep(500 * time.Millisecond)

	studentprogram.KillProgram(0)

	time.Sleep(3 * time.Second)
	studentprogram.StartProgram(0)

	<-awaitOrFail("cab_orders_restored", func(es []events.ElevatorState) bool {
		return es[0].CabLights[2] && es[0].CabLights[3]
	}, time.Second*10, result)

	awaitAllAsync(
		processCabOrder(0, 2, result),
		processCabOrder(0, 3, result),
	)
}
