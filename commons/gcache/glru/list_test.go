package glru

import (
	"testing"
)

func Test_List_Insert(t *testing.T) {
	l := NewList[int]()
	assertList(t, l)

	l.Insert(1)
	assertList(t, l, 1)

	l.Insert(2)
	assertList(t, l, 2, 1)

	l.Insert(3)
	assertList(t, l, 3, 2, 1)
}

func Test_List_Remove(t *testing.T) {
	l := NewList[int]()
	assertList(t, l)

	node := l.Insert(1)
	l.Remove(node)
	assertList(t, l)

	n5 := l.Insert(5)
	n4 := l.Insert(4)
	n3 := l.Insert(3)
	n2 := l.Insert(2)
	n1 := l.Insert(1)

	l.Remove(n5)
	assertList(t, l, 1, 2, 3, 4)

	l.Remove(n1)
	assertList(t, l, 2, 3, 4)

	l.Remove(n3)
	assertList(t, l, 2, 4)

	l.Remove(n2)
	assertList(t, l, 4)

	l.Remove(n4)
	assertList(t, l)
}

func Test_List_MoveToFront(t *testing.T) {
	l := NewList[int]()

	n1 := l.Insert(1)
	l.MoveToFront(n1)
	assertList(t, l, 1)

	n2 := l.Insert(2)
	l.MoveToFront(n1)
	assertList(t, l, 1, 2)
	l.MoveToFront(n2)
	assertList(t, l, 2, 1)
}

func assertList(t *testing.T, list *List[int], expected ...int) {
	t.Helper()

	if len(expected) == 0 {
		Nil(t, list.Head)
		Nil(t, list.Tail)
		return
	}

	node := list.Head
	for _, expected := range expected {
		Equal(t, node.Value, expected)
		node = node.Next
	}

	node = list.Tail
	for i := len(expected) - 1; i >= 0; i-- {
		Equal(t, node.Value, expected[i])
		node = node.Prev
	}
}
