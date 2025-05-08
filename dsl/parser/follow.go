package parser

import (
	"autofat/dsl/structure"
	"autofat/dsl/tokens"
)

func Follow(cfg CFG, grammar tokens.Grammar, first FirstSet) FirstSet {
	follow := make(FirstSet)
	for _, term := range grammar.NonTerminals {
		follow[term] = *structure.NewSet[tokens.Symbol]()
	}

	follow[grammar.StartSymbol].Add(tokens.ItemEOF)

	changing := true
	for changing {
		changing = false
		for _, rule := range cfg.Productions() {
			lhs := rule.A
			alt := rule.B
			trailer := follow[lhs].Copy()
			for i := len(alt) - 1; i >= 0; i-- {
				if alt[i].IsNonTerminal() {
					prev_size := follow[alt[i]].Size()
					follow[alt[i]].Union(trailer)

					if !changing && prev_size != follow[alt[i]].Size() {
						changing = true
					}

					if first[alt[i]].Contains(tokens.ItemEpsilon) {
						alreadyContainsEpsilon := trailer.Contains(tokens.ItemEpsilon)
						trailer.Union(first[alt[i]])
						if !alreadyContainsEpsilon {
							trailer.Remove(tokens.ItemEpsilon)
						}
					} else {
						trailer.Union(first[alt[i]])
					}
				} else if alt[i].IsTerminal() {
					trailer = first[alt[i]].Copy()
				}
			}
		}
	}
	return follow
}
