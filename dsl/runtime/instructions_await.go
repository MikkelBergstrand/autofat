package runtime

import (
	"autofat/dsl/variables"
	"autofat/statemanager"
	"fmt"
	"time"
)

type InstrAwaitStateListen struct {
	AwaitVal       variables.Symbol
	TimeoutSeconds variables.Symbol
}

func (instr *InstrAwaitStateListen) Execute(rt *RuntimeInstance) {
	// Initialize a new state capturing channel
	// if none exist at the symbol location.
	if rt.Get(instr.AwaitVal) == nil {
		timeout := time.NewTimer(time.Duration(rt.GetInt(instr.TimeoutSeconds) * int(time.Millisecond)))
		stateChan := statemanager.RegisterStateChannel()
		rt.Set(instr.AwaitVal, variables.AwaitVal{
			Timeout:   timeout,
			StateChan: stateChan,
		})
	}
}

type InstrAwait struct {
	AwaitVal           variables.Symbol
	StateFunction      variables.Symbol
	ConditionFuncValue variables.Symbol
	Timeout            variables.Symbol
}

func (instr *InstrAwait) Execute(runtime *RuntimeInstance) {
	// Wait for new state
	await_obj := runtime.Get(instr.AwaitVal).(variables.AwaitVal)
	stateChan := await_obj.StateChan
	timeout := await_obj.Timeout

	var states statemanager.States = nil
	more := true
	select {
	case states, more = <-stateChan:
		if !more {
			fmt.Println("Closing await")
			return
		}
	case <-timeout.C:
		fmt.Println("Timeout!")
		runtime.Set(instr.Timeout, true)
		return
	}

	call := InstrCallFunction{
		SymbolicLabel: instr.StateFunction,
		State:         &states,
		RetVal:        instr.ConditionFuncValue,
	}
	call.Execute(runtime)
}

type InstrEndAwait struct {
	Label              string // Label to start of await, if the await must be ran again.
	Timeout            variables.Symbol
	ConditionFuncValue variables.Symbol //Return value of state function.
	AwaitVal           variables.Symbol
}

func (instr *InstrEndAwait) Execute(runtime *RuntimeInstance) {
	await_val := runtime.GetBool(instr.ConditionFuncValue)

	fmt.Println(await_val)
	if !runtime.GetBool(instr.Timeout) && !await_val {
		jmp := InstrJmp{
			Label: instr.Label,
		}
		jmp.Execute(runtime)
	} else {
		await := runtime.Get(instr.AwaitVal).(variables.AwaitVal)
		statemanager.UnregisterStateChannel(await.StateChan)
	}
}
