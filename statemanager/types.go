package statemanager

import (
	"fmt"
	"math/rand/v2"
	"time"
)

type EventType byte

type TestConditionFunction func([]ElevatorState) bool

type t_eventData struct {
	Timeout bool
	TestId  string
	Id      string
}

type t_assert struct {
	Condition   TestConditionFunction
	AllowedTime time.Duration
	assert      int
	Data        t_eventData
}

func (w t_assert) IsAsserted() bool {
	return w.assert > 0
}

func (w t_assert) Abort() t_assert {
	if w.IsAsserted() {
		fmt.Println("Safety assert ", w.Data, "aborted")
		w.assert = 0
	}
	return w
}

func (w t_assert) Assert() t_assert {
	if w.IsAsserted() {
		return w
	}

	//Assign the assert "unique" id, so we know that
	//if the assert id is the same after the allowed time,
	//we know that that assertion event is what kept the assert alive.
	w.assert = rand.Int()

	return w

}

type t_await struct {
	Condition     TestConditionFunction
	Timeout       time.Duration
	chan_internal chan t_eventData
	triggered     bool
	timer         *time.Timer
	Data          t_eventData
}

// If what we are waiting for does not happen in the allotted time,
// send something on the failure channel
func awaitWatchDog(id string) {
	w := _awaits[id]
	//Result is false, we failed (unless we already succeeded).
	if w.triggered {
		return
	}

	w.timer = time.NewTimer(w.Timeout)
	<-w.timer.C

	//Get updated struct. See if it has triggered or it has been deleted..
	w, ok := _awaits[id]
	if !ok {
		return
	}
	w.triggered = true
	_awaits[id] = w

	fmt.Println("Await ", w.Data.Id, "Timeout")
	w.Data.Timeout = true
	w.chan_internal <- w.Data
}

// Trigger the Await by putting out on the channel.
// Channel may only put out once.
func (w t_await) Trigger() t_await {
	if w.triggered {
		return w
	}

	fmt.Println("Await ", w.Data.Id, "Trigger")
	if w.timer != nil {
		w.timer.Reset(w.Timeout)
	}
	w.triggered = true
	w.Data.Timeout = false
	w.chan_internal <- w.Data
	return w
}

func (w t_await) Delete() t_await {
	w.triggered = true
	fmt.Println("Deleting await", w.Data)
	return w
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
	}
	return toStr[t]
}
