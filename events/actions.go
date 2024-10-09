package events

import (
	"autofat/fatelevator"
	"fmt"
	"time"
)


type ElevatorState struct {
	Floor uint8			
}

type FloorEvent struct {
	ElevatorId uint8
	Floor uint8
}


type Event struct {
	WaitFor time.Duration	 //0 indicates unused
	ExpectedFloors []uint8	 //Floors we want the elevator to be at. 255 = dont care
	MustSatisfyAll bool		 //Indicate if one condition fulfilled is sufficient, if true; all must be set.
}

type eventState struct {
	Elevators []ElevatorState 
	Timer time.Timer
	TimerSatisified bool
	FloorsSatisfied bool
}


func (self eventState) isDone() bool {
	return self.TimerSatisified && self.FloorsSatisfied
}

func updateTimer(state *eventState, event *Event, ch_Done chan int) {
	<-state.Timer.C
	fmt.Println("Timer fired.")
	state.TimerSatisified = true
	if state.isDone() {
		finishEvent(*state, *event)
	}
}

func updateFloors(state *eventState, idx int, event *Event, ch_FloorSensor chan int,  ch_Done chan int) {
	for {
		fmt.Println("Polling floors")
		select {
		case newFloor := <-ch_FloorSensor:
			state.Elevators[idx].Floor =  uint8(newFloor)
			fmt.Printf("Updated floor: %d %d", idx, state.Elevators[idx].Floor)
			if event != nil {
				for i := range event.ExpectedFloors { 
					state.FloorsSatisfied = true
					if event.ExpectedFloors[i] != 255 && event.ExpectedFloors[i] != state.Elevators[i].Floor {
						state.FloorsSatisfied = false
					}
				}
				if state.isDone() {
					finishEvent(*state, *event)
				}
			}
		}
	}
}


func finishEvent(state eventState, event Event) {
	fmt.Println("Done!")
}


func EventListener(simulatedElevators []fatelevator.SimulatedElevator,  event Event, ch_Done chan int) {
	var state eventState
	state.Elevators = make([]ElevatorState, len(simulatedElevators))

	for i := range simulatedElevators {
		state.Elevators[i].Floor = simulatedElevators[i].CurrentFloor	
	}

	//Set up timer. 
	//If we want to enable it, we start it. Else, we initialize it, but simply stop it.
	state.TimerSatisified = true

	if event.WaitFor > 0 {
		state.TimerSatisified = false
		state.Timer = *time.NewTimer(event.WaitFor)
		fmt.Println("Event configured with timer ", event.WaitFor)
	} else {
		fmt.Println("Event timer not configured.")
	}


	state.FloorsSatisfied = true
	for idx := range event.ExpectedFloors {
		if event.ExpectedFloors[idx] != 255 {
			state.FloorsSatisfied = false
		}	
	}


	if state.isDone() {
		finishEvent(state, event)
		<-ch_Done
	}

	go updateTimer(&state, &event, ch_Done)

	for i := range simulatedElevators {
		go updateFloors(&state, i, &event, simulatedElevators[i].Chan_FloorSensor, ch_Done)
	} 

}
