package runtime

import (
	"autofat/dsl/variables"
	"autofat/elevio"
	"autofat/network"
	"autofat/simulator"
	"autofat/statemanager"
	"autofat/studentprogram"
	"time"
)

func bool_to_int(b bool) int {
	if b {
		return 1
	}
	return 0
}

type InstrInitializeElevators struct {
	Count        variables.Symbol
	ElevArraySym variables.Symbol
}

func (instr *InstrInitializeElevators) Execute(rt *thread) {
	n_elevators := rt.get(instr.Count).(int)
	var elev_array []any
	for i := 0; i < n_elevators; i++ {
		simulator.Init(rt.Runtime.Config.GetElevatorConfig(i), simulator.InitializationParams{InitialFloor: 0, BetweenFloors: false})
		simulator.Run(i)
		elev_array = append(elev_array, i)
	}
	rt.set(instr.ElevArraySym, elev_array)

	time.Sleep(500 * time.Millisecond)
	studentprogram.InitalizeFromConfig(
		time.Duration(rt.Runtime.Config.StudentProgramWaitTime)*time.Second,
		rt.Runtime.Config.StudentProgramDir,
		rt.Runtime.Config.GetAllElevatorConfigs(),
		n_elevators)
	time.Sleep(1000 * time.Millisecond)

	statemanager.EventListener("LOL")
}

type InstrGetFloor struct {
	ArraySymbol variables.Symbol
	Result      variables.Symbol
}

func (instr *InstrGetFloor) Execute(rt *thread) {
	collapseState(rt, instr.Result, instr.ArraySymbol,
		func(state statemanager.ElevatorState) int {
			return state.Floor
		})
}

type InstrGetFloorLight struct {
	ArraySymbol variables.Symbol
	Result      variables.Symbol
}

func (instr *InstrGetFloorLight) Execute(rt *thread) {
	collapseState(rt, instr.Result, instr.ArraySymbol,
		func(state statemanager.ElevatorState) int {
			return state.FloorLamp
		})
}

type InstrGetStatusLight struct {
	ArraySymbol variables.Symbol
	OrderType   variables.Symbol
	Floor       variables.Symbol
	Result      variables.Symbol
}

func (instr *InstrGetStatusLight) Execute(rt *thread) {
	floor := rt.getInt(instr.Floor)
	ordertype := rt.get(instr.OrderType).(elevio.ButtonType)

	collapseState(rt, instr.Result, instr.ArraySymbol,
		func(state statemanager.ElevatorState) int {
			b := state.OrderLight(ordertype, floor)
			return bool_to_int(b)
		})
}

type InstrGetDoorStatus struct {
	ArraySymbol variables.Symbol
	Result      variables.Symbol
}

func (instr *InstrGetDoorStatus) Execute(rt *thread) {
	collapseState(rt, instr.Result, instr.ArraySymbol,
		func(state statemanager.ElevatorState) int {
			return bool_to_int(state.DoorOpen)
		})
}

type InstrGetMovementStatus struct {
	ArraySymbol variables.Symbol
	Result      variables.Symbol
}

func (instr *InstrGetMovementStatus) Execute(rt *thread) {
	collapseState(rt, instr.Result, instr.ArraySymbol,
		func(state statemanager.ElevatorState) int {
			return bool_to_int(state.Direction != elevio.MD_Stop)
		})
}

// Takes in a function that outputs an integer as a product of a single elevator's state.
// This integer is then compared across all inputted elevators.
// If it is non-equal for some elevators, it sets result to -1
// If it is equal for all elevators, it sets result to that value.
func collapseState(rt *thread, result variables.Symbol, elevatorListSym variables.Symbol, stateFunc func(state statemanager.ElevatorState) int) {
	arr := rt.get(elevatorListSym).([]any)
	state := *rt.getState()

	if len(arr) == 0 {
		rt.set(result, -1)
		return
	}

	val := stateFunc(state[arr[0].(int)])
	for i := 1; i < len(arr); i++ {
		if stateFunc(state[arr[i].(int)]) != val {
			rt.set(result, -1)
			return
		}
	}
	rt.set(result, val)
}

type InstrMakeOrder struct {
	Elevator  variables.Symbol
	Floor     variables.Symbol
	OrderType variables.Symbol
}

func (instr *InstrMakeOrder) Execute(rt *thread) {
	floor := rt.getInt(instr.Floor)
	elev := rt.getInt(instr.Elevator)
	order_type := rt.get(instr.OrderType).(elevio.ButtonType)

	simulator.MakeOrder(elev, order_type, floor)
}

type InstrKillApplication struct {
	Elevator variables.Symbol
}

func (instr *InstrKillApplication) Execute(rt *thread) {
	elev := rt.getInt(instr.Elevator)
	studentprogram.KillProgram(elev)
}

type InstrRebootApplication struct {
	Elevator variables.Symbol
}

func (instr *InstrRebootApplication) Execute(rt *thread) {
	elev := rt.getInt(instr.Elevator)
	studentprogram.StartProgram(elev)
}

type InstrSetPacketLoss struct {
	Elevators  variables.Symbol
	Percentage variables.Symbol
}

func (instr *InstrSetPacketLoss) Execute(rt *thread) {
	elevs := rt.get(instr.Elevators).([]any)
	//convert []any to []int
	var elevs_int []int
	for _, i := range elevs {
		elevs_int = append(elevs_int, i.(int))
	}
	network.SetGlobalPacketLoss(elevs_int, rt.getInt(instr.Percentage))
}
