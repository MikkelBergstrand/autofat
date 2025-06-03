package runtime

import (
	"autofat/dsl/color"
	"autofat/dsl/variables"
	"autofat/statemanager"
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
	Execute(*thread)
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

func (instr *InstrArithmetic) Execute(runtime *thread) {
	switch instr.Operator {
	case ADD:
		runtime.set(instr.Result, runtime.getInt(instr.A)+runtime.getInt(instr.B))
	case MULT:
		runtime.set(instr.Result, runtime.getInt(instr.A)*runtime.getInt(instr.B))
	case DIV:
		runtime.set(instr.Result, runtime.getInt(instr.A)/runtime.getInt(instr.B))
	case SUB:
		runtime.set(instr.Result, runtime.getInt(instr.A)-runtime.getInt(instr.B))
	case MOD:
		runtime.set(instr.Result, runtime.getInt(instr.A)%runtime.getInt(instr.B))
	}
}

type InstrCompareInt struct {
	A        variables.Symbol
	B        variables.Symbol
	Operator BooleanOperator
	Result   variables.Symbol
}

func (instr *InstrCompareInt) Execute(runtime *thread) {
	switch instr.Operator {
	case EQUALS:
		runtime.set(instr.Result, runtime.getInt(instr.A) == runtime.getInt(instr.B))
	case NOTEQUALS:
		runtime.set(instr.Result, runtime.getInt(instr.A) != runtime.getInt(instr.B))
	case LESS:
		runtime.set(instr.Result, runtime.getInt(instr.A) < runtime.getInt(instr.B))
	case LESSOREQUAL:
		runtime.set(instr.Result, runtime.getInt(instr.A) <= runtime.getInt(instr.B))
	case GREATER:
		runtime.set(instr.Result, runtime.getInt(instr.A) > runtime.getInt(instr.B))
	case GREATEROREQUAL:
		//fmt.Println("Comparing")
		runtime.set(instr.Result, runtime.getInt(instr.A) >= runtime.getInt(instr.B))
	}
}

type InstrCompareBool struct {
	A        variables.Symbol
	B        variables.Symbol
	Operator BooleanOperator
	Result   variables.Symbol
}

func (instr *InstrCompareBool) Execute(runtime *thread) {
	switch instr.Operator {
	case EQUALS:
		runtime.set(instr.Result, runtime.getBool(instr.A) == runtime.getBool(instr.B))
	case NOTEQUALS:
		runtime.set(instr.Result, runtime.getBool(instr.A) != runtime.getBool(instr.B))
	case AND:
		runtime.set(instr.Result, runtime.getBool(instr.A) && runtime.getBool(instr.B))
	case OR:
		runtime.set(instr.Result, runtime.getBool(instr.A) || runtime.getBool(instr.B))
	}
}

type InstrJmp struct {
	Label string
}

func (instr *InstrJmp) Execute(runtime *thread) {
	runtime.Programcounter = runtime.Runtime.getLabel(instr.Label) - 1 // decrement, since it is autoincremented
}

type InstrJmpVar struct {
	Label string
}

func (instr *InstrJmpVar) Execute(runtime *thread) {
	runtime.Programcounter = runtime.Runtime.getLabel(instr.Label) - 1
}

type InstrJmpIf struct {
	Label     string
	Condition variables.Symbol
}

func (instr *InstrJmpIf) Execute(runtime *thread) {
	if !runtime.getBool(instr.Condition) {
		runtime.Programcounter = runtime.Runtime.getLabel(instr.Label) - 1
	}
}

type InstrLoadImmediate struct {
	Dest  variables.Symbol
	Value any
}

func (instr *InstrLoadImmediate) Execute(runtime *thread) {
	runtime.set(instr.Dest, instr.Value)
}

type InstrLoadFunction struct {
	Symbol variables.Symbol
	Label  string
}

func (instr *InstrLoadFunction) Execute(runtime *thread) {
	// Copy the current address stack.
	runtime.set(instr.Symbol, FunctionVar{
		Label:        instr.Label,
		AddressStack: runtime.CallStack.PeekRef().AddressStack.Copy(),
	})
}

func (instr *InstrAssign) Execute(runtime *thread) {
	runtime.set(instr.Dest, runtime.get(instr.Source))
}

type InstructionEcho struct {
	A variables.Symbol
}

func (instr *InstructionEcho) Execute(runtime *thread) {
	rt, addr := runtime.addressFromSymbol(instr.A)
	color.Println(color.Green, rt.Variables[addr])
}

type InstrCallFunction struct {
	Arguments     []variables.Symbol
	RetVal        variables.Symbol
	SymbolicLabel variables.Symbol
	State         *statemanager.States
	Fork          bool
}

func (instr *InstrCallFunction) Execute(runtime *thread) {
	//Fetch and copy argument values
	var arg_values []any
	for i := range instr.Arguments {
		arg_values = append(arg_values, runtime.get(instr.Arguments[i]))
	}
	//Fetch func_ptr
	func_ptr := runtime.get(instr.SymbolicLabel).(FunctionVar)
	conv_addr_stack := func_ptr.AddressStack.Copy()

	if !instr.Fork {
		// Bind return value
		top_ar := runtime.CallStack.PeekRef()
		top_ar.Retval = instr.RetVal
		//fmt.Println("Bound ret val to", top_ar.Retval)

		runtime.pushCall(conv_addr_stack, instr.State)

		// Once "inside" the function, load argument values
		for i := range arg_values {
			runtime.set(variables.Symbol{Offset: i, Scope: 0, Type: instr.Arguments[i].Type}, arg_values[i])
		}

		//Then, jump to the function's label
		jmp_instr := InstrJmpVar{
			Label: func_ptr.Label,
		}
		jmp_instr.Execute(runtime)
	} else {
		//Get thread object
		done := make(chan bool)
		kill := make(chan bool)
		runtime.set(instr.RetVal, variables.Thread{
			Done: done,
			Kill: kill,
		})

		// Create a new runtime
		new_runtime := runtime.fork(runtime.Runtime.Labels[func_ptr.Label], conv_addr_stack)

		// Set arguments in new runtime
		for i := range arg_values {
			new_runtime.set(variables.Symbol{Offset: i, Scope: 0, Type: instr.Arguments[i].Type}, arg_values[i])
		}
		go new_runtime.Run(kill, done)
	}

}

type InstrBeginScope struct{}

func (instr *InstrBeginScope) Execute(runtime *thread) {
	runtime.pushAddress()
}

type InstrEndScope struct{}

func (instr *InstrEndScope) Execute(runtime *thread) {
	runtime.popAddress()
}

type InstrExitFunction struct {
	RetVal variables.Symbol
}

func (instr *InstrExitFunction) Execute(runtime *thread) {
	ret_val := runtime.get(instr.RetVal)

	runtime.popCall()

	//If callstack is empty, this thread is done.
	if len(runtime.CallStack) == 0 {
		runtime.Programcounter = RT_EXIT
	} else {
		top_ar := runtime.CallStack.PeekRef()
		//fmt.Println("Ret val on exit", top_ar.Retval, ret_val)
		runtime.set(top_ar.Retval, ret_val)
	}

}

// Does nothing.
type InstrNOP struct{}

func (instr *InstrNOP) Execute(runtime *thread) {}

type InstrLoadArray struct {
	SrcSymbols []variables.Symbol
	DestSymbol variables.Symbol
}

func (instr *InstrLoadArray) Execute(rt *thread) {
	var arr []any
	for _, sym := range instr.SrcSymbols {
		arr = append(arr, rt.get(sym))
	}
	rt.set(instr.DestSymbol, arr)
}

type InstrArrayLookup struct {
	Array  variables.Symbol
	Index  variables.Symbol
	Result variables.Symbol
}

type InstrArrayAppend struct {
	Array variables.Symbol
	Value variables.Symbol
}

func (instr *InstrArrayAppend) Execute(rt *thread) {
	array := rt.get(instr.Array).([]any)
	value := rt.get(instr.Value)

	rt.set(instr.Array, append(array, value))
}

type InstrArrayDiff struct {
	ArrayA variables.Symbol
	ArrayB variables.Symbol
	Result variables.Symbol
}

func (instr *InstrArrayDiff) Execute(rt *thread) {
	arrayA := rt.get(instr.ArrayA).([]any)
	arrayB := rt.get(instr.ArrayB).([]any)

	var result []any
	for i := range arrayA {
		overlap := false
		for j := range arrayB {
			if arrayA[i] == arrayB[j] {
				overlap = true
				break
			}
		}
		if !overlap {
			result = append(result, arrayA[i])
		}
	}
	rt.set(instr.Result, result)
}

func (instr *InstrArrayLookup) Execute(rt *thread) {
	index := rt.getInt(instr.Index)
	array := rt.get(instr.Array).([]any)

	rt.set(instr.Result, array[index])
}

type InstrArrayLen struct {
	A      variables.Symbol
	Result variables.Symbol
}

func (instr *InstrArrayLen) Execute(rt *thread) {
	arr := rt.get(instr.A).([]any)
	rt.set(instr.Result, len(arr))
}

type InstrExit struct {
	Value variables.Symbol
}

func (instr *InstrExit) Execute(rt *thread) {
	rt.exit(rt.getBool(instr.Value))
}

type InstrSleep struct {
	Duration variables.Symbol
}

func (instr *InstrSleep) Execute(rt *thread) {
	msec := rt.getInt(instr.Duration)
	time.Sleep(time.Duration(msec) * time.Millisecond)
}

type InstrSync struct {
	Threads variables.Symbol
}

func (instr *InstrSync) Execute(rt *thread) {
	threads := rt.get(instr.Threads).([]any)
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

type InstructionClose struct {
	Thread variables.Symbol
}

func (instr *InstructionClose) Execute(rt *thread) {
	thread := rt.get(instr.Thread).(variables.Thread)
	thread.Kill <- true
}

type InstrFetchTime struct {
	Result variables.Symbol
}

func (instr *InstrFetchTime) Execute(rt *thread) {
	rt.set(instr.Result, int(time.Since(rt.Runtime.startTime).Milliseconds()))
}
