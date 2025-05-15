package parser

import (
	"autofat/dsl/structure"
	"autofat/dsl/tokens"
	"fmt"
	"sort"
	"strings"
)

// An alternative is the right-hand side of a single production, e.g. the "B" in A -> B
type cfg_alternative []tokens.Symbol

// A set of right-hand side productions, typically grouped together by a common left-hand side symbol.
type cfg_rule []cfg_alternative

type cfg_pattern struct {
	A tokens.Symbol
	B cfg_alternative
}

// A CFG (context-free grammar) holds all valid productions (A -> B) in a grammar.
type CFG struct {
	_array []cfg_pattern
	_map   map[tokens.Symbol]structure.Set[int]
}

func (cfg *CFG) Productions() []cfg_pattern {
	return cfg._array
}

func NewCFG() CFG {
	return CFG{
		_array: make([]cfg_pattern, 0),
		_map:   make(map[tokens.Symbol]structure.Set[int]),
	}
}

// Add a rule A -> B
func (cfg *CFG) addRule(A tokens.Symbol, B cfg_alternative) {
	cfg._array = append(cfg._array, cfg_pattern{A: A, B: B})
}

// Add multiple rules of form A -> B1, A -> B2, ... where B1, B2, .. are elements of Bs
func (cfg *CFG) addRules(A tokens.Symbol, Bs []cfg_alternative) {
	for _, B := range Bs {
		cfg.addRule(A, B)
	}
}

// Compiles all the rules by creating a map with the left-hand side of the production
// as the key, and a set as the value corresponding to all rules with this left-hand side
// The map significantly speeds up parsing by quickly retrieving rules from a left-hand side symbol.
// Must be called when done adding rules!
func (cfg *CFG) compile() {
	for idx, alt := range cfg._array {
		_, ok := cfg._map[alt.A]
		if !ok {
			cfg._map[alt.A] = *structure.NewSet[int]()
		}
		cfg._map[alt.A].Add(idx)
	}
}

func (cfg *CFG) updateRule(A tokens.Symbol, B cfg_rule) {
	indices := cfg._map[A].List()

	if len(indices) != len(B) {
		fmt.Printf("Cannot update a rule like this! %d %d", len(indices), len(B))
		panic("!")
	}
	for _, index := range indices {
		cfg._array[index].B = B[index]
	}

}

func (cfg CFG) GetRulesForA(token tokens.Symbol) []cfg_alternative {
	ret := make([]cfg_alternative, 0)
	for _, idx := range cfg._map[token].List() {
		ret = append(ret, cfg._array[idx].B)
	}
	return ret
}

func (cfg CFG) GetRuleIndexesForA(token tokens.Symbol) map[int]struct{} {
	return cfg._map[token].Elements()
}

func (cfg CFG) RuleByIndex(index int) cfg_pattern {
	return cfg._array[index]
}

