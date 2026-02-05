package datastructures

type Node struct {
	Value int
	Prev  *Node
	Next  *Node
}

type DoublyLinkedList struct {
	Head *Node
	Tail *Node
	Size int
}

func NewDoublyLinkedList() *DoublyLinkedList {
	return &DoublyLinkedList{
		Head: nil,
		Tail: nil,
		Size: 0,
	}
}

func (dll *DoublyLinkedList) AddToFront(value int) {
	newNode := &Node{Value: value}

	if dll.Head == nil {
		dll.Head = newNode
		dll.Tail = newNode
	} else {
		newNode.Next = dll.Head
		dll.Head.Prev = newNode
		dll.Head = newNode
	}
	dll.Size++
}

func (dll *DoublyLinkedList) AddToBack(value int) {
	newNode := &Node{Value: value}

	if dll.Tail == nil {
		dll.Tail = newNode
		dll.Head = newNode
	} else {
		newNode.Prev = dll.Tail
		dll.Tail.Next = newNode
		dll.Tail = newNode
	}
	dll.Size++
}
