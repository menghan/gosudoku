package main

import (
	list "container/list"
	"flag"
	"fmt"
	"log"
	"os"
	pprof "runtime/pprof"
	"strconv"
)

type Puzzle struct {
	grid       [9][9]uint8
	candidates [9][9]uint16
	n_slot     uint8
}

func NewPuzzle(f *os.File) (*Puzzle, error) {
	buf := make([]byte, 10)
	candidates := [9][9]uint16{}
	grid := [9][9]uint8{}
	for x := 0; x < 9; x++ {
		n, err := f.Read(buf)
		if err != nil || n != len(buf) {
			return nil, fmt.Errorf("read puzzle from file failed: %v", err)
		}
		for y := 0; y < 9; y++ {
			candidates[x][y] = 0
			if buf[y] != '_' {
				n, _ := strconv.ParseUint(string(buf[y]), 10, 8)
				grid[x][y] = uint8(n)
			}
		}
	}
	p := &Puzzle{
		candidates: candidates,
		grid:       grid,
	}
	p.n_slot = p.Slotcount()
	return p, nil
}

func (puzzle *Puzzle) Copy() *Puzzle {
	return &Puzzle{
		candidates: puzzle.candidates,
		grid:       puzzle.grid,
		n_slot:     puzzle.n_slot,
	}
}

func (puzzle *Puzzle) Print() {
	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			if puzzle.grid[x][y] == 0 {
				fmt.Printf("_")
			} else {
				fmt.Printf("%d", puzzle.grid[x][y])
			}
		}
		fmt.Printf("\n")
	}
}

func (puzzle *Puzzle) GetSlot() (rx, ry int) {
	min_cdd := uint(10) // 9 is the largest candidates count
	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			if puzzle.grid[x][y] != 0 {
				continue
			}
			cdd := puzzle.getCandidateCount(x, y)
			if cdd < min_cdd {
				rx, ry = x, y
				min_cdd = cdd
				if min_cdd == 1 {
					return
				}
			}
		}
	}
	return
}

func (puzzle *Puzzle) GetCandidates(x, y int) []uint8 {
	if puzzle.candidates[x][y] == 0 {
		puzzle.candidates[x][y] = puzzle.CalculateCandidates(x, y)
	}
	candidates := make([]uint8, 0, 9)
	bit := puzzle.candidates[x][y]
	for i := uint8(1); i < 10; i++ {
		if (bit & (1 << i)) == 0 {
			candidates = append(candidates, i)
		}
	}
	return candidates
}

func (puzzle *Puzzle) CalculateCandidates(x, y int) (bit uint16) {
	for xx := 0; xx < 9; xx++ {
		bit |= 1 << puzzle.grid[xx][y]
	}
	for yy := 0; yy < 9; yy++ {
		bit |= 1 << puzzle.grid[x][yy]
	}
	x_base := x / 3 * 3
	y_base := y / 3 * 3
	for xx := x_base; xx < x_base+3; xx++ {
		for yy := y_base; yy < y_base+3; yy++ {
			bit |= 1 << puzzle.grid[xx][yy]
		}
	}
	bit &= 0x3FE // FIXME: hardcode
	return
}

func (puzzle *Puzzle) getCandidateCount(x, y int) (count uint) {
	if puzzle.candidates[x][y] == 0 {
		puzzle.candidates[x][y] = puzzle.CalculateCandidates(x, y)
	}
	bit := puzzle.candidates[x][y]
	for i := uint(1); i < 10; i++ {
		if (bit & (1 << i)) == 0 {
			count += 1
		}
	}
	return
}

func (puzzle *Puzzle) Set(x, y int, value uint8) {
	if puzzle.grid[x][y] == 0 {
		puzzle.n_slot -= 1
	}
	puzzle.grid[x][y] = value
	for xx := 0; xx < 9; xx++ {
		puzzle.candidates[xx][y] = 0
	}
	for yy := 0; yy < 9; yy++ {
		puzzle.candidates[x][yy] = 0
	}
	x_base := x / 3 * 3
	y_base := y / 3 * 3
	for xx := x_base; xx < x_base+3; xx++ {
		for yy := y_base; yy < y_base+3; yy++ {
			puzzle.candidates[xx][yy] = 0
		}
	}
}

func (puzzle *Puzzle) Slotcount() (r uint8) {
	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			if puzzle.grid[x][y] == 0 {
				r += 1
			}
		}
	}
	return
}

type Stack struct {
	count uint
	nodes list.List
}

func NewStack() *Stack {
	s := &Stack{}
	s.nodes.Init()
	return s
}

func (stack *Stack) Push(item interface{}) {
	stack.nodes.PushBack(item)
	stack.count++
}

func (stack *Stack) Pop() interface{} {
	stack.count--
	poped := stack.nodes.Back()
	value := stack.nodes.Remove(poped)
	return value
}

func resolve(puzzle *Puzzle) []*Puzzle {
	stack := NewStack()
	results := make([]*Puzzle, 0, 1024)
	stack.Push(puzzle)
	for stack.count != 0 {
		current, ok := stack.Pop().(*Puzzle)
		if !ok {
			log.Fatal("Pop invalid")
		}
		x, y := current.GetSlot()
		candidates := current.GetCandidates(x, y)
		for _, c := range candidates {
			next := Puzzle(*current)
			next.Set(x, y, c)
			if next.n_slot == 0 {
				results = append(results, &next)
			} else {
				stack.Push(&next)
			}
		}
	}
	return results
}

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	count := flag.Int("count", 100, "calculation count")
	puzzleFile := flag.String("file", "", "target puzzle file")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}
	file, err := os.Open(*puzzleFile)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	puzzle, err := NewPuzzle(file)
	if err != nil {
		log.Fatal(err)
	}
	puzzle.Print()
	var results []*Puzzle
	for i := 0; i < *count; i++ {
		results = resolve(puzzle)
	}
	fmt.Println("result")
	for _, result := range results {
		result.Print()
		fmt.Println()
	}
}
