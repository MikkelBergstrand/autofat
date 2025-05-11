package variables

import (
	"autofat/dsl/structure"
	"fmt"
	"strings"
)

type Argument struct {
	Definition TypeDefinition
	Identifier string
}

func (arg Argument) String() string {
	return fmt.Sprintf("%s %s", arg.Definition.String(), arg.Identifier)
}

type ArgumentList []Argument
type TypeDefinition struct {
	BaseType Type

	//Used if type is a function pointer.
	ArgumentList ArgumentList
	ReturnType   *TypeDefinition
}

func (arg TypeDefinition) String() string {
	s := ""
	s += arg.BaseType.String()

	if arg.BaseType == FUNC {
		s += " ("
		var arg_strings []string
		for _, arg := range arg.ArgumentList {
			arg_strings = append(arg_strings, arg.String())
		}
		s += strings.Join(arg_strings, ",")
		s += ") "

		s += arg.ReturnType.String()
	}
	return s
}

// Check for type equality
func (a TypeDefinition) Equals(b TypeDefinition) bool {
	if a.BaseType != b.BaseType {
		return false
	}

	if a.BaseType == FUNC {
		if len(a.ArgumentList) != len(b.ArgumentList) {
			return false
		}
		if !a.ReturnType.Equals(*b.ReturnType) {
			return false
		}

		for i := range a.ArgumentList {
			if !a.ArgumentList[i].Definition.Equals(b.ArgumentList[i].Definition) {
				return false
			}
		}
	}

	return true
}

func GetBaseTypeDef(t Type) TypeDefinition {
	return TypeDefinition{BaseType: t}
}

// Holds the label of the function it is referring to.
type FunctionVar struct {
	Label        string
	AddressStack structure.Stack[int]
}

// Verify an argument list of symbols.
// If the argument list length and type of each argument does not match, return false.
func (list ArgumentList) ValidateArgumentList(symbols []Symbol) bool {
	if len(symbols) != len(list) {
		return false
	}

	for i := range list {
		if !symbols[i].Type.Equals(list[i].Definition) {
			return false
		}
	}

	return true
}
