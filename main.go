package main

import (
	"autofat/config"
	"autofat/dsl/color"
	"autofat/logger"
	"autofat/network"
	"autofat/procmanager"
	"autofat/simulator"
	"autofat/statemanager"
	"autofat/studentprogram"
	"autofat/tests"
	"autofat/tmux"
	"os"
	"os/signal"
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
	network.InitNamespaceConfig(cfg)
	logger.Init(cfg)

	procmanager.Init()
	initInterruptHandler()

	statemanager.Init()

	network.Init(cfg.StudentProgramDir, cfg)

	test := tests.CreateTest(cfg.TestFile, func() error {
		//Function that does nothing, just sleeps forever.
		select {}
	}, []simulator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}}, 0)
	runTest(&test)
}

func runTest(test *tests.Test) {
	tmux.Launch()

	eval := test.Run(cfg)
	if eval {
		color.Println(color.Green, test.Name, "succeeded.")
	} else {
		color.Println(color.Red, test.Name, "failed.")
	}

	simulator.TerminateAll()
	statemanager.Kill()
	studentprogram.KillAll()
}
