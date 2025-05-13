package parser

import (
	"autofat/dsl/runtime"
	"autofat/dsl/storage"
	"autofat/dsl/variables"
	"fmt"
	"log"
	"strconv"
)

type condition_tree_entry struct {
	start_label string
	end         *runtime.InstructionLabelPair // Instruction at the end of a conditonal block. Can be nil, if no block follows it.
	jmp         *runtime.InstrJmpIf           // Instruction that starts the conditional block, of type InstrJmpIf. Can again be nil, for else statement.
}

type List[T any] struct {
	First  T
	Second *List[T]
}

type FunctionCall struct {
	FuncSymbol variables.Symbol
	ArgList    []variables.Symbol
}

func (list List[T]) Iterate() (ret []T) {
	ret = append(ret, list.First)

	node := list.Second
	for node != nil {
		ret = append(ret, node.First)
		node = node.Second
	}
	return ret
}

// Convert string to integer. Should not fail!
func intval(s string) int {
	intval, err := strconv.Atoi(s)
	if err != nil {
		panic(err)
	}
	return intval
}

func integerArithmetic(words []any, storage *storage.Compiler, op runtime.Operator) variables.Symbol {
	new_addr := storage.NewLiteral(variables.TypeDefinition{BaseType: variables.INT})
	storage.LoadInstruction(&runtime.InstrArithmetic{
		A:        words[0].(variables.Symbol),
		B:        words[2].(variables.Symbol),
		Result:   new_addr,
		Operator: op,
	})
	return new_addr
}
func validateBooleanArithmetic(a variables.Symbol, b variables.Symbol, op runtime.BooleanOperator) error {
	if a.Type.BaseType != b.Type.BaseType || !op.IsValidFor(a.Type.BaseType) {
		return fmt.Errorf("invalid type comparison of %s and %s", a.Type, b.Type)
	}
	return nil
}

func booleanArithmetic(words []any, s *storage.Compiler, op runtime.BooleanOperator) variables.Symbol {
	a := words[0].(variables.Symbol)
	b := words[2].(variables.Symbol)

	err := validateBooleanArithmetic(a, b, op)
	if err != nil {
		log.Fatalln(err.Error())
	}

	newaddr := s.NewLiteral(variables.TypeDefinition{BaseType: variables.BOOL})
	if a.Type.BaseType == variables.BOOL {
		s.LoadInstruction(&runtime.InstrCompareBool{
			A:        words[0].(variables.Symbol),
			B:        words[2].(variables.Symbol),
			Result:   newaddr,
			Operator: op,
		})
	} else if a.Type.BaseType == variables.INT {
		s.LoadInstruction(&runtime.InstrCompareInt{
			A:        words[0].(variables.Symbol),
			B:        words[2].(variables.Symbol),
			Result:   newaddr,
			Operator: op,
		})
	}
	return newaddr
}

func doAssignment(src variables.Symbol, dest variables.Symbol, storage *storage.Compiler) variables.Symbol {
	if dest.Type.BaseType == variables.UNDETERMINED {
		dest.Type.BaseType = src.Type.BaseType
	}
	if src.Type.BaseType == variables.UNDETERMINED {
		src.Type.BaseType = dest.Type.BaseType
	}
	if !src.Type.Equals(dest.Type) {
		log.Fatalf("invalid type assignment: expected %s, got %s", src.Type.String(), dest.Type.String())
	}

	storage.LoadInstruction(&runtime.InstrAssign{
		Source: src,
		Dest:   dest,
	})
	return dest
}

func doFunctionCall(name string, arguments []variables.Symbol, storage *storage.Compiler) (variables.Symbol, error) {
	sym, err := storage.GetNamedSymbol(name)
	if err != nil {
		return sym, err
	}
	if sym.Type.BaseType != variables.FUNC {
		return sym, fmt.Errorf("attempting to call %s, a non-function variable", name)
	}

	if !sym.Type.ArgumentList.ValidateArgumentList(arguments) {
		return sym, fmt.Errorf("argument list to function %s invalid", name)
	}

	return sym, nil
}

