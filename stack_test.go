package stack_test

// go test -test.benchmem -test.bench Stack stack_test.go

import (
	"sync"
	"testing"
)

type stack struct {
	lock *sync.Mutex
	cond *sync.Cond

	items []int
}

func newStack(preallocSize int) *stack {
	var l sync.Mutex
	return &stack{
		lock:  &l,
		cond:  sync.NewCond(&l),
		items: make([]int, 0, preallocSize),
	}
}

func (s *stack) Push(item int) {
	s.lock.Lock()
	s.items = append(s.items, item)
	s.lock.Unlock()
	s.cond.Signal()
}

func (s *stack) Pop() int {
	s.lock.Lock()
	for len(s.items) == 0 {
		s.cond.Wait()
	}
	l := len(s.items)
	v := s.items[l-1]
	s.items = s.items[:l-1]
	s.lock.Unlock()
	return v
}

func benchmarkStack(b *testing.B, concurrency int) {
	pushes := func(s *stack, count int) {
		for i := 0; i < count; i++ {
			s.Push(1)
		}
	}

	pops := func(s *stack, count int, wg *sync.WaitGroup) {
		for i := 0; i < count; i++ {
			s.Pop()
		}
		wg.Done()
	}

	count := 1000
	s := newStack(count * concurrency)
	wg := &sync.WaitGroup{}
	for n := 0; n < b.N; n++ {
		wg.Add(concurrency)
		for i := 0; i < concurrency; i++ {
			go pops(s, count, wg)
		}
		pushes(s, count*concurrency)
		wg.Wait()
	}
}

func BenchmarkStack1(b *testing.B) {
	benchmarkStack(b, 1)
}

func BenchmarkStack2(b *testing.B) {
	benchmarkStack(b, 2)
}

func BenchmarkStack4(b *testing.B) {
	benchmarkStack(b, 4)
}
