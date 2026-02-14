package lrucache

type Node struct {
	Prev, Next *Node
	key, Value int
}

func NewNode(key, val int) *Node {
	return &Node{
		key:   key,
		Value: val,
	}
}

type LRUCache struct {
	Head, Tail *Node
	cache      map[int]*Node
	capacity   int
}

func NewLRUCache(head, tail *Node, cap int) LRUCache {
	return LRUCache{
		Head:     head,
		Tail:     tail,
		cache:    make(map[int]*Node),
		capacity: cap,
	}
}

func Constructor(capacity int) LRUCache {
	head, tail := NewNode(0, 0), NewNode(0, 0)

	head.Next = tail
	tail.Prev = head
	return NewLRUCache(head, tail, capacity)
}

func (lru *LRUCache) remove(node *Node) {
	delete(lru.cache, node.key)
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
}

func (lru *LRUCache) insert(node *Node) {
	lru.cache[node.key] = node
	next := lru.Head.Next
	lru.Head.Next = node
	node.Prev = lru.Head
	next.Prev = node
	node.Next = next
}
func (lru *LRUCache) Get(key int) int {
	if n, ok := lru.cache[key]; ok {
		lru.remove(n)
		lru.insert(n)
		return n.Value
	}

	return -1
}

func (lru *LRUCache) Put(key int, value int) {
	if _, ok := lru.cache[key]; ok {
		lru.remove(lru.cache[key])
	}

	if len(lru.cache) == lru.capacity {
		lru.remove(lru.Tail.Prev)
	}

	lru.insert(NewNode(key, value))
}
