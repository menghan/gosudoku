package main

// go test -test.benchmem -test.bench . solver_test.go solver.go

import (
	"testing"
)

func benchmarkSolver(b *testing.B, puzzleFilename string, concurrency int) {
	var puzzle Puzzle
	if err := puzzle.LoadFromFile(puzzleFilename); err != nil {
		b.Fatal(err)
	}
	solver := newSolver(concurrency)
	for n := 0; n < b.N; n++ {
		solver.Solve(&puzzle)
	}
}

func BenchmarkSolver1(b *testing.B) {
	benchmarkSolver(b, "puzzle4", 1)
}

func BenchmarkSolver2(b *testing.B) {
	benchmarkSolver(b, "puzzle4", 2)
}

func BenchmarkSolver4(b *testing.B) {
	benchmarkSolver(b, "puzzle4", 4)
}
