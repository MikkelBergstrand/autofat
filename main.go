package main

import (
	"autofat/config"
	"autofat/elevio"
	"autofat/fatelevator"
	"autofat/procmanager"
	"autofat/studentprogram"
	"autofat/tests"
	"autofat/tmux"
	"context"
	"fmt"
	"net/netip"
	"os"
	"os/signal"
	"time"
)

// var LAUNCH_PROGRAM_DIR = "../sanntidsprosjekt/TT"
var LAUNCH_PROGRAM_DIR = "../sanntidsprosjekt"

type ElevatorInstance struct {
	Chan_KillProcess chan int
	Chan_ProcessDone chan int
	InitialFloor     int
	CurrentFloor     int
}

// Cleanup for interrupts, such as when the program is CTRL+C-ed
func initInterruptHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			fmt.Println("Catched signal", sig.String())
			procmanager.KillAll()
			os.Exit(1)
		}
	}()
}

func main() {
	USERPROGRAM_PORTS := [3]uint16{12345, 12346, 12347}
	FATPROGRAM_PORTS := [3]uint16{12348, 12349, 12350}
	LOCALHOST := [4]byte{127, 0, 0, 1}
	N_ELEVATORS := 1

	procmanager.Init()

	initInterruptHandler()

	//Display
	tmux.Cleanup()
	tmux.Launch()

	//Create context, to supply to shell commands.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

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

		fatelevator.RunSimulator(&elevios[i], simulatedElevators[i], i+1)
	}

	time.Sleep(500 * time.Millisecond)

	studentprogram.InitalizeFromConfig(ctx, LAUNCH_PROGRAM_DIR, elevators, N_ELEVATORS)

	time.Sleep(1000 * time.Millisecond)

	eval := tests.TestFloorLamp(simulatedElevators)
	fmt.Println("Value of test was", eval)

	//FIXME
	time.Sleep(10000000 * time.Second)
}
