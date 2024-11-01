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
  Condition TestConditionFunction
  AllowedTime time.Duration
  assert int 
  C <-chan bool
}

func (w* SafetyAssert) IsAsserted() bool {
  return w.assert > 0 
}

func (w* SafetyAssert) Abort() {
  if(!w.IsAsserted()) {
    return 
  }
  w.assert = 0
}

func (w* SafetyAssert) Assert() {
  if(w.IsAsserted()) {
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
      fmt.Println("fire")
      <-w.C
    }
  }()
  
}



type WaitFor struct {
  Condition TestConditionFunction
  Timeout time.Duration
  C   chan bool
  triggered bool
}

//Trigger the WaitFor by putting out on the channel.
//Channel may only put out once.
func (w WaitFor) Trigger() {
  if(w.triggered) {
    return
  }

  w.triggered = true
  w.C <- true
}


type Trigger int
const (
	TRIGGER_INIT Trigger = iota +1 
	TRIGGER_TIMER        
  TRIGGER_ARRIVE_FLOOR 
	TRIGGER_DOOR_OPEN    
	TRIGGER_DOOR_CLOSE   
	TRIGGER_FLOOR_LIGHT  
	TRIGGER_LOAD         
	TRIGGER_DELAY        
	TRIGGER_ORDER_LIGHT  
)

func (t Trigger) String() string {
  toStr := map[Trigger]string {
    TRIGGER_ARRIVE_FLOOR: "ARRIVE_FLOOR",
    TRIGGER_DOOR_OPEN: "DOOR_OPEN",
    TRIGGER_DOOR_CLOSE: "DOOR_CLOSE",
    TRIGGER_FLOOR_LIGHT: "FLOOR_LIGHT",
    TRIGGER_ORDER_LIGHT: "ORDER_LIGHT",
  }
  return toStr[t]
}

type Action int
const (
	ACTION_NONE Action = iota+1
	ACTION_OPEN_DOOR  
	ACTION_MAKE_ORDER 
	ACTION_CLOSE_DOOR 
)

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
