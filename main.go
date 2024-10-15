package main

import (
	"autofat/elevio"
	"autofat/events"
	"autofat/fatelevator"
	"autofat/tmux"
	"fmt"
	"net/netip"
	"os/exec"
	"strconv"
	"time"
)

var LAUNCH_PROGRAM_CMD = "go"
var LAUNCH_PROGRAM_DIR = "/home/mikkel/dev/prosjekt/sanntidsprosjekt"

type ElevatorConfig struct {
	UserAddrPort netip.AddrPort
	FatAddrPort  netip.AddrPort
}

type ElevatorInstance struct {
	Chan_KillProcess chan int
	Chan_ProcessDone chan int
	InitialFloor     int
	CurrentFloor     int
}

func userProcess(userPort uint16, id int) {
	cmd := exec.Command(LAUNCH_PROGRAM_CMD, "run", "main.go", ":"+strconv.Itoa(int(userPort)), strconv.Itoa(int(id)))
	cmd.Dir = LAUNCH_PROGRAM_DIR
	fmt.Printf("Launching user process, port=%d\n", userPort)
	if err := cmd.Run(); err != nil {
		fmt.Println(err)
	}
}

func main() {
	N_ELEVATORS := 1
	USERPROGRAM_PORTS := [3]uint16{12345, 12346, 12347}
	FATPROGRAM_PORTS := [3]uint16{12348, 12349, 12350}

	LOCALHOST := [4]byte{127, 0, 0, 1}
	var elevators []ElevatorConfig

	// Create tmux environment, for display
	tmux.Cleanup()
	tmux.Launch()

	var simulatedElevators []fatelevator.SimulatedElevator
	var elevios []elevio.ElevIO

	for i := 0; i < N_ELEVATORS; i++ {
		var simulatedElevator fatelevator.SimulatedElevator
		simulatedElevator.Init(
			USERPROGRAM_PORTS[i],
			FATPROGRAM_PORTS[i],
			tmux.GetTTYFromPane(i+1),
			0,
		)

		simulatedElevators = append(simulatedElevators, simulatedElevator)
		elevios = append(elevios, elevio.ElevIO{})

		fatelevator.RunSimulator(&elevios[i], simulatedElevators[i])
	}

	time.Sleep(500 * time.Millisecond)

	for i := 0; i < N_ELEVATORS; i++ {
		elevators = append(elevators, ElevatorConfig{
			UserAddrPort: netip.AddrPortFrom(netip.AddrFrom4(LOCALHOST), USERPROGRAM_PORTS[i]),
			FatAddrPort:  netip.AddrPortFrom(netip.AddrFrom4(LOCALHOST), FATPROGRAM_PORTS[i]),
		})
		go userProcess(elevators[i].UserAddrPort.Port(), i)

	}

	time.Sleep(1000 * time.Millisecond)

	var eventList = []events.Event{
		events.Event{
			ID:          "init",
			Description: "Initial event",
			TriggerType: events.TRIGGER_INIT,
			ActionType:  events.ACTION_MAKE_ORDER,
			ActionParams: events.Button{
				Elevator: 0,
				ButtonEvent: elevio.ButtonEvent{
					Button: elevio.BT_Cab,
					Floor:  3,
				},
			},
		},
	}
	go events.EventListener(simulatedElevators, eventList)

	//FIXME
	time.Sleep(10000000 * time.Second)
}
