package events

import "autofat/elevio"

type EventType byte

const (
	TRIGGER_INIT         = 1
	TRIGGER_TIMER        = 2
	TRIGGER_ARRIVE_FLOOR = 3
	TRIGGER_DOOR_OPEN    = 4
	TRIGGER_DOOR_CLOSE   = 5
	TRIGGER_FLOOR_LIGHT  = 6
	TRIGGER_LOAD         = 7
	TRIGGER_DELAY        = 8
	TRIGGER_ORDER_LIGHT  = 9
)

const (
	ACTION_NONE       = 0
	ACTION_OPEN_DOOR  = 1
	ACTION_MAKE_ORDER = 2
	ACTION_CLOSE_DOOR = 3
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

type ElevatorState struct {
	FloorLamp   int
	Floor       int
	Direction   elevio.MotorDirection
	DoorOpen    bool
	Obstruction bool

	CabLights      []bool
	HallUpLights   []bool
	HallDownLights []bool
}

func InitElevatorState(nFloors int) ElevatorState {
	//Initialize a new ElevatorState object.
	ret := ElevatorState{
		FloorLamp:   -1,
		Floor:       -1,
		Direction:   elevio.MD_Stop,
		DoorOpen:    false,
		Obstruction: false,
	}

	ret.CabLights = make([]bool, nFloors)
	ret.HallDownLights = make([]bool, nFloors)
	ret.HallUpLights = make([]bool, nFloors)

	return ret
}
