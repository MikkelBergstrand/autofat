package runtime

import (
	"autofat/config"
	"autofat/dsl/structure"
	"autofat/dsl/variables"
	"autofat/statemanager"
	"log"
)

const RT_EXIT = 1000000

type Runtime struct {
	Instructions []Instruction
	Labels       map[string]int
	Config       config.Config
}

type RuntimeInstance struct {
	Retval          bool
	Runtime         *Runtime
	Parent          *RuntimeInstance
	ParentAddrStack structure.Stack[int]
	Variables       []any
	Programcounter  int
	CallStack       structure.Stack[ActivationRegister]
	Instances       []RuntimeInstance
}

type ActivationRegister struct {
	SavedPC      int
	Retval       variables.Symbol
	AddressStack structure.Stack[int]
	AddressBegin int
	StackTop     int
	State        *statemanager.States
}

func New(config config.Config) *Runtime {
	runTime := Runtime{
		Labels: map[string]int{},
		Config: config,
	}

	return &runTime
}

func (runtime *Runtime) NewInstance(entryPoint int, parent *RuntimeInstance, parent_addressStack structure.Stack[int]) RuntimeInstance {
	first_ar := ActivationRegister{
		SavedPC:      0,
		AddressBegin: 0,
	}
	first_ar.AddressStack.Push(0)
	instance := RuntimeInstance{
		Retval:          true,
		Programcounter:  entryPoint,
		Runtime:         runtime,
		Parent:          parent,
		ParentAddrStack: parent_addressStack,
		Variables:       make([]any, 1000),
	}
	instance.CallStack.Push(first_ar)
	return instance
}

func (runtime *Runtime) GetLabel(label string) int {
	value, ok := runtime.Labels[label]
	if !ok {
		log.Fatalf("No such label %s", label)
	}
	return value
}

func (ar *ActivationRegister) PushAddress() {
	ar.AddressStack.Push(ar.StackTop + 1)
	//fmt.Println("Adress stack pushed at ", ar)
}

func (ar *ActivationRegister) PopAddress() {
	ar.StackTop = ar.AddressStack.Pop()
}

func (runtime *RuntimeInstance) PushCall(func_address_stack structure.Stack[int], state *statemanager.States) {
	// The address stack in the function must have an address stack equal to how it looked
	// when the function was defined.
	top_of_callstack := runtime.CallStack.Peek()
	var addr_stack structure.Stack[int]
	for i := range func_address_stack {
		addr_stack.Push(func_address_stack[i])
	}
	// The beginning of the next address stack then begins at the next avaiable address
	addr_stack.Push(top_of_callstack.StackTop + 1)
	runtime.CallStack.Push(ActivationRegister{
		SavedPC:      runtime.Programcounter + 1,
		AddressStack: addr_stack,
		AddressBegin: top_of_callstack.StackTop + 1,
		State:        state,
	})

	//fmt.Println("PushCall with AR = ", runtime.CallStack.Peek(), func_address_stack)
}

func (runtime *RuntimeInstance) PopCall() {
	val := runtime.CallStack.Pop()
	runtime.Programcounter = val.SavedPC - 1
	//fmt.Println("PopCall, AR = ", runtime.CallStack.Peek())
}

// Add the set of instructions. Return the first and last index of the inserted instructions.
func (runtime *Runtime) LoadInstructions(instructions []InstructionLabelPair) (start int, end int) {
	for _, pair := range instructions {
		runtime.Instructions = append(runtime.Instructions, pair.Instruction)
		if pair.Label != "" {
			runtime.Labels[pair.Label] = len(runtime.Instructions) - 1
		}
	}
	start, end = len(runtime.Instructions)-len(instructions), len(runtime.Instructions)-1
	return start, end
}

func (runTime *Runtime) NextInstruction() int {
	return len(runTime.Instructions)
}

func (runtime *RuntimeInstance) Run(done chan bool) {
	for runtime.Programcounter != RT_EXIT+1 {
		//color.Println(color.Yellow, reflect.TypeOf(runtime.Runtime.Instructions[runtime.Programcounter]), "PC = ", runtime.Programcounter)
		runtime.Runtime.Instructions[runtime.Programcounter].Execute(runtime)
		runtime.Programcounter += 1
	}

	done <- runtime.Retval
}

func (runtime *RuntimeInstance) Fork(entryPoint int, caller_address_stack structure.Stack[int]) *RuntimeInstance {
	new_runtime := runtime.Runtime.NewInstance(entryPoint, runtime, caller_address_stack)
	return &new_runtime
}

func (r *RuntimeInstance) AddressFromSymbol(symbol variables.Symbol) (*RuntimeInstance, int) {
	top_of_callstack := r.CallStack.PeekRef()
	len_address_stack := len(top_of_callstack.AddressStack)

	//fmt.Println("Resolving address symbol", symbol, top_of_callstack.AddressStack)
	if symbol.Scope >= len_address_stack {
		ar := r.ParentAddrStack[len(r.ParentAddrStack)-1-(symbol.Scope-len_address_stack)]
		return r.Parent, ar + symbol.Offset
	} else {
		ar := top_of_callstack.AddressStack[len(top_of_callstack.AddressStack)-1-symbol.Scope]
		return r, ar + symbol.Offset
	}
}

func (s *RuntimeInstance) Get(symbol variables.Symbol) any {
	rt, addr := s.AddressFromSymbol(symbol)
	resolve := rt.Variables[addr]

	//fmt.Println("Get", symbol, "val=", resolve, "addr=", addr)
	return resolve
}

func (r *RuntimeInstance) GetInt(symbol variables.Symbol) int {
	val := r.Get(symbol).(int)

	return val
}

func (r *RuntimeInstance) GetBool(symbol variables.Symbol) bool {
	return r.Get(symbol).(bool)
}

func (r *RuntimeInstance) GetState() *statemanager.States {
	for i := len(r.CallStack) - 1; i >= 0; i-- {
		if r.CallStack[i].State != nil {
			return r.CallStack[i].State
		}
	}
	return nil
}

func (s *RuntimeInstance) Set(symbol variables.Symbol, value any) {
	rt, addr := s.AddressFromSymbol(symbol)
	rt.Variables[addr] = value
	if s == rt {
		stack_top := &s.CallStack.PeekRef().StackTop
		if addr > *stack_top {
			*stack_top = addr
		}
	}
	//fmt.Println("Set", symbol, "value=", value, "addr=", addr)
}

func (rt *RuntimeInstance) Exit(value bool) {
	rt.Programcounter = RT_EXIT
	rt.Retval = value

}
