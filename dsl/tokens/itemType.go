package tokens

import (
	"fmt"
	"log"
	"strconv"
)

type Symbol int

// Constant types across all grammars
const (
	ItemError   Symbol = 10000
	ItemEpsilon Symbol = 10001
	ItemEOF     Symbol = 10002
)

const TERMINAL_START = 1
const NONTERMINAL_START = 1001

const (
	//TERMINALS
	ItemNumber Symbol = iota + TERMINAL_START
	ItemOpPlus
	ItemOpMinus
	ItemOpMult
	ItemOpDiv
	ItemOpMod
	ItemIdentifier // starts with [a-zA-Z_], followed by [a-zA-Z0-9_]
	ItemParOpen
	ItemParClosed
	ItemSemicolon
	ItemText
	ItemKeyInt
	ItemKeyBool
	ItemFalse
	ItemTrue
	ItemEquals
	ItemScopeOpen
	ItemScopeClose
	ItemComma
	ItemBoolAnd
	ItemBoolOr
	ItemBoolNot
	ItemBoolLess
	ItemBoolLessOrEqual
	ItemBoolEqual
	ItemBoolGreaterOrEqual
	ItemBoolGreater
	ItemBoolNotEqual
	ItemFunction
	ItemIf
	ItemElse
	ItemReturn
	ItemArrayOpen
	ItemArrayClose
	ItemAt
	ItemTilde
	ItemElevStatus
	ItemFor
	ItemKeyString
	ItemKeyThread
	TERMINALS_LENGTH
)

const (
	//NON-Terminals
	NTGoal Symbol = iota + NONTERMINAL_START
	NTStatement
	NTStatementList
	NTExpr
	NTTerm
	NTFactor
	NTScopeBegin
	NTScopeClose
	NTFunctionCall
	NTArgument
	NTArgList
	NTNExpr
	NTAndTerm
	NTNotTerm
	NTRelExpr
	NTRels
	NTArgumentDeclaration
	NTArgumentDeclarationList
	NTVarType
	NTFunctionOpen
	NTFunctionClose
	NTFunctionDefinition
	NTFunctionBody
	NTIfStatement
	NTLabelledScopeBegin
	NTLabelledScopeClose
	NTIfHeader
	NTWithElse
	NTEndConditionalScope
	NTBeginElseIf
	NTTypeList
	NTImplicitFunctionDefinition
	NTArrayDeclaration
	NTFunctionCallHeader
	NTBaseType
	NTForHeader
	NTEndLoopScope
	NTForPrelude
	NONTERMINALS_LENGTH
)

type Grammar struct {
	Terminals    []Symbol
	NonTerminals []Symbol
	StartSymbol  Symbol
}

func NewGrammar(startSymbol Symbol) Grammar {
	grammar := Grammar{
		StartSymbol: startSymbol,
	}

	for i := TERMINAL_START; i < int(TERMINALS_LENGTH); i++ {
		grammar.Terminals = append(grammar.Terminals, Symbol(i))
	}
	for i := NONTERMINAL_START; i < int(NONTERMINALS_LENGTH); i++ {
		grammar.NonTerminals = append(grammar.NonTerminals, Symbol(i))
	}
	return grammar
}

func (grammar *Grammar) MapToArrayindex(item Symbol) int {
	if item.IsTerminal() {
		return int(item) - 1
	} else if item.IsNonTerminal() {
		return int(item) - 1002
	} else if item == ItemEOF {
		return len(grammar.Terminals)
	}

	log.Fatalf("Attempted to store %s in an array!", item.String())
	panic("See log")
}

func (grammar *Grammar) NewCategoryID(item Symbol) Symbol {
	temp := (Symbol)(grammar.NonTerminals[len(grammar.NonTerminals)-1] + 1)
	fmt.Println(temp)
	_new_str[temp] = item.String() + "'"
	grammar.NonTerminals = append(grammar.NonTerminals, temp)
	return temp
}

type Token struct {
	Symbol Symbol
	Lexeme string
	Line   int
	Col    int
}

func (l Token) String() string {
	switch l.Symbol {
	case ItemEOF:
		return "EOF"
	case ItemError:
		return l.Lexeme
	}

	if len(l.Lexeme) > 50 {
		return fmt.Sprintf("%d %s", int(l.Symbol), l.Lexeme[:50])
	}

	return fmt.Sprintf("%d %q", int(l.Symbol), l.Lexeme)
}

func (l Symbol) IsTerminal() bool {
	return l >= 0 && l < 1000
}

func (l Symbol) IsNonTerminal() bool {
	return l >= 1000 && l < 10000
}

var _new_str map[Symbol]string = make(map[Symbol]string)

func (i Symbol) String() string {
	return strconv.Itoa(int(i))
}
