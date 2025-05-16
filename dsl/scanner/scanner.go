package scanner

import (
	"autofat/dsl/tokens"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

type lexer struct {
	input string
	start int
	pos   int
	width int
	items chan tokens.Token
}

const eof rune = '\x00' // necessary in 2025?

type stateFn func(*lexer) stateFn

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

func isAlphaNumeric(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9')
}

func isAlpha(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z')
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\n'
}

func lexInsideExpression(l *lexer) stateFn {
	for {
		r := l.next()

		if isSpace(r) {
			l.ignore()
		} else if r == '=' {
			if l.peek() == '=' {
				l.next()
				l.emit(tokens.ItemBoolEqual)
			} else {
				l.emit(tokens.ItemEquals)
			}
			return lexInsideExpression
		} else if r == '"' {
			return lexQuote
		} else if r == '+' {
			l.emit(tokens.ItemOpPlus)
			return lexInsideExpression
		} else if r == '-' {
			l.emit(tokens.ItemOpMinus)
			return lexInsideExpression
		} else if r == '*' {
			l.emit(tokens.ItemOpMult)
			return lexInsideExpression
		} else if r == '/' {
			l.emit(tokens.ItemOpDiv)
			return lexInsideExpression
		} else if r == '%' {
			l.emit(tokens.ItemOpMod)
			return lexInsideExpression
		} else if r == '(' {
			l.emit(tokens.ItemParOpen)
			return lexInsideExpression
		} else if r == ')' {
			l.emit(tokens.ItemParClosed)
			return lexInsideExpression
		} else if r == '{' {
			l.emit(tokens.ItemScopeOpen)
			return lexInsideExpression
		} else if r == '}' {
			l.emit(tokens.ItemScopeClose)
			return lexInsideExpression
		} else if r == '[' {
			l.emit(tokens.ItemArrayOpen)
			return lexInsideExpression
		} else if r == ']' {
			l.emit(tokens.ItemArrayClose)
			return lexInsideExpression
		} else if r == ',' {
			l.emit(tokens.ItemComma)
			return lexInsideExpression
		} else if r == '&' {
			l.emit(tokens.ItemBoolAnd)
			return lexInsideExpression
		} else if r == '|' {
			l.emit(tokens.ItemBoolOr)
			return lexInsideExpression
		} else if r == '<' {
			if l.peek() == '=' {
				l.next()
				l.emit(tokens.ItemBoolLessOrEqual)
			} else {
				l.emit(tokens.ItemBoolLess)
			}
		} else if r == '>' {
			if l.peek() == '=' {
				l.next()
				l.emit(tokens.ItemBoolGreaterOrEqual)
			} else {
				l.emit(tokens.ItemBoolGreater)
			}
			return lexInsideExpression
		} else if r == '!' {
			if l.peek() == '=' {
				l.next()
				l.emit(tokens.ItemBoolNotEqual)
			} else {
				l.emit(tokens.ItemBoolNot)
			}
			return lexInsideExpression
		} else if r == '@' {
			l.emit(tokens.ItemAt)
		} else if r == '~' {
			l.emit(tokens.ItemTilde)
		} else if '0' <= r && r <= '9' {
			l.backup()
			return lexNumber
		} else if isAlpha(r) {
			l.backup()
			return lexIdentifier
		} else if r == ';' {
			l.emit(tokens.ItemSemicolon)
			return lexInsideScope
		} else {
			return lexInsideScope
		}
	}
}

func lexInsideScope(l *lexer) stateFn {
	for {
		r := l.next()
		if isSpace(r) {
			l.ignore()
		} else if r == eof {
			l.emit(tokens.ItemEOF)
			return nil
		} else {
			l.backup()
			return lexInsideExpression
		}
	}
}

func lexNumber(l *lexer) stateFn {
	//Optional leading sign
	l.accept("+-")
	l.acceptRun("0123456789")

	if isAlphaNumeric(l.peek()) {
		l.next()
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	l.emit(tokens.ItemNumber)
	return lexInsideExpression
}

func lexIdentifier(l *lexer) stateFn {
	//We know the first is alphanumeric
	l.acceptRegex("0-9a-zA-Z_")

	current := l.input[l.start:l.pos]
	if current == "int" {
		l.emit(tokens.ItemKeyInt)
	} else if current == "bool" {
		l.emit(tokens.ItemKeyBool)
	} else if current == "false" {
		l.emit(tokens.ItemFalse)
	} else if current == "true" {
		l.emit(tokens.ItemTrue)
	} else if current == "func" {
		l.emit(tokens.ItemFunction)
	} else if current == "if" {
		l.emit(tokens.ItemIf)
	} else if current == "return" {
		l.emit(tokens.ItemReturn)
	} else if current == "else" {
		l.emit(tokens.ItemElse)
	} else if current == "for" {
		l.emit(tokens.ItemFor)
	} else {
		l.emit(tokens.ItemIdentifier)
	}

	return lexInsideExpression
}

func lexQuote(l *lexer) stateFn {
	for {
		r := l.next()
		if r == eof || r == '\n' {
			return l.errorf("Unterminated string literal")
		} else if r == '"' {
			l.emit(tokens.ItemText)
			return lexInsideExpression
		}
	}
}
