package tests

import (
	"autofat/elevio"
	"autofat/simulator"
	"autofat/statemanager"
	"autofat/studentprogram"
	"fmt"
	"time"
)

// Wait for all elevators to reach a valid floor with their doors closed.
func waitForInit() error {
	//Add default assertions
	statemanager.Assert("program_crash", assertAll(func(e statemanager.ElevatorState) bool { return e.Status != studentprogram.CRASHED }), 0)
	statemanager.Assert("program_outofbound", assertAll(func(e statemanager.ElevatorState) bool { return !e.Outofbounds }), 0)

	fmt.Println("Wait for initial state")
	err := statemanager.Await("init", assertAll(func(e statemanager.ElevatorState) bool {
		return e.Floor != -1 && !e.DoorOpen
	}), time.Second*10)

	return err
}

// Make a hall order.
// This will press the hall order buttons 3 times, waiting for the order process lights to shine.
// If only a subset of lights light up, the test will fail.
func makeHallOrder(elevator int, btn elevio.ButtonType, floor int, elevatorsOnline []int) error {
	if btn != elevio.BT_HallDown && btn != elevio.BT_HallUp {
		panic("Wrong type of ButtonEvent in makeHallOrder()!")
	}

	prefix := fmt.Sprintf("%d_%d_%d_", elevator, floor, btn)

	//Wait for all the hall orders to light up. At the same time, we do not allow only
	//a subset of them lighting up.
	statemanager.Assert(prefix+"all_hall_orders_or_none", func(es []statemanager.ElevatorState) bool {
		n_lights := 0
		for i := range elevatorsOnline {
			if (btn == elevio.BT_HallUp && es[i].HallUpLights[floor]) ||
				(btn == elevio.BT_HallDown && es[i].HallDownLights[floor]) {
				n_lights += 1
			}
		}
		return n_lights == 0 || n_lights == len(elevatorsOnline)
	}, 1*time.Second)

	var err error = nil

	//Now, wait for all hall orders to be set.
	//We repeat making the button press three times before giving up
	for repeat := 0; repeat < 3; repeat++ {
		simulator.MakeOrder(elevator, btn, floor)

		err = statemanager.Await(prefix+"confirmed_all_hall_orders", func(es []statemanager.ElevatorState) bool {
			n_lights := 0
			for i := range elevatorsOnline {
				if (btn == elevio.BT_HallUp && es[i].HallUpLights[floor]) ||
					(btn == elevio.BT_HallDown && es[i].HallDownLights[floor]) {
					n_lights += 1
				}
			}
			return n_lights == len(elevatorsOnline)
		}, 10*time.Second)
		if err == nil {
			break
		}
	}

	statemanager.Disassert(prefix + "all_hall_orders_or_none")
	//Once we have lit the hall orders, we assume we are OK.
	return err
}

// Wait for one elevator to move.
// The test will fail if 0 or more than one elevator moves.
// The ID of the moving elevator will be returned on success.
func awaitOneMovingElevator(elevatorsOnline []int) (int, error) {
	statemanager.Assert("only_one_moving_elevator", func(es []statemanager.ElevatorState) bool {
		elevs_moving := 0
		for _, id := range elevatorsOnline {
			if es[id].Direction != elevio.MD_Stop {
				elevs_moving += 1
			}
		}
		return elevs_moving < 2
	}, 0)

	ret_val := 0
	err := statemanager.Await("await_moving_elevator", func(es []statemanager.ElevatorState) bool {
		for _, id := range elevatorsOnline {
			if es[id].Direction != elevio.MD_Stop {
				ret_val = id
				return true
			}
		}
		return false
	}, 5*time.Second)

	return ret_val, err
}

// Wait for an order to be processed. By processed we mean:
// 1) elevator stops at said floor and opens the doors
// 2) the light is shut.
// 3) The doors close again.
func processOrder(elevator int, btn elevio.ButtonType, floor int) func() error {
	return func() error {
		err := statemanager.Await(fmt.Sprintf("process_order_%d_%d_open_door", elevator, floor), func(es []statemanager.ElevatorState) bool {
			return es[elevator].DoorOpen && es[elevator].Floor == floor && es[elevator].Direction == elevio.MD_Stop
		}, time.Second*30)
		if err != nil {
			return err
		}
		err = statemanager.Await(fmt.Sprintf("process_order_%d_%d_shut_light", elevator, floor), func(es []statemanager.ElevatorState) bool {
			return !es[elevator].OrderLight(btn, floor)
		}, time.Second*30)
		if err != nil {
			return err
		}

		err = statemanager.Await(fmt.Sprintf("process_order_%d_%d_door_not_open", elevator, floor), func(es []statemanager.ElevatorState) bool {
			return !es[elevator].DoorOpen
		}, time.Second*30)
		if err != nil {
			return err
		}

		return nil
	}
}

// Helper function. We want to evaluate a condition that should be true for all
// elevators in the system.
func assertAll(test func(e statemanager.ElevatorState) bool) func([]statemanager.ElevatorState) bool {
	return func(es []statemanager.ElevatorState) bool {
		for i := range es {
			if !test(es[i]) {
				return false
			}
		}
		return true
	}
}

// Waits asynchronously for the set of "threads" (functions) to
// finish running. The functions are responsible for calling an
// outside global timeout channel should they time out.
func awaitAllAsync(functions ...(func() error)) error {
	total := len(functions)
	current := 0
	done := make(chan error)
	for i := range functions {
		go func() {
			done <- functions[i]()
		}()
	}

	//Wait for all functions above to complete.
	for {
		err := <-done
		if err != nil {
			return err
		}
		current += 1
		if current >= total {
			return nil
		}
	}
}
