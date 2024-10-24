package events

import "autofat/elevio"

type EventType byte

const (
	TRIGGER_INIT         = 1
	TRIGGER_TIMER        = 2
	TRIGGER_ARRIVE_FLOOR = 3
	TRIGGER_DOOR_OPEN    = 4
	TRIGGER_DOOR_CLOSE   = 5
	TRIGGER_FLOOR_LIGHT
)

const (
	ACTION_NONE       = 0
	ACTION_OPEN_DOOR  = 1
	ACTION_MAKE_ORDER = 2
)

type Floor struct {
	Elevator int
	Floor    int
}

func (f Floor) Equals(other Floor) bool {
	return f.Elevator == other.Elevator && f.Floor == other.Floor
}

type Button struct {
	elevio.ButtonEvent
	Elevator int
}

func (b Button) Equals(other Button) bool {
	return b.Button == other.Button && b.Floor == other.Floor && b.Elevator == other.Elevator
}

type Event struct {
	ID          string
	Description string
	RepeatCount byte

	TriggerType   byte
	TriggerParams interface{}

	ActionType byte

	ActionParams interface{}

	//Time (in milliseconds) from event is loaded to a timeout is triggered.
	TimeoutMillisec int

	LoadOnTrigger []string
	LoadOnRepeat  []string
	LoadOnFailure []string
}

type EventTriggerParams struct {
	Event        *Event
	Milliseconds int
}
