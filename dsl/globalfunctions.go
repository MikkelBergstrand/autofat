package dsl

import (
	"autofat/dsl/runtime"
	"autofat/dsl/storage"
	"autofat/dsl/variables"
	"log"
)

func defineFunction(rt *runtime.Runtime, compiler *storage.Compiler, fn_name string, fn_def variables.TypeDefinition, instructions func(compiler *storage.Compiler)) {
	compiler.NewFunction(fn_name, fn_def)
	instructions(compiler)
	compiler.DestroyFunctionScope(rt)
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
	ret_val := storage.NewLiteral(variables.TypeDefinition{BaseType: variables.BOOL})
	state_func, err := storage.GetNamedSymbol("state_func")
	if err != nil {
		log.Fatal(err)
	}
	chan_sym := storage.NewLiteral(variables.TypeDefinition{BaseType: variables.CHAN})
	storage.LoadInstruction(&runtime.InstrLoadImmediate{Dest: ret_val, Value: false})
	storage.LoadInstruction(&runtime.InstrLoadImmediate{Dest: chan_sym, Value: nil})
	label := storage.NewAutoLabel()
	storage.LoadLabeledInstruction(&runtime.InstrStateListen{Symbol: chan_sym}, label)
	storage.LoadInstruction(&runtime.InstrAwait{
		Channel: chan_sym,
		StateFunction: state_func,
		RetVal:        ret_val,
	})
	storage.LoadInstruction(&runtime.InstrEndAwait{
		Label:      label,
		AwaitValue: ret_val,
	})
	storage.LoadInstruction(&runtime.InstrExitFunction{})
}

func floor(compiler *storage.Compiler) {
	elevs, err := compiler.GetNamedSymbol("i")
	if err != nil {
		log.Fatal(err)
	}
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
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.NONE},
	}, await)

	//Create function to check floor
	defineFunction(rt, storage, "floor", variables.TypeDefinition{
		BaseType: variables.FUNC,
		ArgumentList: []variables.Argument{
			{
				Definition: variables.TypeDefinition{ BaseType: variables.ARRAY },
				Identifier: "i",
			},
		},
		ReturnType: &variables.TypeDefinition{BaseType: variables.INT},
	}, floor)
}
