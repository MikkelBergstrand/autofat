package main

import (
	"autofat/config"
	"autofat/network"
	"autofat/procmanager"
	"autofat/simulator"
	"autofat/statemanager"
	"autofat/studentprogram"
	"autofat/tests"
	"autofat/tmux"
	"fmt"
	"os"
	"os/signal"
	"time"
)

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
		for range c {
			procmanager.KillAll()
			os.Exit(1)
		}
	}()
}

var cfg config.Config

func main() {
	cfg = config.LoadFromFlags()

	simulator.SetExecutablePath(cfg.SimElevatorServerPath)
	network.InitNamespaceConfig(cfg.NetworkNamespaces[:])

	procmanager.Init()
	initInterruptHandler()

	statemanager.Init()

	network.Init(cfg.StudentProgramDir, cfg)

	test_cab_backup := tests.CreateTest("cab_backup", tests.TestCabBackup, []simulator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}, {
		InitialFloor:  1,
		BetweenFloors: false,
	}}, 0)

	test := tests.CreateTest("floor_lamp", tests.TestFloorLamp, []simulator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}}, 0)
	test2 := tests.CreateTest("init_between_floors", tests.TestInitBetweenFloors, []simulator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: true,
	}}, 0)
	engine_fail_test := tests.CreateTest("engine_failure", tests.TestEngineOutage, []simulator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}, {
		InitialFloor:  0,
		BetweenFloors: false,
	}}, 0)
	hall_clear_one_test := tests.CreateTest("hall_clear_one", tests.TestHallClearOne, []simulator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}}, 0)
	door_timer_test := tests.CreateTest("door_timer", tests.TestDoorOpenTime, []simulator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}}, 0)

	obstruction_open_door_test := tests.CreateTest("obstruction_opens_door", tests.TestObstructionOpenDoor, []simulator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}}, 0)

	obstruction_buffer_order_test := tests.CreateSingleElevatorTest("obstruction_buffer_orders", tests.TestObstructionCompleteOrders)

	empty_test := tests.CreateTest("empty_test", func() error {
		//Function that does nothing, just sleeps forever.
		select {}
	}, []simulator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}, {
		InitialFloor:  1,
		BetweenFloors: false,
	}, {
		InitialFloor:  2,
		BetweenFloors: false,
	}}, 0)

	if !cfg.NoTests {
		runTest(&test_cab_backup)
		runTest(&obstruction_buffer_order_test)
		runTest(&obstruction_open_door_test)
		runTest(&door_timer_test)
		runTest(&hall_clear_one_test)
		runTest(&test2)
		runTest(&engine_fail_test)
		runTest(&test)
	} else {
		runTest(&empty_test)
	}
}

func runTest(test *tests.Test) {
	fmt.Println("Beginning test", test.Id)
	tmux.Launch()

	for i := 0; i < test.NumElevators(); i++ {
		simulator.Init(cfg.GetElevatorConfig(i), test.InitialParams[i])
		simulator.Run(i)
	}

	time.Sleep(500 * time.Millisecond)
	studentprogram.InitalizeFromConfig(cfg.StudentProgramWaitTime, cfg.StudentProgramDir, cfg.GetAllElevatorConfigs(), test.NumElevators())
	time.Sleep(1000 * time.Millisecond)

	statemanager.EventListener(test.Id)

	eval := test.Run()
	fmt.Printf("Value of test %s was %t\n", test.Id, eval)

	simulator.TerminateAll()
	statemanager.Kill()
	studentprogram.KillAll()
}
