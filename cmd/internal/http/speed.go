package http

import "fmt"

type Speed struct {
	fifo   *FIFO
	offset int64
}

func NewSpeed(n int) *Speed {
	return &Speed{
		fifo: NewFiFO(n),
	}
}

func (s *Speed) Get() int64 {
	total, count := s.fifo.Get()
	fmt.Println(total, count)
	if count != 0 {
		return total / count
	}
	return 0
}
func (s *Speed) Push(offset int64) {
	if offset > s.offset {
		s.fifo.Push(offset - s.offset)
		s.offset = offset
	}
}
