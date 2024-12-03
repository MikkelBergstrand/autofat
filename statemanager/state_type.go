package statemanager

import (
	"autofat/elevio"
	"autofat/studentprogram"
)

type ElevatorState struct {
	Status      studentprogram.ProgramStatus
	FloorLamp   int
	Floor       int
	Direction   elevio.MotorDirection
	DoorOpen    bool
	Obstruction bool
	Outofbounds bool

	CabLights      []bool
	HallUpLights   []bool
	HallDownLights []bool
}

func (es *ElevatorState) OrderLight(btn elevio.ButtonType, floor int) bool {
	switch btn {
	case elevio.BT_HallDown:
		return es.HallDownLights[floor]
	case elevio.BT_HallUp:
		return es.HallUpLights[floor]
	case elevio.BT_Cab:
		return es.CabLights[floor]
	default:
		panic("Invalid elevio.ButtonType")
	}
}

func InitElevatorState(nFloors int) ElevatorState {
	//Initialize a new ElevatorState object.
	ret := ElevatorState{
		FloorLamp:   -1,
		Floor:       -1,
		Direction:   elevio.MD_Stop,
		DoorOpen:    false,
		Obstruction: false,
		Status:      studentprogram.RUNNING,
		Outofbounds: false,
	}

	ret.CabLights = make([]bool, nFloors)
	ret.HallDownLights = make([]bool, nFloors)
	ret.HallUpLights = make([]bool, nFloors)

	return ret
}
