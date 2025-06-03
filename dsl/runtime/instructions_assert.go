package runtime

import (
	"autofat/dsl/variables"
	"autofat/statemanager"
	"time"
)

type InstrAssert struct {
	AssertVal          variables.Symbol
	StateFunction      variables.Symbol
	ConditionFuncValue variables.Symbol
	DeadzoneViolated   variables.Symbol
}

func (instr *InstrAssert) Execute(runtime *thread) {
	// Wait for new state
	assert_obj := runtime.get(instr.AssertVal).(variables.AssertVal)

	var states statemanager.States = nil
	more := true

	if assert_obj.DeadzoneTimer != nil {
		select {
		case states, more = <-runtime.Listener:
			if !more {
				return
			}
		case <-assert_obj.DeadzoneTimer.C:
			runtime.set(instr.DeadzoneViolated, true)
			return
		}
	} else {
		states, more = <-runtime.Listener
		if !more {
			return
		}
	}

	call := InstrCallFunction{
		SymbolicLabel: instr.StateFunction,
		State:         states.Copy(),
		RetVal:        instr.ConditionFuncValue,
	}
	call.Execute(runtime)
}

type InstrEndAssert struct {
	Label                string           // Label to start of assert
	Deadzoneviolated     variables.Symbol // Has the deadzone been violated?
	ConditionFuncValue   variables.Symbol //Return value of state function.
	AssertVal            variables.Symbol // Data types associated with assert
	DeadzoneMilliseconds variables.Symbol //Deadzone length, in milliseconds
}

func (instr *InstrEndAssert) Execute(runtime *thread) {
	assert_val := runtime.getBool(instr.ConditionFuncValue)
	assert := runtime.get(instr.AssertVal).(variables.AssertVal)
	if runtime.getBool(instr.Deadzoneviolated) {
		// Assert has been false for too long: the program must exit with an error.
		runtime.removeListener()
		assert.DeadzoneTimer.Stop()
		runtime.exit(false)
	} else {
		//Repeat to check state again.
		jmp := InstrJmp{
			Label: instr.Label,
		}
		jmp.Execute(runtime)

		deadzone_ms := runtime.getInt(instr.DeadzoneMilliseconds)
		if !assert_val && assert.DeadzoneTimer == nil {
			if deadzone_ms > 0 {
				runtime.set(instr.AssertVal, variables.AssertVal{
					DeadzoneTimer: time.NewTimer(time.Duration(deadzone_ms) * time.Millisecond),
				})
			} else {
				runtime.removeListener()
				runtime.exit(false)
			}
		} else if assert_val && assert.DeadzoneTimer != nil {
			new_assert_val := variables.AssertVal{
				DeadzoneTimer: nil,
			}
			runtime.set(instr.AssertVal, new_assert_val)
		}
	}
}

type InstrAssertStateListen struct {
	AssertVal       variables.Symbol
	DeadzoneSeconds variables.Symbol
}

func (instr *InstrAssertStateListen) Execute(rt *thread) {
	// Initialize a new state capturing channel
	// if none exist at the symbol location.
	rt.newListener()
	if rt.get(instr.AssertVal) == nil {
		rt.set(instr.AssertVal, variables.AssertVal{
			DeadzoneTimer: nil,
		})
	}
}
