package tests

import (
	"autofat/elevio"
	"autofat/events"
	"autofat/fatelevator"
	"autofat/studentprogram"
	"time"
)

func TestFloorLamp() error {
	err := waitForInit()
	if err != nil {
		return err
	}

	fatelevator.MakeOrder(0, elevio.BT_Cab, 3)
	events.Assert("floor_light_correct", func(es []events.ElevatorState) bool { return es[0].FloorLamp == es[0].Floor },
		time.Millisecond*500)

	err = events.Await("reached_dest", func(es []events.ElevatorState) bool { return es[0].Floor == 3 && es[0].FloorLamp == 3 }, time.Second*15)
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

	events.Assert("dont_move_other_elevator", func(es []events.ElevatorState) bool { return es[1].Direction == elevio.MD_Stop }, 0)

	fatelevator.MakeOrder(0, elevio.BT_Cab, 3)
	fatelevator.MakeOrder(0, elevio.BT_Cab, 2)

	err = events.Await("cab_order_confirm", func(es []events.ElevatorState) bool {
		return es[0].CabLights[3] && es[0].CabLights[2]
	}, time.Second*1)
	if err != nil {
		return err
	}
	time.Sleep(500 * time.Millisecond)

	studentprogram.KillProgram(0)

	time.Sleep(3 * time.Second)
	studentprogram.StartProgram(0)

	err = events.Await("cab_orders_restored_and_moving", func(es []events.ElevatorState) bool {
		return es[0].CabLights[2] && es[0].CabLights[3] && es[0].Direction != elevio.MD_Stop
	}, time.Second*10)

	if err != nil {
		return err
	}

	err = processOrder(0, elevio.BT_Cab, 2)()
	if err != nil {
		return err
	}

	events.Assert("dont_stop_invalid_floors", func(es []events.ElevatorState) bool {
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
