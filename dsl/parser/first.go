package parser

import (
	"autofat/dsl/structure"
	"autofat/dsl/tokens"
	"fmt"
)

type FirstSet map[tokens.Symbol]structure.Set[tokens.Symbol]

func (set FirstSet) String() string {
	s := ""
	for k, v := range set {
		s += fmt.Sprintf("%s: ", k)

		for _, val := range v.List() {
			s += fmt.Sprintf("%s ", val)
		}
		s += "\n"
	}
	return s
}
func First(cfg CFG, grammar tokens.Grammar) FirstSet {
	firstSet := make(FirstSet)

	for _, terminal := range grammar.Terminals {
		firstSet[terminal] = *structure.NewSet[tokens.Symbol]()
		firstSet[terminal].Add(terminal)
	}

	firstSet[tokens.ItemEOF] = *structure.NewSet[tokens.Symbol]()
	firstSet[tokens.ItemEOF].Add(tokens.ItemEOF)

	firstSet[tokens.ItemEpsilon] = *structure.NewSet[tokens.Symbol]()
	firstSet[tokens.ItemEpsilon].Add(tokens.ItemEpsilon)

	for _, terminal := range grammar.NonTerminals {
		firstSet[terminal] = *structure.NewSet[tokens.Symbol]()
	}

	for _, rule := range cfg.Productions() {
		if len(rule.B) == 0 {
			firstSet[rule.A].Add(tokens.ItemEpsilon)
		}
	}

	changing := true
	for changing {
		changing = false
		for _, rule := range cfg.Productions() {
			A := rule.A
			Bs := rule.B

			if len(Bs) == 0 {
				continue
			}

			rhs := firstSet[Bs[0]].Copy().Remove(tokens.ItemEpsilon)
			trailing := true

			for i := 0; i < len(Bs)-1; i++ {
				if firstSet[Bs[i]].Contains(tokens.ItemEpsilon) {
					epsilonAlreadyPresent := rhs.Contains(tokens.ItemEpsilon)
					rhs.Union(firstSet[Bs[i+1]])
					if !epsilonAlreadyPresent {
						rhs.Remove(tokens.ItemEpsilon)
					}
				} else {
					trailing = false
					break
				}
			}

			if trailing && firstSet[Bs[len(Bs)-1]].Contains(tokens.ItemEpsilon) {
				rhs.Add(tokens.ItemEpsilon)
			}

			prevCount := firstSet[A].Size()
			firstSet[A].Union(rhs)

			// Compare size of set before and after union, determine if FIRST is still changing
			if !changing && prevCount != firstSet[A].Size() {
				changing = true
			}
		}
	}

	return firstSet
}
