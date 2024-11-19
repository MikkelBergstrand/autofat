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
	TestId string
	Id     string
}

type SafetyAssert struct {
	Condition   TestConditionFunction
	AllowedTime time.Duration
	assert      int
	C           chan<- bool
	Data        EventMetadata
}

func (w SafetyAssert) IsAsserted() bool {
	return w.assert > 0
}

func (w SafetyAssert) Abort() SafetyAssert {
	if w.IsAsserted() {
		w.assert = 0
	}
	return w
}

func (w SafetyAssert) Assert() SafetyAssert {
	if w.IsAsserted() {
		return w
	}

	//Assign the assert "unique" id, so we know that
	//if the assert id is the same after the allowed time,
	//we know that that assertion event is what kept the assert alive.
	w.assert = rand.Int()

	go func() {
		assert_at_beginning := w.assert
		timer := time.NewTimer(w.AllowedTime)
		<-timer.C
		if w.assert == assert_at_beginning {
			w.C <- false
		}
	}()

	return w

}

type WaitFor struct {
	Condition    TestConditionFunction
	Timeout      time.Duration
	Chan_OK      chan<- bool
	Chan_Timeout chan<- EventMetadata
	triggered    bool
	timer        *time.Timer
	Data         EventMetadata
}

// If what we are waiting for does not happen in the allotted time,
// send something on the failure channel
func AwaitWatchdog(id string) {
	w := _untilAsserts[id]
	//Result is false, we failed (unless we already succeeded).
	if w.triggered {
		return
	}

	w.timer = time.NewTimer(w.Timeout)
	<-w.timer.C

	//Get updated struct. See if it has triggered or it has been deleted..
	w, ok := _untilAsserts[id]
	if !ok {
		return
	}
	w.triggered = true
	_untilAsserts[id] = w

	fmt.Println("WaitFor ", w.Data.Id, "Timeout")
	w.Chan_Timeout <- w.Data
}

// Trigger the WaitFor by putting out on the channel.
// Channel may only put out once.
func (w WaitFor) Trigger() WaitFor {
	if w.triggered {
		return w
	}

	fmt.Println("WaitFor ", w.Data.Id, "Trigger")
	if w.timer != nil {
		w.timer.Reset(w.Timeout)
	}
	w.triggered = true
	w.Chan_OK <- true
	return w
}

func (w WaitFor) Delete() WaitFor {
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
)

func (t Trigger) String() string {
	toStr := map[Trigger]string{
		TRIGGER_ARRIVE_FLOOR: "ARRIVE_FLOOR",
		TRIGGER_DOOR:         "DOOR",
		TRIGGER_FLOOR_LIGHT:  "FLOOR_LIGHT",
		TRIGGER_ORDER_LIGHT:  "ORDER_LIGHT",
		TRIGGER_OBSTRUCTION:  "OBSTRUCTION",
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
