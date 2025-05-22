package runtime

import (
	"autofat/config"
	"autofat/dsl/color"
	"autofat/dsl/structure"
	"autofat/dsl/variables"
	"autofat/statemanager"
	"log"
	"reflect"
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
	runtime *thread
}

type Runtime struct {
	Instructions []Instruction
	Labels       map[string]int
	Config       config.Config
}

type thread struct {
	Retval         bool
	Runtime        *Runtime
	Parent         *thread
	Variables      []any
	Programcounter int
	CallStack      structure.Stack[ActivationRegister]
	Children       []*thread
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

func (runtime *Runtime) NewInstance(entryPoint int, parent *thread) *thread {
	first_ar := ActivationRegister{
		SavedPC: 0,
	}

	instance := thread{
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

func (runtime *Runtime) getLabel(label string) int {
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
func (rt *thread) pushAddress() {
	ar := rt.CallStack.PeekRef()
	ar.AddressStack.Push(address_stack_entry{
		runtime: rt,
		start:   ar.getStackTop() + 1,
		end:     ar.getStackTop() + 1})

	//fmt.Println("Adress stack pushed at ", ar)
}

func (rt *thread) popAddress() {
	rt.CallStack.PeekRef().AddressStack.Pop()
}

func (runtime *thread) pushCall(func_address_stack structure.Stack[address_stack_entry], state *statemanager.States) {
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

func (runtime *thread) popCall() {
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

func (runtime *thread) Run(done chan bool) {
	for runtime.Programcounter != RT_EXIT+1 {
		color.Println(color.Yellow, &runtime, reflect.TypeOf(runtime.Runtime.Instructions[runtime.Programcounter]), "PC = ", runtime.Programcounter)
		runtime.Runtime.Instructions[runtime.Programcounter].Execute(runtime)
		runtime.Programcounter += 1
	}

	done <- runtime.Retval
}

func (runtime *thread) fork(entryPoint int, addressStack structure.Stack[address_stack_entry]) *thread {
	new_runtime := runtime.Runtime.NewInstance(entryPoint, runtime)
	new_addr_stack := addressStack.Copy()
	new_addr_stack.Push(address_stack_entry{
		runtime: new_runtime,
		start:   0,
		end:     0})
	new_runtime.CallStack.PeekRef().AddressStack = new_addr_stack
	runtime.Children = append(runtime.Children, new_runtime)
	return new_runtime
}

func (r *thread) addressFromSymbol(symbol variables.Symbol) (*thread, int) {
	top_of_callstack := r.CallStack.PeekRef()

	//fmt.Println("Resolving address symbol", symbol, top_of_callstack.AddressStack)
	ar := top_of_callstack.AddressStack[len(top_of_callstack.AddressStack)-1-symbol.Scope]
	return ar.runtime, ar.start + symbol.Offset
}

func (s *thread) get(symbol variables.Symbol) any {
	rt, addr := s.addressFromSymbol(symbol)
	resolve := rt.Variables[addr]

	//fmt.Printf("Get %v addr=%d rt=%p\n", symbol, addr, rt)
	return resolve
}

func (r *thread) getInt(symbol variables.Symbol) int {
	val := r.get(symbol).(int)

	return val
}

func (r *thread) getBool(symbol variables.Symbol) bool {
	return r.get(symbol).(bool)
}

func (r *thread) getState() *statemanager.States {
	for i := len(r.CallStack) - 1; i >= 0; i-- {
		if r.CallStack[i].State != nil {
			return r.CallStack[i].State
		}
	}
	return nil
}

func (s *thread) set(symbol variables.Symbol, value any) {
	rt, addr := s.addressFromSymbol(symbol)
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

func (rt *thread) exit(value bool) {
	rt.Programcounter = RT_EXIT
	rt.Retval = value

	//Propagate exit.
	if rt.Parent != nil && rt.Programcounter < RT_EXIT {
		rt.Parent.exit(value)
	}

	for i := range rt.Children {
		if rt.Children[i].Programcounter < RT_EXIT {
			rt.Children[i].exit(value)
		}
	}
	clear(rt.Children)
}
