package runtime

import (
	"autofat/dsl/color"
	"autofat/dsl/structure"
	"autofat/dsl/variables"
	"autofat/elevio"
	"autofat/simulator"
	"autofat/statemanager"
	"autofat/studentprogram"
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
	var address_stack structure.Stack[int]
	src_address_stack := runtime.CallStack.Peek().AddressStack
	for i := range src_address_stack {
		address_stack.Push(src_address_stack[i])
	}
	runtime.Set(instr.Symbol, variables.FunctionVar{
		Label:        instr.Label,
		AddressStack: address_stack,
	})
}

func (instr *InstrAssign) Execute(runtime *RuntimeInstance) {
	runtime.Set(instr.Dest, runtime.Get(instr.Source))
}

type InstructionEcho struct {
	A variables.Symbol
}

func (instr *InstructionEcho) Execute(runtime *RuntimeInstance) {
	color.Println(color.Green, runtime.Runtime.Variables[runtime.AddressFromSymbol(instr.A)])
}

type InstrCallFunction struct {
	Arguments     []variables.Symbol
	RetVal        variables.Symbol
	SymbolicLabel variables.Symbol
	State         *statemanager.States
}

func (instr *InstrCallFunction) Execute(runtime *RuntimeInstance) {
	//Fetch and copy argument values
	var arg_values []any
	for i := range instr.Arguments {
		arg_values = append(arg_values, runtime.Get(instr.Arguments[i]))
	}
	//Fetch func_ptr
	func_ptr := runtime.Get(instr.SymbolicLabel).(variables.FunctionVar)

	// Bind return value
	top_ar := runtime.CallStack.PeekRef()
	top_ar.Retval = instr.RetVal
	//fmt.Println("Bound ret val to", top_ar.Retval)

	// Account for prelude length
	runtime.PushCall(func_ptr.AddressStack, instr.State)

	// Once "inside" the function, load argument values
	for i := range arg_values {
		runtime.Set(variables.Symbol{Offset: i, Scope: 0, Type: instr.Arguments[i].Type}, arg_values[i])
	}

	//Then, jump to the function's label
	jmp_instr := InstrJmpVar{
		Label: func_ptr.Label,
	}
	jmp_instr.Execute(runtime)
}

type InstrBeginScope struct{}

func (instr *InstrBeginScope) Execute(runtime *RuntimeInstance) {
	runtime.CallStack.PeekRef().PushAddress()
}

type InstrEndScope struct{}

func (instr *InstrEndScope) Execute(runtime *RuntimeInstance) {
	runtime.CallStack.PeekRef().PopAddress()
}

type InstrExitFunction struct {
	RetVal variables.Symbol
}

func (instr *InstrExitFunction) Execute(runtime *RuntimeInstance) {
	ret_val := runtime.Get(instr.RetVal)

	runtime.PopCall()
	top_ar := runtime.CallStack.PeekRef()

	//fmt.Println("Ret val on exit", top_ar.Retval, ret_val)
	runtime.Set(top_ar.Retval, ret_val)
}

// Does nothing.
type InstrNOP struct{}

func (instr *InstrNOP) Execute(runtime *RuntimeInstance) {}

type InstrAwait struct {
	Channel            variables.Symbol
	StateFunction      variables.Symbol
	ConditionFuncValue variables.Symbol
	Timeout            variables.Symbol
}

func (instr *InstrAwait) Execute(runtime *RuntimeInstance) {
	// Wait for new state
	await_obj := runtime.Get(instr.Channel).(variables.AwaitVal)
	stateChan := await_obj.StateChan
	timeout := await_obj.Timeout

	var states statemanager.States = nil
	closed := false
	select {
	case states, closed = <-stateChan:
		if !closed {
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
}

func (instr *InstrEndAwait) Execute(runtime *RuntimeInstance) {
	await_val := runtime.GetBool(instr.ConditionFuncValue)

	fmt.Println(await_val)
	if !runtime.GetBool(instr.Timeout) && !await_val {
		jmp := InstrJmp{
			Label: instr.Label,
		}
		jmp.Execute(runtime)
	}
}

type InstrStateListen struct {
	Symbol         variables.Symbol
	TimeoutSeconds variables.Symbol
}

func (instr *InstrStateListen) Execute(rt *RuntimeInstance) {
	// Initialize a new state capturing channel
	// if none exist at the symbol location.
	if rt.Get(instr.Symbol) == nil {
		timeout := time.NewTimer(time.Duration(rt.GetInt(instr.TimeoutSeconds) * int(time.Millisecond)))
		stateChan := make(statemanager.StateChannel)
		statemanager.RegisterStateChannel(stateChan)
		rt.Set(instr.Symbol, variables.AwaitVal{
			Timeout:   timeout,
			StateChan: stateChan,
		})
	}
}

type InstrLoadArray struct {
	SrcSymbols []variables.Symbol
	DestSymbol variables.Symbol
}

func (instr *InstrLoadArray) Execute(rt *RuntimeInstance) {
	var arr []int
	for _, sym := range instr.SrcSymbols {
		arr = append(arr, rt.GetInt(sym))
	}
	rt.Set(instr.DestSymbol, arr)
}

type InstrInitializeElevators struct {
	ArraySymbol variables.Symbol
}

func (instr *InstrInitializeElevators) Execute(rt *RuntimeInstance) {
	arr := rt.Get(instr.ArraySymbol).([]int)
	n_elevators := len(arr)
	for i := 0; i < n_elevators; i++ {
		simulator.Init(rt.Runtime.Config.GetElevatorConfig(i), simulator.InitializationParams{InitialFloor: 0, BetweenFloors: false})
		simulator.Run(i)
	}

	time.Sleep(500 * time.Millisecond)
	studentprogram.InitalizeFromConfig(
		rt.Runtime.Config.StudentProgramWaitTime,
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

func (instr *InstrGetFloor) Execute(rt *RuntimeInstance) {
	arr := rt.Get(instr.ArraySymbol).([]int)
	state := *rt.GetState()

	if len(arr) == 0 {
		rt.Set(instr.Result, -1)
		return
	}

	val := state[0].Floor
	for i := 1; i < len(arr); i++ {
		if state[i].Floor != val {
			rt.Set(instr.Result, -1)
			return
		}
	}

	rt.Set(instr.Result, val)
}

type InstrGetStatusLight struct {
	ArraySymbol variables.Symbol
	OrderType   variables.Symbol
	Floor       variables.Symbol
	Result      variables.Symbol
}

func (instr *InstrGetStatusLight) Execute(rt *RuntimeInstance) {
	arr := rt.Get(instr.ArraySymbol).([]int)
	state := *rt.GetState()

	if len(arr) == 0 {
		rt.Set(instr.Result, -1)
		return
	}

	floor := rt.GetInt(instr.Floor)
	ordertype := rt.Get(instr.OrderType).(elevio.ButtonType)

	val := state[0].OrderLight(ordertype, floor)
	for i := 1; i < len(arr); i++ {
		if state[i].OrderLight(ordertype, floor) != val {
			rt.Set(instr.Result, -1)
			return
		}
	}

	if !val {
		rt.Set(instr.Result, 0)
	} else {
		rt.Set(instr.Result, 1)
	}

}
