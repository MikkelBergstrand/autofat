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
	STRING
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
	case STRING:
		return "string"
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

type AssertVal struct {
	StateChan     statemanager.StateChannel
	DeadzoneTimer *time.Timer
	TimerActive   bool
}

type Thread struct {
	Done chan bool
}
