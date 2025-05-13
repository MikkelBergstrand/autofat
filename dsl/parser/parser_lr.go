package parser

import (
	"autofat/dsl/runtime"
	"autofat/dsl/storage"
	"autofat/dsl/structure"
	"autofat/dsl/tokens"
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

type lr_item struct {
	production_id int
	dot_pos       int
	lookahead     tokens.Symbol
}

func (c lr_item) hash() string {
	return fmt.Sprintf("%s %s %s", strconv.Itoa(c.production_id), strconv.Itoa(c.dot_pos), strconv.Itoa(int(c.lookahead)))
}

type item_collection struct {
	items     []lr_item
	hashTable structure.Set[string]
}

func newCollection() item_collection {
	return item_collection{
		hashTable: *structure.NewSet[string](),
	}
}

func (c *item_collection) String(cfg CFG) string {
	s := ""
	for _, item := range c.items {
		s += fmt.Sprintf("[ %s -> ", cfg._array[item.production_id].A)
		for indexB, B := range cfg._array[item.production_id].B {
			if item.dot_pos == indexB {
				s += ". "
			}
			s += fmt.Sprintf("%s ", B)
		}
		if item.dot_pos == len(cfg._array[item.production_id].B) {
			s += "."
		}
		s += fmt.Sprintf(", %s]\n", item.lookahead)
	}
	return s
}

func (c *item_collection) add(new_obj lr_item) {
	hash := new_obj.hash()
	if !c.hashTable.Contains(hash) {
		c.hashTable.Add(hash)
		c.items = append(c.items, new_obj)
	}
}

func (c *item_collection) compare(other *item_collection) bool {
	return reflect.DeepEqual(c.hashTable, other.hashTable) // May be possible to optimize?
}

func makeClosure(col item_collection, cfg CFG, first FirstSet) item_collection {
	oldHashSize := -1
	for oldHashSize != col.hashTable.Size() {
		oldHashSize = col.hashTable.Size()
		for i := range col.items {
			lookahead := col.items[i].lookahead
			rule := cfg.RuleByIndex(col.items[i].production_id)
			C := tokens.ItemError
			if col.items[i].dot_pos < len(rule.B) {
				C = rule.B[col.items[i].dot_pos]
			}
			if col.items[i].dot_pos < len(rule.B)-1 {
				lookahead = rule.B[col.items[i].dot_pos+1]
			}

			if C != tokens.ItemError {
				for index := range cfg.GetRuleIndexesForA(C) {
					for _, firstItem := range first[lookahead].List() {
						new_obj := lr_item{
							production_id: index,
							dot_pos:       0,
							lookahead:     firstItem,
						}
						col.add(new_obj)
					}
				}
			}

		}
	}
	return col
}

func makeGoto(col item_collection, cfg CFG, first FirstSet, x tokens.Symbol) item_collection {
	new_col := newCollection()

	for i := range col.items {
		rule := cfg.RuleByIndex(col.items[i].production_id).B
		if col.items[i].dot_pos < len(rule) && rule[col.items[i].dot_pos] == x {
			new_obj := col.items[i]
			new_obj.dot_pos += 1
			new_col.add(new_obj)
		}
	}

	ret := makeClosure(new_col, cfg, first)
	return ret
}

func computeCanonicalCollections(grammar tokens.Grammar, cfg CFG, first FirstSet) []item_collection {
	initial_closure := newCollection()

	for rule_id := range cfg.GetRuleIndexesForA(grammar.StartSymbol) {
		initial_closure.add(
			lr_item{
				production_id: rule_id,
				dot_pos:       0,
				lookahead:     tokens.ItemEOF,
			})
	}

	cc0 := makeClosure(initial_closure, cfg, first)
	cc_list := []item_collection{cc0}

	cc_index := 0
	prev_cc_size := 0
	for len(cc_list) != prev_cc_size {
		prev_cc_size = len(cc_list)
		for ; cc_index < len(cc_list); cc_index++ {
			items := cc_list[cc_index].items
			follows_dot := structure.NewSet[tokens.Symbol]()
			for i := range items {
				rule := cfg.RuleByIndex(items[i].production_id)
				if items[i].dot_pos < len(rule.B) {
					follows_dot.Add(rule.B[items[i].dot_pos])
				}
			}

			for _, x := range follows_dot.List() {
				temp := makeGoto(cc_list[cc_index], cfg, first, x)
				alreadyExists := false
				for cc_indexj := range cc_list {
					if temp.compare(&cc_list[cc_indexj]) {
						alreadyExists = true
						break
					}
				}

				if !alreadyExists {
					cc_list = append(cc_list, temp)
				}
			}
		}
	}

	return cc_list
}

type ActionType int

const (
	ACTION_REDUCE ActionType = iota
	ACTION_SHIFT
	ACTION_ACCEPT
)

type Action struct {
	Type  ActionType
	Value int
}

type ActionTable [][]Action
type GotoTable [][]int

type LRParser struct {
	ActionTable ActionTable
	GotoTable   GotoTable
}

func CreateLRParser(grammar tokens.Grammar, cfg CFG, first FirstSet) LRParser {
	closures := computeCanonicalCollections(grammar, cfg, first)

	type goto_key struct {
		x       tokens.Symbol
		closure int
	}

	actionTable := make(ActionTable, len(closures))
	gotoTable := make(GotoTable, len(closures))

	goto_cache := make(map[goto_key]int)
	_goto := func(i int, x tokens.Symbol) int {
		goto_key := goto_key{closure: i, x: x}
		_, ok := goto_cache[goto_key]
		if !ok {
			tmp := makeGoto(closures[i], cfg, first, goto_key.x)
			val := -1
			for idx := range closures {
				if tmp.compare(&closures[idx]) {
					val = idx
					break
				}
			}
			goto_cache[goto_key] = val
		}
		return goto_cache[goto_key]
	}

	for i := range len(closures) {
		actionTable[i] = make([]Action, len(grammar.Terminals)+1)
		gotoTable[i] = make([]int, len(grammar.NonTerminals)-1)

		for j := range actionTable[i] {
			actionTable[i][j] = Action{Type: -1, Value: -1}
		}
		for j := range gotoTable[i] {
			gotoTable[i][j] = -1
		}
	}

	eofIndex := len(grammar.Terminals)

	for cc_i, closure := range closures {
		for _, item := range closure.items {
			set := false
			rule := cfg.RuleByIndex(item.production_id)
			if item.dot_pos < len(rule.B) && rule.B[item.dot_pos].IsTerminal() {
				x := rule.B[item.dot_pos]
				cc_j := _goto(cc_i, x)
				set = cc_j >= 0
				if set {
					idx := grammar.MapToArrayindex(x)
					actionTable[cc_i][idx] = Action{
						Type:  ACTION_SHIFT,
						Value: cc_j,
					}
				}
			}

			if !set && item.dot_pos == len(rule.B) &&
				rule.A == grammar.StartSymbol && item.lookahead == tokens.ItemEOF {
				actionTable[cc_i][eofIndex] = Action{
					Type: ACTION_ACCEPT,
				}
			} else if !set && item.dot_pos == len(rule.B) {
				idx := grammar.MapToArrayindex(item.lookahead)
				actionTable[cc_i][idx] = Action{
					Type:  ACTION_REDUCE,
					Value: item.production_id,
				}
			}
		}

		for _, nt := range grammar.NonTerminals {
			if nt == grammar.StartSymbol {
				continue
			}
			cc_j := _goto(cc_i, nt)
			if cc_j >= 0 {
				idx := grammar.MapToArrayindex(nt)
				gotoTable[cc_i][idx] = cc_j
			}
		}
	}
	return LRParser{
		ActionTable: actionTable,
		GotoTable:   gotoTable,
	}
}

func (parser *LRParser) Parse(words <-chan tokens.Token, cfg CFG, grammar tokens.Grammar,
	storage *storage.Compiler, rt *runtime.Runtime) (int, error) {
	type stack_state struct {
		symbol tokens.Symbol
		state  int
		value  any
	}

	stack := structure.NewStack[stack_state]()
	stack.Push(stack_state{tokens.ItemError, -1, nil})
	stack.Push(stack_state{grammar.StartSymbol, 0, nil})

	actionTable := parser.ActionTable
	gotoTable := parser.GotoTable

	word := <-words

	for {
		state := stack.Peek()
		action := actionTable[state.state][grammar.MapToArrayindex(word.Symbol)]

		switch action.Type {
		case ACTION_REDUCE:
			rule := cfg.RuleByIndex(action.Value)

			popped := make([]any, len(rule.B))
			for i := len(rule.B) - 1; i >= 0; i-- {
				pop := stack.Pop()
				popped[i] = pop.value
			}

			value := DoActions(action.Value, popped, storage, rt)

			state = stack.Peek()
			_goto := gotoTable[state.state][grammar.MapToArrayindex(rule.A)]
			if _goto < 0 {
				return 0, errors.New("bad goto")
			}
			stack.Push(stack_state{rule.A, _goto, value})
		case ACTION_SHIFT:
			stack.Push(stack_state{word.Symbol, action.Value, word.Lexeme})
			word = <-words
		case ACTION_ACCEPT:
			if word.Symbol == tokens.ItemEOF {
				storage.LoadInstruction(&runtime.InstrExitFunction{})
				start, _ := storage.DestroyFunctionScope(rt) //Destroy the final (outermost) scope
				return start, nil
			} else {
				return 0, errors.New("syntax error")
			}
		default:
			return 0, errors.New(fmt.Sprintln("invalid action state on", word))
		}
	}
}
