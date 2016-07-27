package main

// go test -test.benchmem -test.bench Stack stack_test.go

import (
	"sync"
	"sync/atomic"
	"testing"
)

type stack struct {
	sync.Mutex

	c     chan struct{}
	top   int32
	items []int
}

func newStack(preallocSize int, bufSize int) *stack {
	return &stack{
		c:     make(chan struct{}, bufSize),
		items: make([]int, preallocSize),
	}
}

func (s *stack) Push(item int) {
	s.Lock()
	top := atomic.AddInt32(&s.top, 1)
	s.items[top-1] = item
	s.Unlock()
	s.c <- struct{}{}
}

func (s *stack) Pop() int {
	<-s.c
	s.Lock()
	top := atomic.AddInt32(&s.top, -1)
	v := s.items[top]
	s.Unlock()
	return v
}

func benchmarkStack(b *testing.B, concurrency int) {
	pushes := func(s *stack, count int) {
		for i := 0; i < count; i++ {
			s.Push(1)
		}
	}

	pops := func(s *stack, count int, done chan<- bool) {
		for i := 0; i < count; i++ {
			s.Pop()
		}
		done <- true
	}

	count := 1000
	s := newStack(count*concurrency, count*concurrency)
	done := make(chan bool, concurrency)
	for n := 0; n < b.N; n++ {
		for i := 0; i < concurrency; i++ {
			go pops(s, count, done)
			go pushes(s, count)
		}
		for i := 0; i < concurrency; i++ {
			<-done
		}
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
