package dsl

import (
	"autofat/dsl/runtime"
	"autofat/dsl/storage"
	"autofat/dsl/variables"
	"autofat/elevio"
	"log"
)

func defineFunction(rt *runtime.Runtime, compiler *storage.Compiler, fn_name string, fn_def variables.TypeDefinition,
	instructions func(compiler *storage.Compiler)) {
	compiler.NewFunction(fn_name, fn_def)
	instructions(compiler)
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

func echo(storage *storage.Compiler) {
	A, err := storage.GetNamedSymbol("i")
	if err != nil {
		log.Fatal(err)
	}
	storage.LoadInstruction(&runtime.InstructionEcho{
		A: A,
	})
	storage.LoadInstruction(&runtime.InstrExitFunction{})
}

func await(storage *storage.Compiler) {
	//Create boolean value to hold return value of the await.
	//Return value of await can be 3 values: OK (statefunc == true) ,NOTOK (statefunc == false) or TIMEOUT
	cond_func_ret_val := storage.NewLiteral(variables.TypeDefinition{BaseType: variables.INT})
	state_func, err := storage.GetNamedSymbol("state_func")
	if err != nil {
		log.Fatal(err)
	}
	timeout, err := storage.GetNamedSymbol("timeout")
	if err != nil {
		log.Fatal(err)
	}

	chan_sym := storage.NewLiteral(variables.TypeDefinition{BaseType: variables.CHAN})
	timeout_sym := storage.NewLiteral(variables.GetBaseTypeDef(variables.BOOL))
	storage.LoadInstruction(&runtime.InstrLoadImmediate{Dest: cond_func_ret_val, Value: false})
	storage.LoadInstruction(&runtime.InstrLoadImmediate{Dest: timeout_sym, Value: false})
	storage.LoadInstruction(&runtime.InstrLoadImmediate{Dest: chan_sym, Value: nil})
	label := storage.NewAutoLabel()
	storage.LoadLabeledInstruction(&runtime.InstrStateListen{
		Symbol:         chan_sym,
		TimeoutSeconds: timeout,
	}, label)
	storage.LoadInstruction(&runtime.InstrAwait{
		Channel:            chan_sym,
		StateFunction:      state_func,
		ConditionFuncValue: cond_func_ret_val,
		Timeout:            timeout_sym,
	})
	storage.LoadInstruction(&runtime.InstrEndAwait{
		Label:              label,
		ConditionFuncValue: cond_func_ret_val,
		Timeout:            timeout_sym,
	})
	storage.LoadInstruction(&runtime.InstrExitFunction{})
}

func floor(compiler *storage.Compiler) {
	elevs, err := compiler.GetNamedSymbol("i")
	if err != nil {
		log.Fatal(err)
	}
	result := compiler.NewLiteral(variables.GetBaseTypeDef(variables.INT))

	compiler.LoadInstruction(&runtime.InstrGetFloor{
		ArraySymbol: elevs,
		Result:      result,
	})
	compiler.LoadInstruction(&runtime.InstrExitFunction{RetVal: result})
}

func statuslight(compiler *storage.Compiler) {
	elevs, err := compiler.GetNamedSymbol("i")
	if err != nil {
		log.Fatal(err)
	}
	floor, _ := compiler.GetNamedSymbol("floor")
	ordertype, _ := compiler.GetNamedSymbol("ordertype")

	result := compiler.NewLiteral(variables.GetBaseTypeDef(variables.BOOL))

	compiler.LoadInstruction(&runtime.InstrGetStatusLight{
		ArraySymbol: elevs,
		Result:      result,
		Floor:       floor,
		OrderType:   ordertype,
	})
	compiler.LoadInstruction(&runtime.InstrExitFunction{RetVal: result})
}

func generateGlobalVariables(rt *runtime.Runtime, compiler *storage.Compiler) {
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
		ReturnType: &variables.TypeDefinition{BaseType: variables.NONE},
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
}
