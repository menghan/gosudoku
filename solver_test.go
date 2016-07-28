package main

// go test -test.benchmem -test.bench . solver_test.go solver.go

import (
	"testing"
)

func benchmarkSolver(b *testing.B, puzzleFilename string, expectResultCount int, concurrency int) {
	var puzzle Puzzle
	puzzle.MustLoadFromFile(puzzleFilename)
	solver := newSolver(concurrency)
	for n := 0; n < b.N; n++ {
		solver.Solve(&puzzle)
	}
	if len(solver.results) != expectResultCount {
		b.Fatalf("Wrong result count, expect %d, got %d", expectResultCount, len(solver.results))
	}
}

func benchmarkSmallSolver(b *testing.B, concurrency int) {
	benchmarkSolver(b, "puzzle4", 1, concurrency)
}

func benchmarkBigSolver(b *testing.B, concurrency int) {
	benchmarkSolver(b, "puzzle6", 5122, concurrency)
}

func BenchmarkSmallSolver1(b *testing.B) {
	benchmarkSmallSolver(b, 1)
}

func BenchmarkSmallSolver2(b *testing.B) {
	benchmarkSmallSolver(b, 2)
}

func BenchmarkSmallSolver4(b *testing.B) {
	benchmarkSmallSolver(b, 4)
}

func BenchmarkBigSolver1(b *testing.B) {
	benchmarkBigSolver(b, 1)
}

func BenchmarkBigSolver2(b *testing.B) {
	benchmarkBigSolver(b, 2)
}

func BenchmarkBigSolver4(b *testing.B) {
	benchmarkBigSolver(b, 4)
}
