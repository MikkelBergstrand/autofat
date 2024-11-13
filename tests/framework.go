package tests

import (
	"autofat/events"
	"autofat/fatelevator"
)

type Test struct {
	InitialParams []fatelevator.InitializationParams
	Func          func()
	ChanResult    chan bool
	Result        bool
}

func CreateTest(testFunc func(), initParams []fatelevator.InitializationParams) Test {
	return Test{
		Func:          testFunc,
		ChanResult:    make(chan bool),
		Result:        false,
		InitialParams: initParams,
	}
}

func (test *Test) Run(simulators []fatelevator.SimulatedElevator) bool {
	go func() {
		events.EventListener(simulators, test.ChanResult)
		test.Func()
		test.ChanResult <- true
	}()

	return <-test.ChanResult
}

func (test Test) NumElevators() int {
	return len(test.InitialParams)
}
