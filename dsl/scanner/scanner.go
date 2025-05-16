package scanner

import (
	"autofat/dsl/tokens"
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
			l.ignore() //Ignore the initial quote
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
			if l.peek() == '/' {
				l.ignore()
				return lexComment
			} else {
				l.emit(tokens.ItemOpDiv)
			}
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
	} else if current == "string" {
		l.emit(tokens.ItemKeyString)
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
			l.backup() // Remove the trailing quote
			l.emit(tokens.ItemText)
			//Then skip over the trailing quote again.
			l.next()
			l.ignore()
			return lexInsideExpression
		}
	}
}

func lexComment(l *lexer) stateFn {
	if l.next() == '\n' {
		l.ignore()
		return lexInsideScope
	}
	return lexComment
}
