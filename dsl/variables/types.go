package variables

import (
	"autofat/statemanager"
	"log"
	"time"
)

type Type int

const (
	INVALID Type = iota
	INT
	BOOL
	FUNC
	STATE
	CHAN
	ORDERTYPE
	NONE
	ANY
	UNDETERMINED
	THREAD
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
	case UNDETERMINED:
		return "undetermined"
	case ORDERTYPE:
		return "type"
	case ANY:
		return "any"
	case THREAD:
		return "thread"
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

type AwaitVal struct {
	StateChan statemanager.StateChannel
	Timeout   *time.Timer
}

// Potential result of an await evaluation
const (
	AWAIT_STATE_OK    = 0
	AWAIT_STATE_NOTOK = 1
	AWAIT_TIMEOUT     = 2
)

type Thread struct {
	Done chan bool
}
