package runtime

import (
	"autofat/dsl/color"
	"autofat/dsl/variables"
	"autofat/statemanager"
	"fmt"
	"slices"
	"time"
)

const (
	ADD Operator = iota
	SUB
	MULT
	DIV
	MOD
)

type InstructionLabelPair struct {
	Instruction Instruction
	Label       string
}
type Instruction interface {
	Execute(*RuntimeInstance)
}

type Operator int
type BooleanOperator int

const (
	EQUALS BooleanOperator = iota
	NOTEQUALS
	LESS
	LESSOREQUAL
	GREATER
	GREATEROREQUAL
	AND
	OR
	NOT
)

func (op BooleanOperator) IsValidFor(t variables.Type) bool {
	legalBools := []BooleanOperator{EQUALS, NOTEQUALS, AND, OR, NOT}
	legalInts := []BooleanOperator{EQUALS, NOTEQUALS, LESS, LESSOREQUAL, GREATER, GREATEROREQUAL}

	switch t {
	case variables.BOOL:
		return slices.Contains(legalBools, op)
	case variables.INT:
		return slices.Contains(legalInts, op)
	}

	return false
}

type InstrArithmetic struct {
	A        variables.Symbol
	B        variables.Symbol
	Operator Operator
	Result   variables.Symbol
}

type InstrAssign struct {
	Dest   variables.Symbol
	Source variables.Symbol
}

func (instr *InstrArithmetic) Execute(runtime *RuntimeInstance) {
	switch instr.Operator {
	case ADD:
		runtime.Set(instr.Result, runtime.GetInt(instr.A)+runtime.GetInt(instr.B))
	case MULT:
		runtime.Set(instr.Result, runtime.GetInt(instr.A)*runtime.GetInt(instr.B))
	case DIV:
		runtime.Set(instr.Result, runtime.GetInt(instr.A)/runtime.GetInt(instr.B))
	case SUB:
		runtime.Set(instr.Result, runtime.GetInt(instr.A)-runtime.GetInt(instr.B))
	case MOD:
		runtime.Set(instr.Result, runtime.GetInt(instr.A)%runtime.GetInt(instr.B))
	}
}

type InstrCompareInt struct {
	A        variables.Symbol
	B        variables.Symbol
	Operator BooleanOperator
	Result   variables.Symbol
}

func (instr *InstrCompareInt) Execute(runtime *RuntimeInstance) {
	switch instr.Operator {
	case EQUALS:
		runtime.Set(instr.Result, runtime.GetInt(instr.A) == runtime.GetInt(instr.B))
	case NOTEQUALS:
		runtime.Set(instr.Result, runtime.GetInt(instr.A) != runtime.GetInt(instr.B))
	case LESS:
		runtime.Set(instr.Result, runtime.GetInt(instr.A) < runtime.GetInt(instr.B))
	case LESSOREQUAL:
		runtime.Set(instr.Result, runtime.GetInt(instr.A) <= runtime.GetInt(instr.B))
	case GREATER:
		runtime.Set(instr.Result, runtime.GetInt(instr.A) > runtime.GetInt(instr.B))
	case GREATEROREQUAL:
		//fmt.Println("Comparing")
		runtime.Set(instr.Result, runtime.GetInt(instr.A) >= runtime.GetInt(instr.B))
	}
}

type InstrCompareBool struct {
	A        variables.Symbol
	B        variables.Symbol
	Operator BooleanOperator
	Result   variables.Symbol
}

func (instr *InstrCompareBool) Execute(runtime *RuntimeInstance) {
	switch instr.Operator {
	case EQUALS:
		runtime.Set(instr.Result, runtime.GetBool(instr.A) == runtime.GetBool(instr.B))
	case NOTEQUALS:
		runtime.Set(instr.Result, runtime.GetBool(instr.A) != runtime.GetBool(instr.B))
	case AND:
		runtime.Set(instr.Result, runtime.GetBool(instr.A) && runtime.GetBool(instr.B))
	case OR:
		runtime.Set(instr.Result, runtime.GetBool(instr.A) || runtime.GetBool(instr.B))
	}
}

type InstrJmp struct {
	Label string
}

func (instr *InstrJmp) Execute(runtime *RuntimeInstance) {
	runtime.Programcounter = runtime.Runtime.GetLabel(instr.Label) - 1 // decrement, since it is autoincremented
}

type InstrJmpVar struct {
	Label string
}

func (instr *InstrJmpVar) Execute(runtime *RuntimeInstance) {
	runtime.Programcounter = runtime.Runtime.GetLabel(instr.Label) - 1
}

type InstrJmpIf struct {
	Label     string
	Condition variables.Symbol
}

func (instr *InstrJmpIf) Execute(runtime *RuntimeInstance) {
	if !runtime.GetBool(instr.Condition) {
		runtime.Programcounter = runtime.Runtime.GetLabel(instr.Label) - 1
	}
}

type InstrLoadImmediate struct {
	Dest  variables.Symbol
	Value any
}

func (instr *InstrLoadImmediate) Execute(runtime *RuntimeInstance) {
	runtime.Set(instr.Dest, instr.Value)
}

type InstrLoadFunction struct {
	Symbol variables.Symbol
	Label  string
}