func CreateCFG() CFG {
	cfg := NewCFG()

	cfg.addRule(tokens.NTGoal, cfg_alternative{tokens.NTStatementList})                              // 1
	cfg.addRule(tokens.NTStatementList, cfg_alternative{tokens.NTStatement})                         // 2
	cfg.addRule(tokens.NTStatementList, cfg_alternative{tokens.NTStatement, tokens.NTStatementList}) // 3

	cfg.addRule(tokens.NTNExpr, cfg_alternative{tokens.NTNExpr, tokens.ItemOpPlus, tokens.NTTerm})  // 4
	cfg.addRule(tokens.NTNExpr, cfg_alternative{tokens.NTNExpr, tokens.ItemOpMinus, tokens.NTTerm}) // 5
	cfg.addRule(tokens.NTNExpr, cfg_alternative{tokens.NTTerm})                                     // 6

	cfg.addRule(tokens.NTTerm, cfg_alternative{tokens.NTTerm, tokens.ItemOpMult, tokens.NTFactor}) // 7
	cfg.addRule(tokens.NTTerm, cfg_alternative{tokens.NTTerm, tokens.ItemOpDiv, tokens.NTFactor})  // 8
	cfg.addRule(tokens.NTTerm, cfg_alternative{tokens.NTFactor})                                   // 9

	cfg.addRules(tokens.NTFactor, []cfg_alternative{
		{tokens.ItemParOpen, tokens.NTExpr, tokens.ItemParClosed}, // 10
		{tokens.ItemNumber},     // 11
		{tokens.ItemIdentifier}, // 12
	})

	cfg.addRules(tokens.NTStatement, []cfg_alternative{
		{tokens.NTVarType, tokens.ItemIdentifier, tokens.ItemEquals, tokens.NTExpr, tokens.ItemSemicolon}, // 13
		{tokens.ItemIdentifier, tokens.ItemEquals, tokens.NTExpr, tokens.ItemSemicolon},                   // 14
		{tokens.NTExpr, tokens.ItemSemicolon},                                                             // 15
	})

	cfg.addRule(tokens.NTStatement, cfg_alternative{tokens.NTScopeBegin, tokens.NTStatement, tokens.NTScopeClose}) // 16

	cfg.addRule(tokens.NTScopeBegin, cfg_alternative{tokens.ItemScopeOpen})
	cfg.addRule(tokens.NTScopeClose, cfg_alternative{tokens.ItemScopeClose})

	cfg.addRule(tokens.NTFactor, cfg_alternative{tokens.NTFunctionCall})
	cfg.addRule(tokens.NTFunctionCallHeader, cfg_alternative{tokens.ItemIdentifier, tokens.ItemParOpen, tokens.NTArgList, tokens.ItemParClosed})
	cfg.addRule(tokens.NTArgList, cfg_alternative{tokens.NTArgument, tokens.ItemComma, tokens.NTArgList})
	cfg.addRule(tokens.NTArgList, cfg_alternative{tokens.NTArgument})

	cfg.addRule(tokens.NTArgument, cfg_alternative{tokens.NTExpr})

	cfg.addRules(tokens.NTExpr, []cfg_alternative{
		{tokens.NTExpr, tokens.ItemBoolOr, tokens.NTAndTerm},
		{tokens.NTAndTerm},
	})
	cfg.addRules(tokens.NTAndTerm, []cfg_alternative{
		{tokens.NTAndTerm, tokens.ItemBoolAnd, tokens.NTNotTerm},
		{tokens.NTNotTerm},
	})
	cfg.addRules(tokens.NTNotTerm, []cfg_alternative{
		{tokens.ItemBoolNot, tokens.NTRelExpr},
		{tokens.NTRelExpr},
	})
	cfg.addRules(tokens.NTRelExpr, []cfg_alternative{
		{tokens.NTNExpr, tokens.NTRels, tokens.NTNExpr},
		{tokens.NTNExpr},
	})

	cfg.addRules(tokens.NTRels, []cfg_alternative{
		{tokens.ItemBoolEqual},
		{tokens.ItemBoolNotEqual},
		{tokens.ItemBoolLess},
		{tokens.ItemBoolLessOrEqual},
		{tokens.ItemBoolGreater},
		{tokens.ItemBoolGreaterOrEqual},
	})

	cfg.addRules(tokens.NTFactor, []cfg_alternative{
		{tokens.ItemFalse},
		{tokens.ItemTrue},
	})

	// Note that we dont use NTScopeOpen here, because other parts of the function need to create the scope for us
	// Closing is done using a special non-terminal, however.
	cfg.addRule(tokens.NTStatement,
		cfg_alternative{tokens.ItemFunction, tokens.NTFunctionDefinition, tokens.NTFunctionBody}) //39

	cfg.addRule(tokens.NTFunctionDefinition,
		cfg_alternative{tokens.ItemIdentifier, tokens.ItemParOpen, tokens.NTArgumentDeclarationList, tokens.ItemParClosed, tokens.NTVarType}) //40

	cfg.addRules(tokens.NTArgumentDeclarationList, []cfg_alternative{
		{tokens.NTArgumentDeclaration, tokens.ItemComma, tokens.NTArgumentDeclarationList}, //41
		{tokens.NTArgumentDeclaration}, //42
	})

	cfg.addRules(tokens.NTArgumentDeclaration, []cfg_alternative{
		{tokens.NTVarType, tokens.ItemIdentifier}, //43
	})

	cfg.addRules(tokens.NTBaseType, []cfg_alternative{
		{tokens.ItemKeyBool}, //44
		{tokens.ItemKeyInt},  //45
	})

	cfg.addRule(tokens.NTFunctionClose, cfg_alternative{tokens.ItemScopeClose}) // 46

	cfg.addRule(tokens.NTFunctionBody, cfg_alternative{tokens.NTFunctionOpen, tokens.NTStatementList, tokens.NTFunctionClose}) // 47

	cfg.addRule(tokens.NTStatement, cfg_alternative{tokens.NTIfHeader, // 48
		tokens.NTScopeBegin,
		tokens.NTStatementList,
		tokens.NTLabelledScopeClose,
	})

	cfg.addRule(tokens.NTLabelledScopeBegin, cfg_alternative{tokens.ItemScopeOpen})  // 49
	cfg.addRule(tokens.NTLabelledScopeClose, cfg_alternative{tokens.ItemScopeClose}) // 50
	cfg.addRule(tokens.NTFunctionOpen, cfg_alternative{tokens.ItemScopeOpen})        // 51

	cfg.addRule(tokens.NTIfHeader, cfg_alternative{tokens.ItemIf, tokens.NTExpr}) // 52

	cfg.addRule(tokens.NTStatement, cfg_alternative{tokens.NTIfHeader, tokens.NTScopeBegin, tokens.NTStatementList, tokens.NTEndConditionalScope,
		tokens.NTWithElse}) // 53

	cfg.addRule(tokens.NTWithElse, cfg_alternative{tokens.NTBeginElseIf, tokens.NTIfHeader, tokens.NTScopeBegin, tokens.NTStatementList, tokens.NTEndConditionalScope,
		tokens.NTWithElse}) // 54
	cfg.addRule(tokens.NTWithElse,
		cfg_alternative{tokens.NTBeginElseIf, tokens.NTIfHeader, tokens.NTScopeBegin, tokens.NTStatementList, tokens.NTLabelledScopeClose}) // 55
	cfg.addRule(tokens.NTWithElse, cfg_alternative{tokens.ItemElse, tokens.NTLabelledScopeBegin, tokens.NTStatementList, tokens.NTLabelledScopeClose}) // 56
	cfg.addRule(tokens.NTEndConditionalScope, cfg_alternative{tokens.NTScopeClose})                                                                    // 57
	cfg.addRule(tokens.NTBeginElseIf, cfg_alternative{tokens.ItemElse})                                                                                //58
	cfg.addRule(tokens.NTTerm, cfg_alternative{tokens.NTTerm, tokens.ItemOpMod, tokens.NTFactor})                                                      // 59
	cfg.addRule(tokens.NTStatement, cfg_alternative{tokens.ItemReturn, tokens.NTExpr, tokens.ItemSemicolon})                                           // 60
	//61
	cfg.addRule(tokens.NTBaseType, cfg_alternative{tokens.ItemFunction, tokens.ItemParOpen, tokens.NTTypeList, tokens.ItemParClosed, tokens.NTVarType})
	//62
	cfg.addRule(tokens.NTTypeList, cfg_alternative{tokens.NTVarType, tokens.ItemComma, tokens.NTTypeList})
	//63
	cfg.addRule(tokens.NTTypeList, cfg_alternative{tokens.NTVarType})
	//64 - Declare function type, empty argument list
	cfg.addRule(tokens.NTBaseType, cfg_alternative{tokens.ItemFunction, tokens.ItemParOpen, tokens.ItemParClosed, tokens.NTVarType})
	//65 - Function definition, no arguments
	cfg.addRule(tokens.NTFunctionDefinition,
		cfg_alternative{tokens.ItemIdentifier, tokens.ItemParOpen, tokens.ItemParClosed, tokens.NTVarType})
	//66 - Function call, no arguments
	cfg.addRule(tokens.NTFunctionCallHeader, cfg_alternative{tokens.ItemIdentifier, tokens.ItemParOpen, tokens.ItemParClosed})
	//67 - Implicit function definition
	cfg.addRule(tokens.NTExpr, cfg_alternative{tokens.NTImplicitFunctionDefinition, tokens.NTFunctionBody})
	//68 - Implicit function definition header
	cfg.addRule(tokens.NTImplicitFunctionDefinition, cfg_alternative{tokens.ItemParOpen, tokens.NTArgumentDeclarationList, tokens.ItemParClosed, tokens.NTVarType})
	//69 - Implicit function definition, 0 arguments
	cfg.addRule(tokens.NTImplicitFunctionDefinition, cfg_alternative{tokens.ItemParOpen, tokens.ItemParClosed, tokens.NTVarType})
	//70 - Array - no elements
	cfg.addRule(tokens.NTArrayDeclaration, cfg_alternative{tokens.ItemArrayOpen, tokens.ItemArrayClose})
	//71 - Array - with elements
	cfg.addRule(tokens.NTArrayDeclaration, cfg_alternative{tokens.ItemArrayOpen, tokens.NTArgList, tokens.ItemArrayClose})
	//72 - array type declaration
	cfg.addRule(tokens.NTBaseType, cfg_alternative{tokens.ItemElevStatus})
	//73 - Expr -> Array
	cfg.addRule(tokens.NTExpr, cfg_alternative{tokens.NTArrayDeclaration})
	//74 - Special assignment to @ to define elevators
	cfg.addRule(tokens.NTStatement, cfg_alternative{tokens.ItemAt, tokens.ItemEquals, tokens.NTNExpr, tokens.ItemSemicolon})
	//75 - Fork function call
	cfg.addRule(tokens.NTFunctionCall, cfg_alternative{tokens.ItemTilde, tokens.NTFunctionCallHeader})
	//76 - Ordinary function call
	cfg.addRule(tokens.NTFunctionCall, cfg_alternative{tokens.NTFunctionCallHeader})
	//77 - Type, no array
	cfg.addRule(tokens.NTVarType, cfg_alternative{tokens.NTBaseType})
	//78 - Type, is array
	cfg.addRule(tokens.NTVarType, cfg_alternative{tokens.NTBaseType, tokens.ItemArrayOpen, tokens.ItemArrayClose})
	//79 - @ used in Expr
	cfg.addRule(tokens.NTExpr, cfg_alternative{tokens.ItemAt})
	fmt.Println("Num rules: ", len(cfg._array))
	cfg.compile()

	return cfg
}

