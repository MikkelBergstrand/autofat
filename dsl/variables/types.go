package variables

import (
	"log"
)

type Type int

const (
	INVALID Type = iota
	INT
	BOOL
	FUNC
	STATE
	CHAN
	ARRAY
	NONE
)

func (t Type) String() string {
	switch t {
	case INT:
		return "int"
	case BOOL:
		return "bool"
	case NONE:
		return "void"
	case FUNC:
		return "func"
	case STATE:
		return "state"
	case CHAN:
		return "chan"
	case ARRAY:
		return "array"
	case INVALID:
		return ""
	}
	log.Panicln("Invalid variable type!")
	return ""
}

type Symbol struct {
	Scope  int
	Offset int
	Type   TypeDefinition
}

type SymbolTableEntry struct {
	Offset int
	Type   TypeDefinition
}
