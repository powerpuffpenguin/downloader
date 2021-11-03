package http

import (
	"time"
)

type Node struct {
	At    time.Time
	Value int64
}
type FIFO struct {
	items         []Node
	current, size int
	total         int64
}

func NewFiFO(n int) *FIFO {
	if n < 1 || n > 60 {
		n = 10
	}
	return &FIFO{
		items: make([]Node, n),
	}
}
func (f *FIFO) Get() (total int64, count int64) {
	now := time.Now()
	size := f.size
	for i := 0; i < size; i++ {
		index := f.current + 1 - size
		if index < 0 {
			index += len(f.items)
		}
		node := f.items[index]

		if now.After(node.At.Add(time.Second * time.Duration(len(f.items)))) {
			f.total -= node.Value
			f.size--
			continue
		}
		count++
	}
	total = f.total
	return
}

func (f *FIFO) Push(value int64) {
	if value < 0 {
		return
	}

	f.total += value
	now := time.Now()

	if f.size == 0 {
		f.size = 1
		f.items[f.current].Value = value
		f.items[f.current].At = now
		return
	}

	if now.After(f.items[f.current].At.Add(time.Second)) {
		f.current = (f.current + 1) % len(f.items)
		if f.size != len(f.items) {
			f.size++
		}
		f.items[f.current].Value = value
		f.items[f.current].At = now
		return
	}

	f.items[f.current].Value += value
}
