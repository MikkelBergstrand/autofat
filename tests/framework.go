package tests

import (
	"autofat/fatelevator"
)

type Test struct {
	Id            string
	InitialParams []fatelevator.InitializationParams
	Func          func() error
	Result        bool
}

func CreateTest(id string, testFunc func() error, initParams []fatelevator.InitializationParams) Test {
	return Test{
		Id:            id,
		Func:          testFunc,
		Result:        false,
		InitialParams: initParams,
	}
}

func (test *Test) Run() bool {
	err := test.Func()
	return err == nil
}

func (test Test) NumElevators() int {
	return len(test.InitialParams)
}
