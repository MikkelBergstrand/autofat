package main

import (
	"autofat/config"
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

var USERPROGRAM_PORTS = [3]uint16{12345, 12346, 12347}
var FATPROGRAM_PORTS = [3]uint16{12348, 12349, 12350}
var LOCALHOST = [4]byte{127, 0, 0, 1}

var _simulatedElevators []fatelevator.SimulatedElevator
var _elevatorConfigs []config.ElevatorConfig

func main() {
	procmanager.Init()
	initInterruptHandler()

	tmux.Cleanup()
	tmux.Launch()

	for i := 0; i < 3; i++ {
		_elevatorConfigs = append(_elevatorConfigs, config.ElevatorConfig{
			UserAddrPort: netip.AddrPortFrom(netip.AddrFrom4(LOCALHOST), USERPROGRAM_PORTS[i]),
			FatAddrPort:  netip.AddrPortFrom(netip.AddrFrom4(LOCALHOST), FATPROGRAM_PORTS[i]),
		})
	}

	for i := 0; i < 3; i++ {
		var simulatedElevator fatelevator.SimulatedElevator
		simulatedElevator.Init(_elevatorConfigs[i])
		_simulatedElevators = append(_simulatedElevators, simulatedElevator)
		_simulatedElevators[i].Run(i + 1)
	}

	test := tests.CreateTest(tests.TestFloorLamp, []fatelevator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}})
	test2 := tests.CreateTest(tests.TestInitBetweenFloors, []fatelevator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: true,
	}})
	runTest(&test)
	runTest(&test2)

}

func runTest(test *tests.Test) {

	//Create context, to supply to shell commands.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for i := 0; i < test.NumElevators(); i++ {
		_simulatedElevators[i].Reload(test.InitialParams[i])
	}

	time.Sleep(500 * time.Millisecond)
	studentprogram.InitalizeFromConfig(ctx, LAUNCH_PROGRAM_DIR, _elevatorConfigs, test.NumElevators())
	time.Sleep(1000 * time.Millisecond)

	eval := test.Run(_simulatedElevators)
	fmt.Println("Value of test was", eval)

	procmanager.KillAll()
}
