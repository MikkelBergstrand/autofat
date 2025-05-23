package dsl

import (
	"autofat/dsl/runtime"
	"autofat/dsl/storage"
	"autofat/dsl/variables"
	"autofat/elevio"
	"log"
)

type params map[string]variables.Symbol

func defineFunction(rt *runtime.Runtime, compiler *storage.Compiler, fn_name string, fn_def variables.TypeDefinition,
	instructions func(compiler *storage.Compiler, p params)) {
	compiler.NewFunction(fn_name, fn_def)
	sym_params := make(params)

	for _, arg := range fn_def.ArgumentList {
		sym, err := compiler.GetNamedSymbol(arg.Identifier)
		if err != nil {
			log.Fatal(err)
		}
		sym_params[arg.Identifier] = sym
	}

	instructions(compiler, sym_params)
	compiler.DestroyFunctionScope(rt)
}

func defineGlobalVar(compiler *storage.Compiler, name string, _type variables.Type, value any) {
	sym, err := compiler.NewVariable(variables.GetBaseTypeDef(_type), name)
	if err != nil {
		log.Fatalln(err)
	}
	compiler.LoadInstruction(&runtime.InstrLoadImmediate{
		Dest:  *sym,
		Value: value,
	})
}

func echo(storage *storage.Compiler, params params) {
	storage.LoadInstruction(&runtime.InstructionEcho{
		A: params["i"],
	})
	storage.LoadInstruction(&runtime.InstrExitFunction{})
}

func array_len(storage *storage.Compiler, params params) {
	ret_val := storage.NewLiteral(variables.GetBaseTypeDef(variables.INT))
	storage.LoadInstruction(&runtime.InstrArrayLen{
		A:      params["array"],
		Result: ret_val,
	})
	storage.LoadInstruction(&runtime.InstrExitFunction{RetVal: ret_val})
}

func await(storage *storage.Compiler, params params) {
	//Create boolean value to hold return value of the await.
	//Return value of await can be 3 values: OK (statefunc == true) ,NOTOK (statefunc == false) or TIMEOUT
	cond_func_ret_val := storage.NewLiteral(variables.TypeDefinition{BaseType: variables.INT})
	state_func := params["state_func"]
	timeout := params["timeout"]

	chan_sym := storage.NewLiteral(variables.TypeDefinition{BaseType: variables.CHAN})
	timeout_sym := storage.NewLiteral(variables.GetBaseTypeDef(variables.BOOL))
	storage.LoadInstruction(&runtime.InstrLoadImmediate{Dest: cond_func_ret_val, Value: false})
	storage.LoadInstruction(&runtime.InstrLoadImmediate{Dest: timeout_sym, Value: false})
	storage.LoadInstruction(&runtime.InstrLoadImmediate{Dest: chan_sym, Value: nil})
	label := storage.NewAutoLabel()
	storage.LoadLabeledInstruction(&runtime.InstrAwaitStateListen{
		AwaitVal:       chan_sym,
		TimeoutSeconds: timeout,
	}, label)
	storage.LoadInstruction(&runtime.InstrAwait{
		AwaitVal:           chan_sym,
		StateFunction:      state_func,
		ConditionFuncValue: cond_func_ret_val,
		Timeout:            timeout_sym,
	})
	storage.LoadInstruction(&runtime.InstrEndAwait{
		Label:              label,
		ConditionFuncValue: cond_func_ret_val,
		Timeout:            timeout_sym,
		AwaitVal:           chan_sym,
	})
	storage.LoadInstruction(&runtime.InstrExitFunction{RetVal: cond_func_ret_val})
}

func floor(compiler *storage.Compiler, params params) {
	elevs := params["i"]
	result := compiler.NewLiteral(variables.GetBaseTypeDef(variables.INT))

	compiler.LoadInstruction(&runtime.InstrGetFloor{
		ArraySymbol: elevs,
		Result:      result,
	})
	compiler.LoadInstruction(&runtime.InstrExitFunction{RetVal: result})
}

func floorlight(compiler *storage.Compiler, params params) {
	elevs := params["i"]
	result := compiler.NewLiteral(variables.GetBaseTypeDef(variables.INT))

	compiler.LoadInstruction(&runtime.InstrGetFloorLight{
		ArraySymbol: elevs,
		Result:      result,
	})
	compiler.LoadInstruction(&runtime.InstrExitFunction{RetVal: result})
}

