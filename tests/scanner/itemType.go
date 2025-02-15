package scanner

import "fmt"

type itemType int

const (
	itemError itemType = iota
	itemInit
	itemDot
	itemEOF
	itemText
	itemLeftMeta
	itemRightMeta
	itemNumber
)

type item struct {
	itemType itemType
	value    string
}

func (i item) String() string {
	switch i.itemType {
	case itemEOF:
		return "EOF"
	case itemError:
		return i.value
	}

	if len(i.value) > 50 {
		return fmt.Sprintf("%d %s", int(i.itemType), i.value[:50])
	}

	return fmt.Sprintf("%d %q", int(i.itemType), i.value)
}
