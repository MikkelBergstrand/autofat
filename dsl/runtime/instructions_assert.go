package runtime

import (
	"autofat/dsl/variables"
	"autofat/statemanager"
	"fmt"
	"time"
)

type InstrAssert struct {
	AssertVal          variables.Symbol
	StateFunction      variables.Symbol
	ConditionFuncValue variables.Symbol
	DeadzoneViolated   variables.Symbol
}

func (instr *InstrAssert) Execute(runtime *RuntimeInstance) {
	// Wait for new state
	assert_obj := runtime.Get(instr.AssertVal).(variables.AssertVal)
	stateChan := assert_obj.StateChan

	var states statemanager.States = nil
	more := true

	if assert_obj.DeadzoneTimer != nil {
		select {
		case states, more = <-stateChan:
			if !more {
				return
			}
		case <-assert_obj.DeadzoneTimer.C:
			fmt.Println("Timeout!")
			runtime.Set(instr.DeadzoneViolated, true)
			return
		}
	} else {
		states, more = <-stateChan
		if !more {
			return
		}
	}

	call := InstrCallFunction{
		SymbolicLabel: instr.StateFunction,
		State:         &states,
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

func (instr *InstrEndAssert) Execute(runtime *RuntimeInstance) {
	assert_val := runtime.GetBool(instr.ConditionFuncValue)
	fmt.Println(assert_val)
	assert := runtime.Get(instr.AssertVal).(variables.AssertVal)
	if runtime.GetBool(instr.Deadzoneviolated) {
		// Assert has been false for too long: the program must exit with an error.
		fmt.Println("Violation of assert!")
		statemanager.UnregisterStateChannel(assert.StateChan)
		assert.DeadzoneTimer.Stop()
		runtime.Exit(false)
	} else {
		//Repeat to check state again.
		jmp := InstrJmp{
			Label: instr.Label,
		}
		jmp.Execute(runtime)

		deadzone_ms := runtime.GetInt(instr.DeadzoneMilliseconds)
		if !assert_val && assert.DeadzoneTimer == nil {
			if deadzone_ms > 0 {
				fmt.Println("Starting assert deadzone timer")
				runtime.Set(instr.AssertVal, variables.AssertVal{
					StateChan:     assert.StateChan,
					DeadzoneTimer: time.NewTimer(time.Duration(deadzone_ms) * time.Millisecond),
				})
			} else {
				fmt.Println("Violation of instant assert!")
				assert.DeadzoneTimer.Stop()
				statemanager.UnregisterStateChannel(assert.StateChan)
				runtime.Exit(false)
			}
		} else if assert_val && assert.DeadzoneTimer != nil {
			fmt.Println("Ending assert deadzone timer")
			new_assert_val := variables.AssertVal{
				StateChan:     assert.StateChan,
				DeadzoneTimer: nil,
			}
			runtime.Set(instr.AssertVal, new_assert_val)
		}
	}
}



type InstrAssertStateListen struct {
	AssertVal       variables.Symbol
	DeadzoneSeconds variables.Symbol
}

func (instr *InstrAssertStateListen) Execute(rt *RuntimeInstance) {
	// Initialize a new state capturing channel
	// if none exist at the symbol location.
	if rt.Get(instr.AssertVal) == nil {
		stateChan := statemanager.RegisterStateChannel()
		rt.Set(instr.AssertVal, variables.AssertVal{
			DeadzoneTimer: nil,
			StateChan:     stateChan,
		})
	}
}