func (instr *InstrLoadFunction) Execute(runtime *RuntimeInstance) {
	// Copy the current address stack.
	runtime.Set(instr.Symbol, FunctionVar{
		Label:        instr.Label,
		AddressStack: runtime.CallStack.PeekRef().AddressStack.Copy(),
	})
}

func (instr *InstrAssign) Execute(runtime *RuntimeInstance) {
	runtime.Set(instr.Dest, runtime.Get(instr.Source))
}

type InstructionEcho struct {
	A variables.Symbol
}

func (instr *InstructionEcho) Execute(runtime *RuntimeInstance) {
	rt, addr := runtime.AddressFromSymbol(instr.A)
	color.Println(color.Green, rt.Variables[addr])
}

type InstrCallFunction struct {
	Arguments     []variables.Symbol
	RetVal        variables.Symbol
	SymbolicLabel variables.Symbol
	State         *statemanager.States
	Fork          bool
}

func (instr *InstrCallFunction) Execute(runtime *RuntimeInstance) {
	//Fetch and copy argument values
	var arg_values []any
	for i := range instr.Arguments {
		arg_values = append(arg_values, runtime.Get(instr.Arguments[i]))
	}
	//Fetch func_ptr
	func_ptr := runtime.Get(instr.SymbolicLabel).(FunctionVar)
	conv_addr_stack := func_ptr.AddressStack.Copy()

	if !instr.Fork {
		// Bind return value
		top_ar := runtime.CallStack.PeekRef()
		top_ar.Retval = instr.RetVal
		//fmt.Println("Bound ret val to", top_ar.Retval)

		runtime.PushCall(conv_addr_stack, instr.State)

		// Once "inside" the function, load argument values
		for i := range arg_values {
			runtime.Set(variables.Symbol{Offset: i, Scope: 0, Type: instr.Arguments[i].Type}, arg_values[i])
		}

		//Then, jump to the function's label
		jmp_instr := InstrJmpVar{
			Label: func_ptr.Label,
		}
		jmp_instr.Execute(runtime)
	} else {
		//Get thread object
		done := make(chan bool)
		runtime.Set(instr.RetVal, variables.Thread{
			Done: done,
		})

		// Create a new runtime
		fmt.Println(conv_addr_stack)
		new_runtime := runtime.Fork(runtime.Runtime.Labels[func_ptr.Label], conv_addr_stack)

		new_runtime.PushAddress()
		// Set arguments in new runtime
		for i := range arg_values {
			new_runtime.Set(variables.Symbol{Offset: i, Scope: 0, Type: instr.Arguments[i].Type}, arg_values[i])
		}
		go new_runtime.Run(done)
		fmt.Println("Thread forked!")
	}

}

type InstrBeginScope struct{}

func (instr *InstrBeginScope) Execute(runtime *RuntimeInstance) {
	runtime.PushAddress()
}

type InstrEndScope struct{}

func (instr *InstrEndScope) Execute(runtime *RuntimeInstance) {
	runtime.PopAddress()
}

type InstrExitFunction struct {
	RetVal variables.Symbol
}

func (instr *InstrExitFunction) Execute(runtime *RuntimeInstance) {
	ret_val := runtime.Get(instr.RetVal)

	runtime.PopCall()

	//If callstack is empty, this thread is done.
	if len(runtime.CallStack) == 0 {
		runtime.Programcounter = RT_EXIT
		fmt.Println("Exiting thread.")
	} else {
		top_ar := runtime.CallStack.PeekRef()
		//fmt.Println("Ret val on exit", top_ar.Retval, ret_val)
		runtime.Set(top_ar.Retval, ret_val)
	}

}

// Does nothing.
type InstrNOP struct{}

func (instr *InstrNOP) Execute(runtime *RuntimeInstance) {}

type InstrLoadArray struct {
	SrcSymbols []variables.Symbol
	DestSymbol variables.Symbol
}

func (instr *InstrLoadArray) Execute(rt *RuntimeInstance) {
	var arr []any
	for _, sym := range instr.SrcSymbols {
		arr = append(arr, rt.Get(sym))
	}
	rt.Set(instr.DestSymbol, arr)
}

type InstrExit struct {
	Value variables.Symbol
}

func (instr *InstrExit) Execute(rt *RuntimeInstance) {
	rt.Exit(rt.GetBool(instr.Value))
}

type InstrSleep struct {
	Duration variables.Symbol
}

func (instr *InstrSleep) Execute(rt *RuntimeInstance) {
	msec := rt.GetInt(instr.Duration)
	time.Sleep(time.Duration(msec) * time.Millisecond)
}

type InstrSync struct {
	Threads variables.Symbol
}

func (instr *InstrSync) Execute(rt *RuntimeInstance) {
	threads := rt.Get(instr.Threads).([]any)
	n_done_threads := 0
	n_threads := len(threads)
	sig := make(chan bool)

	for _, thread := range threads {
		thread := thread.(variables.Thread)
		go func() {
			<-thread.Done
			sig <- true
		}()
	}

	for {
		<-sig
		n_done_threads += 1
		if n_done_threads == n_threads {
			break
		}
	}

}