func door(compiler *storage.Compiler, params params) {
	elevs := params["elevators"]
	result := compiler.NewLiteral(variables.GetBaseTypeDef(variables.INT))

	compiler.LoadInstruction(&runtime.InstrGetDoorStatus{
		ArraySymbol: elevs,
		Result:      result,
	})
	compiler.LoadInstruction(&runtime.InstrExitFunction{RetVal: result})
}

func moving(compiler *storage.Compiler, params params) {
	elevs := params["elevators"]
	result := compiler.NewLiteral(variables.GetBaseTypeDef(variables.INT))

	compiler.LoadInstruction(&runtime.InstrGetMovementStatus{
		ArraySymbol: elevs,
		Result:      result,
	})
	compiler.LoadInstruction(&runtime.InstrExitFunction{RetVal: result})
}

func exit(compiler *storage.Compiler, params params) {
	retval := params["value"]
	compiler.LoadInstruction(&runtime.InstrExit{Value: retval})
}

func statuslight(compiler *storage.Compiler, params params) {

	elevs := params["i"]
	floor := params["floor"]
	ordertype := params["ordertype"]

	result := compiler.NewLiteral(variables.GetBaseTypeDef(variables.BOOL))

	compiler.LoadInstruction(&runtime.InstrGetStatusLight{
		ArraySymbol: elevs,
		Result:      result,
		Floor:       floor,
		OrderType:   ordertype,
	})
	compiler.LoadInstruction(&runtime.InstrExitFunction{RetVal: result})
}

func sleep(compiler *storage.Compiler, params params) {
	millisec := params["milliseconds"]
	compiler.LoadInstruction(&runtime.InstrSleep{Duration: millisec})
	compiler.LoadInstruction(&runtime.InstrExitFunction{})
}

func sync(compiler *storage.Compiler, params params) {
	threads := params["threads"]
	compiler.LoadInstruction(&runtime.InstrSync{Threads: threads})
	compiler.LoadInstruction(&runtime.InstrExitFunction{})
}

func assert(compiler *storage.Compiler, params params) {
	state_func := params["state_func"]
	deadzone_milliseconds := params["deadzone_milliseconds"]

	assert_data := compiler.NewLiteral(variables.TypeDefinition{BaseType: variables.CHAN})
	compiler.LoadInstruction(&runtime.InstrLoadImmediate{Dest: assert_data, Value: nil})

	deadzone_violated := compiler.NewLiteral(variables.GetBaseTypeDef(variables.BOOL))
	compiler.LoadInstruction(&runtime.InstrLoadImmediate{Dest: deadzone_violated, Value: false})

	cond_func_ret_val := compiler.NewLiteral(variables.TypeDefinition{BaseType: variables.BOOL})
	compiler.LoadInstruction(&runtime.InstrLoadImmediate{Dest: cond_func_ret_val, Value: false})

	label := compiler.NewAutoLabel()
	compiler.LoadLabeledInstruction(&runtime.InstrAssertStateListen{
		AssertVal: assert_data,
	}, label)
	compiler.LoadInstruction(&runtime.InstrAssert{
		AssertVal:          assert_data,
		StateFunction:      state_func,
		ConditionFuncValue: cond_func_ret_val,
		DeadzoneViolated:   deadzone_violated,
	})
	compiler.LoadInstruction(&runtime.InstrEndAssert{
		Label:                label,
		ConditionFuncValue:   cond_func_ret_val,
		AssertVal:            assert_data,
		DeadzoneMilliseconds: deadzone_milliseconds,
		Deadzoneviolated:     deadzone_violated,
	})
}

func my_append(compiler *storage.Compiler, params params) {
	array := params["array"]
	value := params["value"]

	compiler.LoadInstruction(&runtime.InstrArrayAppend{
		Array: array,
		Value: value,
	})
	compiler.LoadInstruction(&runtime.InstrExitFunction{RetVal: array})
}

