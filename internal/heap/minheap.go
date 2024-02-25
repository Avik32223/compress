package heap

import "fmt"

type MinHeap struct {
	nodes []Node
}

func NewMinHeap() *MinHeap {
	return &MinHeap{
		nodes: make([]Node, 0),
	}
}

func (m *MinHeap) Empty() bool {
	return m.Size() == 0
}

func (m *MinHeap) Size() int {
	return len(m.nodes)
}

func (m *MinHeap) Insert(h Node) int {
	m.nodes = append(m.nodes, h)
	m.heapifyUp(len(m.nodes) - 1)
	return len(m.nodes)
}

func (m *MinHeap) Extract() (Node, error) {
	if len(m.nodes) == 0 {
		return nil, fmt.Errorf("empty heap")
	}

	x := m.nodes[0]
	t := len(m.nodes) - 1
	m.nodes[0] = m.nodes[t]
	m.nodes = m.nodes[:t]

	m.heapifyDown(0)

	return x, nil
}

func (m *MinHeap) heapifyUp(idx int) {
	for m.nodes[m.parent(idx)].Val() > m.nodes[idx].Val() {
		m.swap(m.parent(idx), idx)
		idx = m.parent(idx)
	}
}

func (m *MinHeap) heapifyDown(idx int) {
	t := len(m.nodes) - 1
	l, r := m.left(idx), m.right(idx)

	for l <= t {
		var c int
		if l == t {
			c = l
		} else if m.nodes[l].Val() < m.nodes[r].Val() {
			c = l
		} else {
			c = r
		}
		if m.nodes[idx].Val() > m.nodes[c].Val() {
			m.swap(idx, c)
			idx = c
			l, r = m.left(idx), m.right(idx)
		} else {
			return
		}
	}
}

func (m *MinHeap) parent(idx int) int {
	return idx / 2
}

func (m *MinHeap) left(idx int) int {
	return idx*2 + 1
}

func (m *MinHeap) right(idx int) int {
	return idx * 2
}

func (m *MinHeap) swap(i, j int) {
	m.nodes[i], m.nodes[j] = m.nodes[j], m.nodes[i]
}
