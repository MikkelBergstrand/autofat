package tests

import (
	"autofat/config"
	"autofat/dsl"
	"autofat/simulator"
	"log"
)

type Test struct {
	Name       string
	Func       func() error
	Result     bool
	PacketLoss int
}

func CreateTest(id string, testFunc func() error, initParams []simulator.InitializationParams, packetLoss int) Test {
	return Test{
		Name:       id,
		Func:       testFunc,
		Result:     false,
		PacketLoss: packetLoss,
	}
}

func CreateSingleElevatorTest(id string, testFunc func() error) Test {
	return CreateTest(id, testFunc, []simulator.InitializationParams{{
		InitialFloor:  0,
		BetweenFloors: false,
	}}, 0)
}

func (test *Test) Run(config config.Config) bool {
	val, err := dsl.Load(test.Name, config)
	if err != nil {
		log.Fatal(err)
	}
	return val
}
