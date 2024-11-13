package events

import (
	"autofat/elevio"
	"fmt"
	"math/rand/v2"
	"time"
)

type EventType byte

type TestConditionFunction func([]ElevatorState) bool

type SafetyAssert struct {
	Condition   TestConditionFunction
	AllowedTime time.Duration
	assert      int
	C           chan<- bool
}

func (w *SafetyAssert) IsAsserted() bool {
	return w.assert > 0
}

func (w *SafetyAssert) Abort() {
	if !w.IsAsserted() {
		return
	}
	w.assert = 0
}

func (w *SafetyAssert) Assert() {
	if w.IsAsserted() {
		return
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

}

type WaitFor struct {
	ID          string
	Condition   TestConditionFunction
	Timeout     time.Duration
	C           chan<- bool
	Chan_Result chan<- bool
	triggered   bool
}

// If what we are waiting for does not happen in the allotted time,
// send something on the failure channel
func (w *WaitFor) Watchdog() {
	timer := time.NewTimer(w.Timeout)
	<-timer.C

	//Result is false, we failed (unless we already succeeded).
	if w.triggered {
		return
	}
	w.triggered = true
	fmt.Println("WaitFor ", w.ID, "Timeout")
	w.Chan_Result <- false
}

// Trigger the WaitFor by putting out on the channel.
// Channel may only put out once.
func (w *WaitFor) Trigger() {
	if w.triggered {
		return
	}

	fmt.Println("WaitFor ", w.ID, "Trigger")
	w.triggered = true
	w.C <- true
}

type Trigger int

const (
	TRIGGER_ARRIVE_FLOOR = iota + 1
	TRIGGER_DOOR_OPEN
	TRIGGER_DOOR_CLOSE
	TRIGGER_FLOOR_LIGHT
	TRIGGER_ORDER_LIGHT
	TRIGGER_OBSTRUCTION
)

func (t Trigger) String() string {
	toStr := map[Trigger]string{
		TRIGGER_ARRIVE_FLOOR: "ARRIVE_FLOOR",
		TRIGGER_DOOR_OPEN:    "DOOR_OPEN",
		TRIGGER_DOOR_CLOSE:   "DOOR_CLOSE",
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
