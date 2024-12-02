package main

import (
	"autofat/config"
	"autofat/events"
	"autofat/fatelevator"
	"autofat/network"
	"autofat/procmanager"
	"autofat/studentprogram"
	"autofat/tests"
	"autofat/tmux"
	"fmt"
	"net/netip"
	"os"
	"os/signal"
	"time"
)

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

var _elevatorConfigs []config.ElevatorConfig

func main() {
	procmanager.Init()
	initInterruptHandler()

	events.Init()

	for i := 0; i < 3; i++ {
		_elevatorConfigs = append(_elevatorConfigs, config.ElevatorConfig{
			UserAddrPort: netip.AddrPortFrom(netip.AddrFrom4(LOCALHOST), USERPROGRAM_PORTS[i]),
			FatAddrPort:  netip.AddrPortFrom(netip.AddrFrom4(LOCALHOST), FATPROGRAM_PORTS[i]),
		})
	}

	network.Init(LAUNCH_PROGRAM_DIR, _elevatorConfigs)

	test_cab_backup := tests.CreateTest("cab_backup", tests.TestCabBackup, []fatelevator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}, {
		InitialFloor:  1,
		BetweenFloors: false,
	}}, 0)

	test := tests.CreateTest("floor_lamp", tests.TestFloorLamp, []fatelevator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}}, 0)
	test2 := tests.CreateTest("init_between_floors", tests.TestInitBetweenFloors, []fatelevator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: true,
	}}, 0)
	engine_fail_test := tests.CreateTest("engine_failure", tests.TestEngineOutage, []fatelevator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}, {
		InitialFloor:  0,
		BetweenFloors: false,
	}}, 0)
	hall_clear_one_test := tests.CreateTest("hall_clear_one", tests.TestHallClearOne, []fatelevator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}}, 0)
	door_timer_test := tests.CreateTest("door_timer", tests.TestDoorOpenTime, []fatelevator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}}, 0)

	runTest(&door_timer_test)
	runTest(&hall_clear_one_test)
	runTest(&test2)
	runTest(&engine_fail_test)
	runTest(&test)
	runTest(&test_cab_backup)

}

func runTest(test *tests.Test) {
	tmux.Launch()

	for i := 0; i < test.NumElevators(); i++ {
		fatelevator.Init(_elevatorConfigs[i], test.InitialParams[i])
		fatelevator.Run(i)
	}

	time.Sleep(500 * time.Millisecond)
	studentprogram.InitalizeFromConfig(LAUNCH_PROGRAM_DIR, _elevatorConfigs, test.NumElevators())
	time.Sleep(1000 * time.Millisecond)

	events.EventListener(test.Id)

	eval := test.Run()
	fmt.Println("Value of test was", eval)

	fatelevator.TerminateAll()
	events.Kill()
	studentprogram.KillAll()
}
