package scanner

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type lexer struct {
	name  string
	input string
	start int
	pos   int
	width int
	items chan item
}

const eof rune = '\x00' // necessary in 2025?

type stateFn func(*lexer) stateFn

func lex(name string, input string) (*lexer, chan item) {
	l := &lexer{
		name:  name,
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l, l.items
}

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.items)
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.input[l.start:l.pos]}
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

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	return nil
}

const leftMeta string = "{{"
const rightMeta string = "}}"

func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], leftMeta) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexLeftMeta
		}
		if l.next() == eof {
			break
		}
	}

	//Correctly reached EOF.
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil
}

func lexLeftMeta(l *lexer) stateFn {
	l.pos += len(leftMeta)
	l.emit(itemLeftMeta)
	return lexInsideAction
}

func lexRightMeta(l *lexer) stateFn {
	l.pos += len(rightMeta)
	l.emit(itemRightMeta)
	return lexText
}

func isAlphaNumeric(c rune) bool {
	return ('a' <= c && c <= 'z') || ('A' <= c && c <= 'Z') || ('0' <= c && c <= '9')
}

func isSpace(c rune) bool {
	return c == ' ' || c == '\t'
}

func lexInsideAction(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], rightMeta) {
			return lexRightMeta
		}

		r := l.next()

		if r == eof || r == '\n' {
			return l.errorf("unclosed action")
		} else if isSpace(r) {
			l.ignore()
		} else if r == '"' {
			return l.errorf("unimplemented")
			//return lexQuote
		} else if r == '`' {
			return l.errorf("unimplemented")
			//return lexRawQuote
		} else if r == '+' || r == '-' || '0' <= r && r <= '9' {
			l.backup()
			return lexNumber
		} else if isAlphaNumeric(r) {
			return l.errorf("unimplemented")
			//return lexIdentifier
		}
	}

}

func lexNumber(l *lexer) stateFn {
	//Optional leading sign
	fmt.Printf("\"%s\"", l.input[l.start:l.pos])
	l.accept("+-")

	fmt.Printf("\"%s\"", l.input[l.start:l.pos])
	l.acceptRun("0123456789")

	fmt.Printf("\"%s\"", l.input[l.start:l.pos])

	if isAlphaNumeric(l.peek()) {
		l.next()
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	l.emit(itemNumber)
	return lexInsideAction
}
