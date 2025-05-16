package scanner

import (
	"autofat/dsl/tokens"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

func Lex(input string) (*lexer, chan tokens.Token) {
	l := &lexer{
		input: input,
		items: make(chan tokens.Token),
	}
	go l.run()
	return l, l.items
}

func (l *lexer) run() {
	for state := lexInsideScope; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t tokens.Symbol) {
	l.items <- tokens.Token{Symbol: t, Lexeme: l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) next() rune {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	_rune, width := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = width
	l.pos += l.width
	return _rune
}

func (l *lexer) match(next string, token tokens.Symbol) bool {
	for i := 0; i < len(next); i++ {
		ch := l.next()
		if ch != rune(next[i]) {
			for j := i; j >= 0; j++ {
				l.backup()
			}
			return false
		}
	}
	l.emit(token)
	return true
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) backup() {
	l.pos -= l.width
}

func (l *lexer) peek() rune {
	_rune := l.next()
	l.backup()
	return _rune
}

func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

func (l *lexer) acceptRegex(valid string) {
	re := regexp.MustCompile(`^[` + valid + `]+`)

	ret := re.FindStringIndex(l.input[l.pos:])
	if ret != nil {
		l.pos += ret[1]
	}
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- tokens.Token{
		Symbol: tokens.ItemError,
		Lexeme: fmt.Sprintf(format, args...),
	}
	return nil
}
