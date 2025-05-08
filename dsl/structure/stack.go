package structure

type Stack[K any] []K

func NewStack[K any]() Stack[K] {
	return make(Stack[K], 0)
}

func (stack *Stack[K]) Push(val K) {
	*stack = append(*stack, val)
}

func (stack *Stack[K]) Pop() K {
	val := (*stack)[len(*stack)-1]
	*stack = (*stack)[0 : len(*stack)-1]
	return val
}

func (stack *Stack[K]) Peek() K {
	val := (*stack)[len(*stack)-1]
	return val
}

func (stack *Stack[K]) PeekRef() *K {
	return &(*stack)[len(*stack)-1]
}

func (stack *Stack[K]) PeekDepth(d int) *K {
	return &(*stack)[len(*stack)-d-1]
}