func DoActions(rule_id int, words []any, storage *storage.Compiler, r *runtime.Runtime) any {
	switch rule_id {
	case 3:
		return integerArithmetic(words, storage, runtime.ADD)
	case 4:
		return integerArithmetic(words, storage, runtime.SUB)
	case 6:
		return integerArithmetic(words, storage, runtime.MULT)
	case 7:
		return integerArithmetic(words, storage, runtime.DIV)
	case 10: //New integer literal
		addr := storage.NewLiteral(variables.TypeDefinition{BaseType: variables.INT})
		storage.LoadInstruction(&runtime.InstrLoadImmediate{
			Dest:  addr,
			Value: intval(words[0].(string)),
		})
		return addr
	case 11:
		sym, err := storage.GetNamedSymbol(words[0].(string))
		if err != nil {
			log.Fatal(err)
		}
		return sym
	case 12: // New variable, eg. int a = 3
		_type := words[0].(variables.TypeDefinition)
		src := words[3].(variables.Symbol)
		addr, err := storage.NewVariable(_type, words[1].(string))
		if err != nil {
			log.Fatal(err)
		}

		return doAssignment(src, *addr, storage)
	case 13: // Reassignment of integer, e.g. a = 3
		addr, err := storage.GetNamedSymbol(words[0].(string))
		if err != nil {
			log.Fatal(err)
		}

		return doAssignment(words[2].(variables.Symbol), addr, storage)
	case 16: // Declare scope
		storage.LoadInstruction(&runtime.InstrBeginScope{})
		storage.NewScope()
	case 17: // End scope
		storage.LoadInstruction(&runtime.InstrEndScope{})
		storage.DestroyScope()
	case 19: // call function e.g. echo ( 0 )
		arg_list := (words[2].(List[variables.Symbol])).Iterate()
		func_name := words[0].(string)
		func_sym, err := doFunctionCall(func_name, arg_list, storage)
		if err != nil {
			log.Fatal(err)
		}
		return FunctionCall{
			ArgList:    arg_list,
			FuncSymbol: func_sym,
		}
	case 20: //argument list construction, input is "symbol , List"
		second := words[2].(List[variables.Symbol])
		return List[variables.Symbol]{
			First:  words[0].(variables.Symbol),
			Second: &second,
		}
	case 21: //Initial list item in an argument list
		return List[variables.Symbol]{
			First:  words[0].(variables.Symbol),
			Second: nil}
	case 23: // a | b
		return booleanArithmetic(words, storage, runtime.OR)
	case 25: // a & b
		return booleanArithmetic(words, storage, runtime.AND)
	case 29: // a == b
		switch words[1].(string) {
		case "==":
			return booleanArithmetic(words, storage, runtime.EQUALS)
		case "!=":
			return booleanArithmetic(words, storage, runtime.NOTEQUALS)
		case "<":
			return booleanArithmetic(words, storage, runtime.LESS)
		case "<=":
			return booleanArithmetic(words, storage, runtime.LESSOREQUAL)
		case ">":
			return booleanArithmetic(words, storage, runtime.GREATER)
		case ">=":
			return booleanArithmetic(words, storage, runtime.GREATEROREQUAL)
		default:
			log.Fatalf("Undefined boolean operator.")
		}
	case 37: // false
		addr := storage.NewLiteral(variables.TypeDefinition{BaseType: variables.BOOL})
		storage.LoadInstruction(&runtime.InstrLoadImmediate{
			Dest:  addr,
			Value: false,
		})
		return addr
	case 38: // true
		addr := storage.NewLiteral(variables.TypeDefinition{BaseType: variables.BOOL})
		storage.LoadInstruction(&runtime.InstrLoadImmediate{
			Dest:  addr,
			Value: true,
		})
		return addr
	case 39: // declare function. func FunctionHeader FunctionBody
	case 40: // Declare new function, format "name ( arglist ) returntype"
		arg_list := words[2].(List[variables.Argument]).Iterate()
		ret_type := words[4].(variables.TypeDefinition)

		def := variables.TypeDefinition{
			BaseType:     variables.FUNC,
			ArgumentList: arg_list,
			ReturnType:   &ret_type,
		}

		storage.NewFunction(words[0].(string), def)

		return def
	case 41: //Function argument declaration list, second+ element
		second := words[2].(List[variables.Argument])
		return List[variables.Argument]{
			First:  words[0].(variables.Argument),
			Second: &second,
		}
	case 42: //Function argument declaration list, first element
		return List[variables.Argument]{
			First:  words[0].(variables.Argument),
			Second: nil,
		}
	case 43: //Function argument declaration
		return variables.Argument{
			Definition: words[0].(variables.TypeDefinition),
			Identifier: words[1].(string),
		}
	case 44: //boolean type
		return variables.TypeDefinition{BaseType: variables.BOOL}
	case 45: //int type
		return variables.TypeDefinition{BaseType: variables.INT}
	case 46: // Function scope close
		storage.LoadInstruction(&runtime.InstrExitFunction{})
		storage.DestroyFunctionScope(r)
	case 48: // If statement, NTIfHeader NTLabelledScopeBegin, NTStatementList, NTLabelledScopeClose
		jmpIfInstr := words[0].(*runtime.InstrJmpIf)
		instrEnd := words[3].(*runtime.InstructionLabelPair)
		jmpIfInstr.Label = instrEnd.Label
	case 49: //NTLabelledScopeBegin
		instr := storage.LoadLabeledInstruction(&runtime.InstrBeginScope{}, storage.NewAutoLabel())
		storage.NewScope()
		return instr
	case 50: //NTLabelledScopeClose
		storage.LoadInstruction(&runtime.InstrEndScope{})
		storage.DestroyScope()

		return storage.LoadLabeledInstruction(&runtime.InstrNOP{}, storage.NewAutoLabel())

	case 51: //Open Function
		storage.LoadInstruction(&runtime.InstrNOP{})
	case 52: //NTIfHeader (if Expr)
		condition := words[1].(variables.Symbol)
		if condition.Type.BaseType != variables.BOOL {
			log.Fatalln("Expected boolean statement in if clause, got", condition.Type)
		}

		instr := storage.LoadInstruction(&runtime.InstrJmpIf{
			Condition: condition,
			Label:     "", // will be set later.
		})
		jmp_instr := instr.Instruction.(*runtime.InstrJmpIf)
		return jmp_instr
	case 53: //If statement + WithElse
		jmpIfInstr := words[0].(*runtime.InstrJmpIf)
		jmpInstr := words[3].(*runtime.InstructionLabelPair)
		tree := words[4].(List[condition_tree_entry])

		// Add the first if clause to the tree.
		final_tree := List[condition_tree_entry]{
			First:  condition_tree_entry{start_label: "", jmp: jmpIfInstr, end: jmpInstr},
			Second: &tree,
		}

		tree_list := final_tree.Iterate()
		fmt.Println(len(tree_list))
		for i := range len(tree_list) - 1 {
			// Make so all JumpIfs (which begins each conditional block) jump to the next condition
			// Except the final one, which escapes the runtime
			tree_list[i].jmp.Label = tree_list[i+1].start_label
			// Make so all Jumps (which ends each condition block) jump to the end of the conditional
			tree_list[i].end.Instruction.(*runtime.InstrJmp).Label = tree_list[len(tree_list)-1].end.Label
		}
	case 54: //WithElse, else if statement, with continuation
		jmp := words[1].(*runtime.InstrJmpIf)
		end := words[4].(*runtime.InstructionLabelPair)
		list := words[5].(List[condition_tree_entry])

		return List[condition_tree_entry]{
			First: condition_tree_entry{
				start_label: words[0].(string),
				jmp:         jmp,
				end:         end,
			},
			Second: &list,
		}
	case 55: //WithElse, else if statement, no continuation
		jmp_if := words[1].(*runtime.InstrJmpIf)
		end := words[4].(*runtime.InstructionLabelPair)

		return List[condition_tree_entry]{
			First: condition_tree_entry{
				start_label: words[0].(string),
				jmp:         jmp_if,
				end:         end,
			},
			Second: nil,
		}

	case 56: //WithElse, else condition
		// Set label to first instruction, as this is not labelled (else condition has no JmpIf clause)
		start := words[1].(*runtime.InstructionLabelPair)
		end := words[3].(*runtime.InstructionLabelPair)

		return List[condition_tree_entry]{

			First: condition_tree_entry{
				start_label: start.Label,
				jmp:         nil,
				end:         end,
			},
			Second: nil,
		}
	case 57: // End conditional statement that is part of a larger conditional statement
		// At end of conditional block, we must jump to skip over the other conditionals
		return storage.LoadInstruction(&runtime.InstrJmp{})
	case 58: // NTBeginElseIf, used to label the first instruction in the else-if construct.
		label := storage.NewAutoLabel()
		storage.NewLabel(label)
		return label
	case 59: // arithmetic: modulo
		return integerArithmetic(words, storage, runtime.MOD)
	case 60: // return Expr
		storage.LoadInstruction(&runtime.InstrExitFunction{
			RetVal: words[1].(variables.Symbol),
		})
	case 61: //NTVarType -> function (type_list) return_type
		return_type := words[4].(variables.TypeDefinition)
		type_list := words[2].(List[variables.TypeDefinition]).Iterate()

		//Convert []TypeDefinition to []Argument by giving each type def. an empty identifier.
		var arg_list variables.ArgumentList
		for _, _type := range type_list {
			arg_list = append(arg_list, variables.Argument{
				Definition: _type,
				Identifier: "",
			})
		}

		return variables.TypeDefinition{
			BaseType:     variables.FUNC,
			ArgumentList: arg_list,
			ReturnType:   &return_type,
		}
	case 62: //Type list - part of list
		list := words[2].(List[variables.TypeDefinition])
		return List[variables.TypeDefinition]{
			First:  words[0].(variables.TypeDefinition),
			Second: &list,
		}
	case 63: //Type list - final type
		return List[variables.TypeDefinition]{
			First:  words[0].(variables.TypeDefinition),
			Second: nil,
		}
	case 64: //NTTypeVar ->  function () return_type  (no arguments)
		ret_type := words[3].(variables.TypeDefinition)
		return variables.TypeDefinition{
			ReturnType: &ret_type,
			BaseType:   variables.FUNC,
		}
	case 65: //FunctionDefinition: 0 arguments "identifier () return_type"
		ret_type := words[3].(variables.TypeDefinition)
		def := variables.TypeDefinition{
			BaseType:   variables.FUNC,
			ReturnType: &ret_type,
		}
		storage.NewFunction(words[0].(string), def)
		return def
	case 66: // Call function, 0 arguments
		var arg_list []variables.Symbol
		func_name := words[0].(string)
		sym, err := doFunctionCall(func_name, arg_list, storage)
		if err != nil {
			log.Fatal(err)
		}
		return FunctionCall{
			ArgList:    arg_list,
			FuncSymbol: sym,
		}
	case 67: // Implicit function definition: TypeDefiniiton + FunctionBody
		return words[0].(variables.Symbol)
	case 68: // New implicit function header "(arg_list) ret_type"
		arg_list := words[1].(List[variables.Argument]).Iterate()
		ret_type := words[3].(variables.TypeDefinition)

		def := variables.TypeDefinition{
			BaseType:     variables.FUNC,
			ArgumentList: arg_list,
			ReturnType:   &ret_type,
		}
		return storage.NewImplicitFunction(def)
	case 69: // New implicit function header w/o args: "() ret_type"
		ret_type := words[2].(variables.TypeDefinition)
		def := variables.TypeDefinition{
			BaseType:   variables.FUNC,
			ReturnType: &ret_type,
		}
		return storage.NewImplicitFunction(def)
	case 70: //Array, no arguments
		arr_sym := storage.NewLiteral(variables.TypeDefinition{BaseType: variables.UNDETERMINED, IsArray: true})
		storage.LoadInstruction(&runtime.InstrLoadArray{
			DestSymbol: arr_sym,
		})
		return arr_sym
	case 71: //Array, with arguments.
		list := words[1].(List[variables.Symbol]).Iterate()
		array_type := variables.UNDETERMINED
		if len(list) > 0 {
			for i := range list {
				if !list[i].Type.Equals(list[0].Type) {
					log.Fatalf("Elements in an array must be of the same type.")
				}
			}
			array_type = list[0].Type.BaseType
		}
		arr_sym := storage.NewLiteral(variables.TypeDefinition{BaseType: array_type, IsArray: true})
		storage.LoadInstruction(&runtime.InstrLoadArray{
			SrcSymbols: list,
			DestSymbol: arr_sym,
		})
		return arr_sym
	case 72:
		return variables.TypeDefinition{BaseType: variables.UNDETERMINED, IsArray: true}
	case 74:
		array := words[2].(variables.Symbol)
		storage.LoadInstruction(&runtime.InstrInitializeElevators{
			ArraySymbol: array,
		})
	//Fork function call: FunctionCall -> ~ FunctionHeader
	case 75:
		func_call := words[1].(FunctionCall)
		ret_val := storage.NewLiteral(variables.GetBaseTypeDef(variables.THREAD))
		storage.LoadInstruction(&runtime.InstrLoadImmediate{
			Dest:  ret_val,
			Value: nil, // will be set as a Thread object in the runtime.
		})
		storage.LoadInstruction(&runtime.InstrCallFunction{
			RetVal:        ret_val,
			Arguments:     func_call.ArgList,
			SymbolicLabel: func_call.FuncSymbol,
			Fork:          true,
		})
		return ret_val
	//Ordinary function call: FunctionCall -> FunctionCallHeader
	case 76:
		func_call := words[0].(FunctionCall)
		ret_val := storage.NewLiteral(*func_call.FuncSymbol.Type.ReturnType)
		storage.LoadInstruction(&runtime.InstrCallFunction{
			RetVal:        ret_val,
			Arguments:     func_call.ArgList,
			SymbolicLabel: func_call.FuncSymbol,
			Fork:          false,
		})
		return ret_val
	//Type declaration: array VarType -> BaseType [ ]
	case 78:
		_type := words[0].(variables.TypeDefinition)
		_type.IsArray = true
		return _type
	}
	return words[0]
}
