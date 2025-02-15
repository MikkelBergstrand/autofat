package scanner

import (
	"fmt"
	"os"
	"testing"
)

func TestLexer(t *testing.T) {
	file_contents, err := os.ReadFile("testfiles/01.txt")
	if err != nil {
		t.Fatal(err)
		return
	}

	fmt.Println("Homie")
	_, ch := lex("test_lexer", string(file_contents))

	for {
		c := <-ch
		switch c.itemType {
		case itemEOF:
			return
		case itemError:
			t.Fatal(c)
			return
		default:
			fmt.Println(c)
		}
	}
}
