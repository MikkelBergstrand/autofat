package statemanager

type EventType byte

type t_eventData struct {
	Timeout bool
	TestId  string
	Id      string
}

type triggerMessage struct {
	Type   trigger
	Params interface{}
}

type trigger int

const (
	TRIGGER_ARRIVE_FLOOR = iota + 1
	TRIGGER_DOOR
	TRIGGER_FLOOR_LIGHT
	TRIGGER_ORDER_LIGHT
	TRIGGER_OBSTRUCTION
	TRIGGER_DIRECTION
	TRIGGER_CRASH
	TRIGGER_OOB
	TRIGGER_NEW_LISTENER
)

func (t trigger) String() string {
	toStr := map[trigger]string{
		TRIGGER_ARRIVE_FLOOR: "ARRIVE_FLOOR",
		TRIGGER_DOOR:         "DOOR",
		TRIGGER_FLOOR_LIGHT:  "FLOOR_LIGHT",
		TRIGGER_ORDER_LIGHT:  "ORDER_LIGHT",
		TRIGGER_OBSTRUCTION:  "OBSTRUCTION",
		TRIGGER_DIRECTION:    "DIRECTION",
		TRIGGER_CRASH:        "CRASH",
		TRIGGER_OOB:          "OOB",
		TRIGGER_NEW_LISTENER: "NEW_LISTENER",
	}
	return toStr[t]
}