func make_order(compiler *storage.Compiler, params params) {
	compiler.LoadInstruction(&runtime.InstrMakeOrder{
		OrderType: params["ordertype"],
		Floor:     params["floor"],
		Elevator:  params["elevator"],
	})
	compiler.LoadInstruction(&runtime.InstrExitFunction{})
}
func generateGlobalVariables(compiler *storage.Compiler) {
	defineGlobalVar(compiler, "CAB", variables.ORDERTYPE, elevio.BT_Cab)
	defineGlobalVar(compiler, "HALLUP", variables.ORDERTYPE, elevio.BT_HallUp)
	defineGlobalVar(compiler, "HALLDOWN", variables.ORDERTYPE, elevio.BT_HallDown)
}
func generateGlobalFunctions(rt *runtime.Runtime, storage *storage.Compiler) {
	// Create function echo
	defineFunction(rt, storage, "echo", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.TypeDefinition{
					BaseType: variables.ANY,
				},
				Identifier: "i",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.NONE},
	}, echo)

	//Create function await
	defineFunction(rt, storage, "await", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.TypeDefinition{
					BaseType:   variables.FUNC,
					ReturnType: &variables.TypeDefinition{BaseType: variables.BOOL},
				},
				Identifier: "state_func",
			},
			{
				Definition: variables.GetBaseTypeDef(variables.INT),
				Identifier: "timeout",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.BOOL},
	}, await)

	//Create function to check floor
	defineFunction(rt, storage, "floor", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.TypeDefinition{BaseType: variables.INT, IsArray: true},
				Identifier: "i",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.INT},
	}, floor)

	//Create function to check floor light
	defineFunction(rt, storage, "floorlight", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.TypeDefinition{BaseType: variables.INT, IsArray: true},
				Identifier: "i",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.INT},
	}, floorlight)

	// Create function to check status light
	defineFunction(rt, storage, "statuslight", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.TypeDefinition{BaseType: variables.INT, IsArray: true},
				Identifier: "i",
			},
			{
				Definition: variables.GetBaseTypeDef(variables.ORDERTYPE),
				Identifier: "ordertype",
			},
			{
				Definition: variables.GetBaseTypeDef(variables.INT),
				Identifier: "floor",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.INT},
	}, statuslight)

	defineFunction(rt, storage, "door", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.TypeDefinition{BaseType: variables.INT, IsArray: true},
				Identifier: "elevators",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.INT},
	}, door)

	defineFunction(rt, storage, "moving", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.TypeDefinition{BaseType: variables.INT, IsArray: true},
				Identifier: "elevators",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.INT},
	}, door)

	defineFunction(rt, storage, "exit", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.TypeDefinition{
					BaseType: variables.BOOL,
				},
				Identifier: "value",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.NONE},
	}, exit)

	defineFunction(rt, storage, "sleep", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.TypeDefinition{
					BaseType: variables.INT,
				},
				Identifier: "millisec",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.NONE},
	}, sleep)

	defineFunction(rt, storage, "len", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.TypeDefinition{
					BaseType: variables.ANY,
					IsArray:  true,
				},
				Identifier: "array",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.INT},
	}, array_len)

	defineFunction(rt, storage, "sync", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.TypeDefinition{
					BaseType: variables.THREAD,
					IsArray:  true,
				},
				Identifier: "threads",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.NONE},
	}, sync)

	defineFunction(rt, storage, "assert", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.TypeDefinition{
					BaseType:   variables.FUNC,
					ReturnType: &variables.TypeDefinition{BaseType: variables.BOOL},
				},
				Identifier: "state_func",
			},
			{
				Definition: variables.TypeDefinition{
					BaseType: variables.INT,
				},
				Identifier: "deadzone_milliseconds",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.NONE},
	}, assert)

	defineFunction(rt, storage, "make_order", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.GetBaseTypeDef(variables.INT),
				Identifier: "elevator",
			},
			{

				Definition: variables.GetBaseTypeDef(variables.ORDERTYPE),
				Identifier: "ordertype",
			},
			{

				Definition: variables.GetBaseTypeDef(variables.INT),
				Identifier: "floor",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.NONE},
	}, make_order)

	defineFunction(rt, storage, "append", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Identifier: "array",
				Definition: variables.TypeDefinition{
					BaseType: variables.ANY,
					IsArray:  true,
				},
			},
			{
				Identifier: "value",
				Definition: variables.TypeDefinition{
					BaseType: variables.ANY,
				},
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.ANY, IsArray: true},
	}, my_append)
}
