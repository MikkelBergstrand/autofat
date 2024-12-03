package tests

import (
	"autofat/network"
	"autofat/simulator"
)

type Test struct {
	Id            string
	InitialParams []simulator.InitializationParams
	Func          func() error
	Result        bool
	PacketLoss    int
}

func CreateTest(id string, testFunc func() error, initParams []simulator.InitializationParams, packetLoss int) Test {
	return Test{
		Id:            id,
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
	network.SetPacketLoss(test.PacketLoss)
	err := test.Func()
	return err == nil
}

func (test Test) NumElevators() int {
	return len(test.InitialParams)
}
