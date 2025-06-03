package runtime

import (
	"autofat/dsl/variables"
	"autofat/statemanager"
	"time"
)

type InstrAwaitStateListen struct {
	AwaitVal       variables.Symbol
	TimeoutSeconds variables.Symbol
}

func (instr *InstrAwaitStateListen) Execute(rt *thread) {
	// Initialize a new state capturing channel
	// if none exist at the symbol location.
	if rt.get(instr.AwaitVal) == nil {
		timeout := time.NewTimer(time.Duration(rt.getInt(instr.TimeoutSeconds) * int(time.Millisecond)))
		rt.newListener()
		rt.set(instr.AwaitVal, variables.AwaitVal{
			Timeout: timeout,
		})
	}
}

type InstrAwait struct {
	AwaitVal           variables.Symbol
	StateFunction      variables.Symbol
	ConditionFuncValue variables.Symbol
	Timeout            variables.Symbol
}

func (instr *InstrAwait) Execute(runtime *thread) {
	// Wait for new state
	await_obj := runtime.get(instr.AwaitVal).(variables.AwaitVal)
	timeout := await_obj.Timeout

	var states statemanager.States = nil
	more := true
	select {
	case states, more = <-runtime.Listener:
		if !more {
			return
		}
	case <-timeout.C:
		runtime.set(instr.Timeout, true)
		return
	}

	call := InstrCallFunction{
		SymbolicLabel: instr.StateFunction,
		State:         states.Copy(),
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

func (instr *InstrEndAwait) Execute(runtime *thread) {
	await_val := runtime.getBool(instr.ConditionFuncValue)
	if !runtime.getBool(instr.Timeout) && !await_val {
		jmp := InstrJmp{
			Label: instr.Label,
		}
		jmp.Execute(runtime)
	} else {
		runtime.removeListener()
	}
}
