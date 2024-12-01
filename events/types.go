package events

import (
	"autofat/elevio"
	"fmt"
	"math/rand/v2"
	"time"
)

type EventType byte

type TestConditionFunction func([]ElevatorState) bool

type EventMetadata struct {
	Timeout bool
	TestId  string
	Id      string
}

type AssertObj struct {
	Condition   TestConditionFunction
	AllowedTime time.Duration
	assert      int
	internal    chan<- bool
	C           chan<- bool
	Data        EventMetadata
}

func (w AssertObj) IsAsserted() bool {
	return w.assert > 0
}

func (w AssertObj) Abort() AssertObj {
	if w.IsAsserted() {
		fmt.Println("Safety assert ", w.Data, "aborted")
		w.assert = 0
	}
	return w
}

func (w AssertObj) Assert() AssertObj {
	if w.IsAsserted() {
		return w
	}

	//Assign the assert "unique" id, so we know that
	//if the assert id is the same after the allowed time,
	//we know that that assertion event is what kept the assert alive.
	w.assert = rand.Int()

	return w

}

type AwaitObj struct {
	Condition     TestConditionFunction
	Timeout       time.Duration
	chan_internal chan EventMetadata
	triggered     bool
	timer         *time.Timer
	Data          EventMetadata
	C_Timeout     chan bool
	C_OK          chan bool
}

// If what we are waiting for does not happen in the allotted time,
// send something on the failure channel
func AwaitWatchdog(id string) {
	w := _awaits[id]
	//Result is false, we failed (unless we already succeeded).
	if w.triggered {
		return
	}

	fmt.Println(w)
	w.timer = time.NewTimer(w.Timeout)
	<-w.timer.C

	//Get updated struct. See if it has triggered or it has been deleted..
	w, ok := _awaits[id]
	if !ok {
		return
	}
	w.triggered = true
	_awaits[id] = w

	fmt.Println("WaitFor ", w.Data.Id, "Timeout")
	w.Data.Timeout = true
	w.chan_internal <- w.Data
}

// Trigger the WaitFor by putting out on the channel.
// Channel may only put out once.
func (w AwaitObj) Trigger() AwaitObj {
	if w.triggered {
		return w
	}

	fmt.Println("WaitFor ", w.Data.Id, "Trigger")
	if w.timer != nil {
		w.timer.Reset(w.Timeout)
	}
	w.triggered = true
	w.Data.Timeout = false
	w.chan_internal <- w.Data
	return w
}

func (w AwaitObj) Delete() AwaitObj {
	w.triggered = true
	return w
}

type TriggerMessage struct {
	Type   Trigger
	Params interface{}
}

type Trigger int

const (
	TRIGGER_ARRIVE_FLOOR = iota + 1
	TRIGGER_DOOR
	TRIGGER_FLOOR_LIGHT
	TRIGGER_ORDER_LIGHT
	TRIGGER_OBSTRUCTION
	TRIGGER_SAFETYASSERT
	TRIGGER_UNTILASSERT
	TRIGGER_DIRECTION
)

func (t Trigger) String() string {
	toStr := map[Trigger]string{
		TRIGGER_ARRIVE_FLOOR: "ARRIVE_FLOOR",
		TRIGGER_DOOR:         "DOOR",
		TRIGGER_FLOOR_LIGHT:  "FLOOR_LIGHT",
		TRIGGER_ORDER_LIGHT:  "ORDER_LIGHT",
		TRIGGER_OBSTRUCTION:  "OBSTRUCTION",
		TRIGGER_DIRECTION:    "DIRECTION",
	}
	return toStr[t]
}

type Action int

type Floor struct {
	Elevator int
	Floor    int
}

type Button struct {
	elevio.ButtonEvent
	Elevator int
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

func (es *ElevatorState) OrderLight(btn elevio.ButtonType, floor int) bool {
	switch btn {
	case elevio.BT_HallDown:
		return es.HallDownLights[floor]
	case elevio.BT_HallUp:
		return es.HallUpLights[floor]
	case elevio.BT_Cab:
		return es.CabLights[floor]
	default:
		panic("Invalid elevio.ButtonType")
	}
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