func EliminateLeftRecursion(cfg CFG, grammar *tokens.Grammar) CFG {
	keys := make([]tokens.Symbol, len(cfg._map))
	i := 0
	for a := range cfg._map {
		keys[i] = a
		i += 1
	}
	sort.Slice(keys, func(i, j int) bool { return int(keys[i]) < int(keys[j]) })

	new_cfg := NewCFG()
	new_cfg._array = make([]cfg_pattern, len(cfg._array))
	copy(new_cfg._array, cfg._array)

	for i := range keys {
		new_cfg._map[keys[i]] = cfg._map[keys[i]].Copy()
		for s := 0; s < i; s++ {
			for _, rule := range cfg.GetRulesForA(keys[i]) {
				if rule.startsWith(keys[s]) { // Find forms A_i -> A_s T
					//Replace them with forms A_i -> d_i T, where d_i is rules of the form A_i -> d_i
					new_cfg.updateRule(keys[i], rule_append(cfg.GetRulesForA(keys[s]), rule[1:]))
				}
			}
		}

		//Eliminate direct A_i recursion
		var A []cfg_alternative // Rules that follow the recursed variable, e.g. E -> E a
		var B []cfg_alternative // Rules that do not follow the recursed variable, e.g. E -> B
		var A_indices []int
		var B_indices []int
		for _, index := range new_cfg._map[keys[i]].List() {
			alt := new_cfg._array[index]
			if alt.B.startsWith(keys[i]) {
				A = append(A, alt.B[1:])
				A_indices = append(A_indices, index)
			} else {
				B = append(B, alt.B)
				B_indices = append(B_indices, index)
			}
		}

		// We found an option of the form E -> E a. Must recurse.
		if len(A) > 0 {
			// Create a new token from E, E' (E tilde)
			new_token := grammar.NewCategoryID(keys[i])

			// Set up so that E' -> (a E' | epsilon)
			A = rule_append(A, []tokens.Symbol{new_token})

			new_cfg._map[new_token] = *structure.NewSet[int]()
			for j, A_index := range A_indices {
				new_cfg._array[A_index] = cfg_pattern{
					A: new_token,
					B: A[j],
				}
				new_cfg._map[keys[i]].Remove(A_index)
				new_cfg._map[new_token].Add(A_index)
			}
			// add the epsilon rule
			new_cfg._array = append(new_cfg._array, cfg_pattern{
				A: new_token,
				B: cfg_alternative{tokens.ItemEpsilon},
			})
			new_cfg._map[new_token].Add(len(new_cfg._array) - 1)

			// Set up such that E -> B E'
			B = rule_append(B, []tokens.Symbol{new_token})
			for j, B_index := range B_indices {
				new_cfg._array[B_index].B = B[j]
			}
		}

	}
	return new_cfg
}

func (rule cfg_rule) String() string {
	var out []string
	for _, alt := range rule {
		out = append(out, alt.String())
	}
	return strings.Join(out, "| ")
}
func (alt cfg_alternative) String() string {
	s := ""
	for _, item := range alt {
		s += fmt.Sprint(item, " ")
	}
	return s
}

func (cfg CFG) String() string {
	s := ""
	for a, b := range cfg._map {
		for _, i := range b.List() {
			s += fmt.Sprintf("%s\t> %s\n", a, cfg._array[i].B)
		}
	}
	return s
}

func (alt cfg_alternative) startsWith(i tokens.Symbol) bool {
	return alt[0] == i
}

func rule_append(rule cfg_rule, to_append []tokens.Symbol) cfg_rule {
	var new_rules cfg_rule
	for i := range rule {
		var new_alt cfg_alternative
		new_alt = append(new_alt, rule[i]...)
		new_alt = append(new_alt, to_append...)
		new_rules = append(new_rules, new_alt)
	}
	return new_rules
}
