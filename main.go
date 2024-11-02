package main

import (
	"autofat/config"
	"autofat/elevio"
	"autofat/fatelevator"
	"autofat/studentprogram"
	"autofat/tests"
	"autofat/tmux"
	"fmt"
	"net/netip"
	"time"
)

var LAUNCH_PROGRAM_CMD = "go"
var LAUNCH_PROGRAM_DIR = "../sanntidsprosjekt"

type ElevatorInstance struct {
	Chan_KillProcess chan int
	Chan_ProcessDone chan int
	InitialFloor     int
	CurrentFloor     int
}

func main() {
	USERPROGRAM_PORTS := [3]uint16{12345, 12346, 12347}
	FATPROGRAM_PORTS := [3]uint16{12348, 12349, 12350}
	LOCALHOST := [4]byte{127, 0, 0, 1}
	N_ELEVATORS := 3

	//Display
	tmux.Cleanup()
	tmux.Launch()

	var elevators []config.ElevatorConfig
	for i := 0; i < 3; i++ {
		elevators = append(elevators, config.ElevatorConfig{
			UserAddrPort: netip.AddrPortFrom(netip.AddrFrom4(LOCALHOST), USERPROGRAM_PORTS[i]),
			FatAddrPort:  netip.AddrPortFrom(netip.AddrFrom4(LOCALHOST), FATPROGRAM_PORTS[i]),
		})
	}


	var simulatedElevators []fatelevator.SimulatedElevator
	var elevios []elevio.ElevIO

	for i := 0; i < N_ELEVATORS; i++ {
		var simulatedElevator fatelevator.SimulatedElevator
		simulatedElevator.Init(elevators[i])

		simulatedElevators = append(simulatedElevators, simulatedElevator)
		elevios = append(elevios, elevio.ElevIO{})

		fatelevator.RunSimulator(&elevios[i], simulatedElevators[i])
	}

	time.Sleep(500 * time.Millisecond)

	studentprogram.InitalizeFromConfig(LAUNCH_PROGRAM_DIR, elevators, N_ELEVATORS)

	time.Sleep(1000 * time.Millisecond)

	eval := tests.TestFloorLamp(simulatedElevators)
	fmt.Println("Value of test was", eval)

	//FIXME
	time.Sleep(10000000 * time.Second)
}
