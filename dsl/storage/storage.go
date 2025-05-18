package storage

import (
	"autofat/dsl/runtime"
	"autofat/dsl/variables"
	"fmt"
	"log"
	"strconv"
)

type Compiler struct {
	CurrentScope *scope
	LabelIndex   int //Used for auto-generated labels. They must be unique across scopes.
	NextLabel    string
}

type scope struct {
	Parent       *scope
	Variables    map[string]variables.SymbolTableEntry
	Offset       int
	Instructions []runtime.InstructionLabelPair //Instructions and associated label from statements/expressions in the local scope.
}

func newScopedStorage() scope {
	return scope{
		Variables: make(map[string]variables.SymbolTableEntry),
	}
}

func NewStorage() Compiler {
	storage := Compiler{}
	first_scope := newScopedStorage()
	storage.CurrentScope = &first_scope
	return storage
}

func (s *Compiler) NewFunction(name string, definition variables.TypeDefinition) error {
	func_symbol, err := s.NewVariable(definition, name)
	if err != nil {
		return err
	}

	label := s.NewAutoLabel()
	s.LoadInstruction(&runtime.InstrLoadFunction{
		Symbol: *func_symbol,
		Label:  label,
	})

	err = s.newFunctionScope(definition)
	if err != nil {
		return err
	}
	s.NewLabel(label)
	return nil
}

func (s *Compiler) NewImplicitFunction(definition variables.TypeDefinition) variables.Symbol {
	func_symbol := s.NewLiteral(definition)
	label := s.NewAutoLabel()

	s.LoadInstruction(&runtime.InstrLoadFunction{
		Symbol: func_symbol,
		Label:  label,
	})

	s.newFunctionScope(definition)
	s.NewLabel(label)
	return func_symbol
}

func (s *Compiler) newFunctionScope(definition variables.TypeDefinition) error {
	s.NewScope()

	// Create variable entries for the arguments. They are placed first in the function's symbol table
	for _, arg := range definition.ArgumentList {
		_, err := s.NewVariable(arg.Definition, arg.Identifier)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Compiler) NewScope() *scope {
	new_scope := newScopedStorage()
	new_scope.Parent = s.CurrentScope
	s.CurrentScope = &new_scope
	return s.CurrentScope
}

func (s *Compiler) DestroyFunctionScope(runTime *runtime.Runtime) (int, int) {
	start, end := runTime.LoadInstructions(s.CurrentScope.Instructions)

	s.CurrentScope = s.CurrentScope.Parent

	return start, end
}

func (s *Compiler) DestroyScope() {
	instructions := s.CurrentScope.Instructions

	s.CurrentScope = s.CurrentScope.Parent
	s.CurrentScope.Instructions = append(s.CurrentScope.Instructions, instructions...)

}

func (s *Compiler) NewLiteral(vartype variables.TypeDefinition) variables.Symbol {
	s.CurrentScope.Offset += 1
	sym := variables.Symbol{Scope: 0, Offset: s.CurrentScope.Offset - 1, Type: vartype}
	return sym
}

func (s *Compiler) NewVariable(vartype variables.TypeDefinition, name string) (*variables.Symbol, error) {
	_, exists := s.CurrentScope.Variables[name]
	if exists {
		log.Fatalf("Redeclaration of variable: %s\n", name)
		return nil, fmt.Errorf("redeclaration of variable: %s\n", name)
	}

	addr := s.CurrentScope.Offset
	s.CurrentScope.Variables[name] = variables.SymbolTableEntry{
		Type:   vartype,
		Offset: addr,
	}
	s.CurrentScope.Offset += 1

	return &variables.Symbol{Scope: 0, Offset: s.CurrentScope.Offset - 1, Type: vartype}, nil
}

func (c *Compiler) GetNamedSymbol(name string) (variables.Symbol, error) {
	scope := c.CurrentScope
	symbol, ok := variables.SymbolTableEntry{}, false
	scopeOffset := 0
	for {
		symbol, ok = (*scope).Variables[name]
		if ok {
			break
		}
		scope = (*scope).Parent
		if scope == nil {
			break
		}
		scopeOffset += 1
	}

	if !ok {
		return variables.Symbol{}, fmt.Errorf("could not resolve variable name: %s", name)
	}
	return variables.Symbol{
		Scope:  scopeOffset,
		Offset: symbol.Offset,
		Type:   symbol.Type,
	}, nil
}

func (s *Compiler) LoadLabeledInstruction(instruction runtime.Instruction, label string) *runtime.InstructionLabelPair {
	instr := s.LoadInstruction(instruction)
	instr.Label = label
	return instr
}
func (s *Compiler) LoadInstruction(instruction runtime.Instruction) *runtime.InstructionLabelPair {
	s.CurrentScope.Instructions = append(s.CurrentScope.Instructions, runtime.InstructionLabelPair{
		Instruction: instruction,
		Label:       s.NextLabel,
	})
	s.NextLabel = ""
	return &s.CurrentScope.Instructions[len(s.CurrentScope.Instructions)-1]
}

func (s *Compiler) InsertInstructionAt(instruction runtime.Instruction, label string, offset int) {

}

func (s *Compiler) NewLabel(label string) {
	if s.NextLabel != "" {
		log.Fatalf("Label %s overridden by %s!", s.NextLabel, label)
	}
	s.NextLabel = label
}

func (s *Compiler) NewAutoLabel() (label string) {
	s.LabelIndex += 1
	return strconv.Itoa(s.LabelIndex)
}
