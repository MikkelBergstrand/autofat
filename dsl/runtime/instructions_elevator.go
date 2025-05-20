package runtime

import (
	"autofat/dsl/variables"
	"autofat/elevio"
	"autofat/simulator"
	"autofat/statemanager"
	"autofat/studentprogram"
	"time"
)

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
			if b {
				return 1
			} else {
				return 0
			}
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
