package tests

import (
	"autofat/events"
	"autofat/fatelevator"
)

type TestParams struct {
	InitialFloors []int //The size of the array denotes the number of active elevators to use.
}

type Test struct {
	Func func()
	ChanResult chan bool
	Result bool
}

func CreateTest(testFunc func()) Test {
	return Test{
		Func: testFunc,
		ChanResult: make(chan bool),
		Result: false,
	}
}

func (test *Test) Run(simulators []fatelevator.SimulatedElevator) bool {
	go func ()  {
		go events.EventListener(simulators, test.ChanResult)
		test.Func()
		test.ChanResult <- true
	}()

	return <-test.ChanResult
}
