package runtime

import (
	"autofat/config"
	"autofat/dsl/structure"
	"autofat/dsl/variables"
	"autofat/statemanager"
	"log"
)

const RT_EXIT = 1000000

// Holds the label of the function it is referring to.
type FunctionVar struct {
	Label        string
	AddressStack structure.Stack[address_stack_entry]
}

type address_stack_entry struct {
	start   int
	end     int
	runtime *RuntimeInstance
}

type Runtime struct {
	Instructions []Instruction
	Labels       map[string]int
	Config       config.Config
}

type RuntimeInstance struct {
	Retval         bool
	Runtime        *Runtime
	Parent         *RuntimeInstance
	Variables      []any
	Programcounter int
	CallStack      structure.Stack[ActivationRegister]
	Instances      []RuntimeInstance
}

type ActivationRegister struct {
	SavedPC      int
	Retval       variables.Symbol
	AddressStack structure.Stack[address_stack_entry]
	State        *statemanager.States
}

func New(config config.Config) *Runtime {
	runTime := Runtime{
		Labels: map[string]int{},
		Config: config,
	}

	return &runTime
}

func (runtime *Runtime) NewInstance(entryPoint int, parent *RuntimeInstance) *RuntimeInstance {
	first_ar := ActivationRegister{
		SavedPC: 0,
	}

	instance := RuntimeInstance{
		Retval:         true,
		Programcounter: entryPoint,
		Runtime:        runtime,
		Parent:         parent,
		Variables:      make([]any, 1000),
	}

	first_ar.AddressStack.Push(address_stack_entry{
		start:   0,
		end:     0,
		runtime: &instance,
	})
	instance.CallStack.Push(first_ar)
	return &instance
}

func (runtime *Runtime) GetLabel(label string) int {
	value, ok := runtime.Labels[label]
	if !ok {
		log.Fatalf("No such label %s", label)
	}
	return value
}

func (ar *ActivationRegister) getStackTop() int {
	return ar.AddressStack.PeekRef().end
}

func (ar *ActivationRegister) setStackTop(i int) {
	ar.AddressStack.PeekRef().end = i
}
func (rt *RuntimeInstance) PushAddress() {
	ar := rt.CallStack.PeekRef()
	ar.AddressStack.Push(address_stack_entry{
		runtime: rt,
		start:   ar.getStackTop() + 1,
		end:     ar.getStackTop() + 1})

	//fmt.Println("Adress stack pushed at ", ar)
}

func (rt *RuntimeInstance) PopAddress() {
	rt.CallStack.PeekRef().AddressStack.Pop()
}

func (runtime *RuntimeInstance) PushCall(func_address_stack structure.Stack[address_stack_entry], state *statemanager.States) {
	// The address stack in the function must have an address stack equal to how it looked
	// when the function was defined.
	top_of_callstack := runtime.CallStack.Peek()
	addr_stack := func_address_stack.Copy()

	// The beginning of the next address stack then begins at the next avaiable address
	addr_stack.Push(address_stack_entry{
		start:   top_of_callstack.getStackTop() + 1,
		end:     top_of_callstack.getStackTop() + 1,
		runtime: runtime,
	})

	runtime.CallStack.Push(ActivationRegister{
		SavedPC:      runtime.Programcounter + 1,
		AddressStack: addr_stack.Copy(),
		State:        state,
	})

	//fmt.Println("PushCall with AR = ", runtime.CallStack.PeekRef().getStackTop(), func_address_stack)
}

func (runtime *RuntimeInstance) PopCall() {
	val := runtime.CallStack.Pop()
	runtime.Programcounter = val.SavedPC - 1
	if len(runtime.CallStack) > 0 {
		//fmt.Println("PopCall, AR = ", runtime.CallStack.PeekRef().getStackTop())
	}
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
		//color.Println(color.Yellow, &runtime, reflect.TypeOf(runtime.Runtime.Instructions[runtime.Programcounter]), "PC = ", runtime.Programcounter)
		runtime.Runtime.Instructions[runtime.Programcounter].Execute(runtime)
		runtime.Programcounter += 1
	}

	done <- runtime.Retval
}

func (runtime *RuntimeInstance) Fork(entryPoint int, addressStack structure.Stack[address_stack_entry]) *RuntimeInstance {
	new_runtime := runtime.Runtime.NewInstance(entryPoint, runtime)
	new_addr_stack := addressStack.Copy()
	new_addr_stack.Push(address_stack_entry{
		runtime: new_runtime,
		start:   0,
		end:     0})
	new_runtime.CallStack.PeekRef().AddressStack = new_addr_stack
	return new_runtime
}

func (r *RuntimeInstance) AddressFromSymbol(symbol variables.Symbol) (*RuntimeInstance, int) {
	top_of_callstack := r.CallStack.PeekRef()

	//fmt.Println("Resolving address symbol", symbol, top_of_callstack.AddressStack)
	ar := top_of_callstack.AddressStack[len(top_of_callstack.AddressStack)-1-symbol.Scope]
	return ar.runtime, ar.start + symbol.Offset
}

func (s *RuntimeInstance) Get(symbol variables.Symbol) any {
	rt, addr := s.AddressFromSymbol(symbol)
	resolve := rt.Variables[addr]

	//fmt.Printf("Get %v addr=%d rt=%p\n", symbol, addr, rt)
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
		stack_top := s.CallStack.PeekRef().getStackTop()
		if addr > stack_top {
			s.CallStack.PeekRef().setStackTop(addr)
			//fmt.Printf("Stack top incremented to %d\n", addr)
		}
	}
	//fmt.Printf("Set %v value=%v addr=%d rt=%p\n", symbol, value, addr, rt)
}

func (rt *RuntimeInstance) Exit(value bool) {
	rt.Programcounter = RT_EXIT
	rt.Retval = value

	//Propagate exit.
	if rt.Parent != nil {
		rt.Parent.Exit(value)
	}
}
