package parser

import (
	"autofat/dsl/structure"
	"autofat/dsl/tokens"
	"fmt"
)

func Start(A tokens.Symbol, B tokens.Symbol, grammar *tokens.Grammar, first FirstSet, follow FirstSet) structure.Set[tokens.Symbol] {
	if !first[B].Contains(tokens.ItemEpsilon) {
		return first[B].Copy()
	}
	ret := first[B].Copy().Remove(tokens.ItemEpsilon).Union(follow[A])

	fmt.Println(first[B], follow[A])
	return ret
}
