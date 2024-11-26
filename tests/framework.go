package tests

import (
	"autofat/fatelevator"
)

type Test struct {
	Id            string
	InitialParams []fatelevator.InitializationParams
	Func          func(chan bool)
	ChanResult    chan bool
	Result        bool
}

func CreateTest(id string, testFunc func(chan bool), initParams []fatelevator.InitializationParams) Test {
	return Test{
		Id:            id,
		Func:          testFunc,
		ChanResult:    make(chan bool),
		Result:        false,
		InitialParams: initParams,
	}
}

func (test *Test) Run(simulators []fatelevator.SimulatedElevator) bool {
	go func() {
		test.Func(test.ChanResult)
		test.ChanResult <- true
	}()

	return <-test.ChanResult
}

func (test Test) NumElevators() int {
	return len(test.InitialParams)
}
