package tests

import (
	"autofat/elevio"
	"autofat/fatelevator"
	"autofat/statemanager"
	"autofat/studentprogram"
	"time"
)

func TestFloorLamp() error {
	err := waitForInit()
	if err != nil {
		return err
	}

	fatelevator.MakeOrder(0, elevio.BT_Cab, 3)
	statemanager.Assert("floor_light_correct", func(es []statemanager.ElevatorState) bool { return es[0].FloorLamp == es[0].Floor },
		time.Millisecond*500)

	err = statemanager.Await("reached_dest", func(es []statemanager.ElevatorState) bool { return es[0].Floor == 3 && es[0].FloorLamp == 3 }, time.Second*15)
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)
	return nil
}

func TestInitBetweenFloors() error {
	err := waitForInit()
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)
	return nil
}

func TestCabBackup() error {
	err := waitForInit()
	if err != nil {
		return err
	}

	statemanager.Assert("dont_move_other_elevator", func(es []statemanager.ElevatorState) bool { return es[1].Direction == elevio.MD_Stop }, 0)

	fatelevator.MakeOrder(0, elevio.BT_Cab, 3)
	fatelevator.MakeOrder(0, elevio.BT_Cab, 2)

	err = statemanager.Await("cab_order_confirm", func(es []statemanager.ElevatorState) bool {
		return es[0].CabLights[3] && es[0].CabLights[2]
	}, time.Second*1)
	if err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)

	studentprogram.KillProgram(0)

	time.Sleep(3 * time.Second)
	studentprogram.StartProgram(0)

	err = statemanager.Await("cab_orders_restored_and_moving", func(es []statemanager.ElevatorState) bool {
		return es[0].CabLights[2] && es[0].CabLights[3] && es[0].Direction != elevio.MD_Stop
	}, time.Second*10)

	if err != nil {
		return err
	}

	err = processOrder(0, elevio.BT_Cab, 2)()
	if err != nil {
		return err
	}

	statemanager.Assert("dont_stop_invalid_floors", func(es []statemanager.ElevatorState) bool {
		return !((es[0].Floor == 0 || es[0].Floor == 1) && es[0].Direction == elevio.MD_Stop)
	}, 0)

	err = processOrder(0, elevio.BT_Cab, 3)()
	if err != nil {
		return err
	}

	return nil
}

func TestEngineOutage() error {
	err := waitForInit()
	if err != nil {
		return err
	}

	err = makeHallOrder(0, elevio.BT_HallDown, 3, []int{0, 1})
	if err != nil {
		return err
	}
	time.Sleep(1 * time.Second)

	moving, err := awaitOneMovingElevator([]int{0, 1})
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)
	fatelevator.SetEngineFailureState(moving, true)

	err = processOrder(1, elevio.BT_HallDown, 3)()
	if err != nil {
		return err
	}

	return nil
}

// Test that the door is opening for 2-4 seconds, then closing, on floor arrival
func TestDoorOpenTime() error {
	err := waitForInit()
	if err != nil {
		return err
	}

	fatelevator.MakeOrder(0, elevio.BT_Cab, 2)
	err = statemanager.Await("floor_arrival", func(es []statemanager.ElevatorState) bool { return es[0].Floor == 2 }, 15*time.Second)
	if err != nil {
		return err
	}

	err = statemanager.Await("door_open", func(es []statemanager.ElevatorState) bool { return es[0].DoorOpen }, 1*time.Second)
	if err != nil {
		return err
	}

	statemanager.Assert("keep_door_open", func(es []statemanager.ElevatorState) bool { return es[0].DoorOpen }, 0)
	time.Sleep(1 * time.Second) //Not the full three seconds, so we have some leeway
	statemanager.Disassert("keep_door_open")

	err = statemanager.Await("door_close", func(es []statemanager.ElevatorState) bool { return !es[0].DoorOpen }, 2*time.Second)
	if err != nil {
		return err
	}

	return nil
}

// Test that when multiple hall lights are asserted at a floor, only one
// is cleared at a time.
func TestHallClearOne() error {
	err := waitForInit()
	if err != nil {
		return err
	}

	statemanager.Assert("dont_reach_dest_prematurely", func(es []statemanager.ElevatorState) bool { return es[0].Floor != 2 }, 0)
	err = makeHallOrder(0, elevio.BT_HallDown, 2, []int{0})
	if err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)
	err = makeHallOrder(0, elevio.BT_HallUp, 2, []int{0})
	if err != nil {
		return err
	}
	statemanager.Disassert("dont_reach_dest_prematurely")

	//Direction UP should be cleared first, as this is the reference direction.
	statemanager.Assert("dont_clear_halldown_first", func(es []statemanager.ElevatorState) bool { return es[0].HallDownLights[2] }, 500*time.Millisecond)
	err = statemanager.Await("clear_hallup", func(es []statemanager.ElevatorState) bool { return !es[0].HallUpLights[2] }, 15*time.Second)
	if err != nil {
		return err
	}
	//Now, halldown should still not be cleared instantly. We wait at least 2 seconds (3 seconds is the time it _should_ take)
	time.Sleep(2 * time.Second)
	statemanager.Disassert("dont_clear_halldown_first")
	err = statemanager.Await("clear_halldown", func(es []statemanager.ElevatorState) bool { return !es[0].HallDownLights[2] }, 5*time.Second)
	if err != nil {
		return err
	}

	return nil
}

// Obstruction should open the door if the elevator is stationary at a floor
func TestObstructionOpenDoor() error {
	err := waitForInit()
	if err != nil {
		return err
	}

	time.Sleep(1 * time.Second)
	fatelevator.SetObstruction(0, true)

	err = statemanager.Await("door_opening", func(es []statemanager.ElevatorState) bool { return es[0].DoorOpen }, 10*time.Second)
	if err != nil {
		return err
	}

	return nil
}

// Orders made during obstruction should be "buffered" until no longer obstructed anymore
func TestObstructionCompleteOrders() error {
	err := waitForInit()
	if err != nil {
		return err
	}

	statemanager.Assert("remain_stationary", func(es []statemanager.ElevatorState) bool { return es[0].DoorOpen && es[0].Direction == elevio.MD_Stop },
		250*time.Millisecond)

	time.Sleep(1 * time.Second)
	fatelevator.SetObstruction(0, true)

	time.Sleep(1 * time.Second)
	fatelevator.MakeOrder(0, elevio.BT_Cab, 2)
	fatelevator.MakeOrder(0, elevio.BT_HallDown, 3)
	time.Sleep(1 * time.Second)

	statemanager.Disassert("remain_stationary")
	fatelevator.SetObstruction(0, false)

	err = processOrder(0, elevio.BT_Cab, 2)()
	if err != nil {
		return err
	}
	err = processOrder(0, elevio.BT_HallDown, 3)()
	if err != nil {
		return err
	}
	return nil
}
