package http

import (
	"container/ring"
	"time"
)

type Speed struct {
	items  *ring.Ring
	last   time.Time
	change bool
	cache  int64
}
type Node struct {
	At    time.Time
	Value int64
}

func NewSpeed(n int) *Speed {
	if n < 1 || n > 60 {
		n = 10
	}
	return &Speed{
		items: ring.New(n),
	}
}
func (s *Speed) Get() int64 {
	if !s.change {
		now := time.Now()
		if now.Before(s.last.Add(time.Second)) {
			return s.cache
		}
		return s.get(now)
	}
	return s.get(time.Now())
}
func (s *Speed) get(now time.Time) int64 {
	var total int64
	s.items.Do(func(i interface{}) {
		if i == nil {
			return
		}
		node := i.(Node)
		if now.After(node.At.Add(time.Second * time.Duration(s.items.Len()))) {
			return
		}
		total += node.Value
	})
	s.cache = total / int64(s.items.Len())
	s.last = now
	s.change = false
	return s.cache
}
func (s *Speed) Push(value int64) {
	if value < 1 {
		return
	}
	s.change = true
	now := time.Now()
	v := s.items.Value
	if v != nil {
		node := v.(Node)
		if now.Before(node.At.Add(time.Second)) {
			node.Value += value
			s.items.Value = node
			return
		}
		s.items.Move(1)
	}
	s.items.Value = Node{
		At:    now,
		Value: value,
	}
}
