package tests

import (
	"autofat/dsl"
	"autofat/simulator"
	"log"
)

type Test struct {
	Name          string
	InitialParams []simulator.InitializationParams
	Func          func() error
	Result        bool
	PacketLoss    int
}

func CreateTest(id string, testFunc func() error, initParams []simulator.InitializationParams, packetLoss int) Test {
	return Test{
		Name:          id,
		Func:          testFunc,
		Result:        false,
		InitialParams: initParams,
		PacketLoss:    packetLoss,
	}
}

func CreateSingleElevatorTest(id string, testFunc func() error) Test {
	return CreateTest(id, testFunc, []simulator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}}, 0)
}

func (test *Test) Run() bool {
	err := dsl.Load(test.Name)
	if err != nil {
		log.Fatal(err)
	}
	return true
}

func (test Test) NumElevators() int {
	return len(test.InitialParams)
}
