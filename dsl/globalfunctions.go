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
	storage.LoadLabeledInstruction(&runtime.InstrStateListen{
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
					BaseType: variables.INT,
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
				Definition: variables.TypeDefinition{BaseType: variables.ARRAY},
				Identifier: "i",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.INT},
	}, floor)

	// Create function to check status light
	defineFunction(rt, storage, "statuslight", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.TypeDefinition{BaseType: variables.ARRAY},
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
}
