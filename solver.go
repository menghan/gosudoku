package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	pprof "runtime/pprof"
	"strconv"
	"sync"
)

var candidateCountTable [512]uint8
var initCandidateCountOnce sync.Once

func getCandidateCount(val uint16) uint8 {
	initCandidateCountOnce.Do(func() {
		for v := 0; v < 512; v++ {
			count := uint8(0)
			for i := uint(0); i < 9; i++ {
				if (v & (1 << i)) == 0 {
					count += 1
				}
			}
			candidateCountTable[v] = count
		}
	})
	return candidateCountTable[val>>1]
}

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
	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			p.candidates[x][y] = p.CalculateCandidates(x, y)
		}
	}
	return p, nil
}

func newPuzzle() interface{} {
	return &Puzzle{}
}

func newPool() *sync.Pool {
	return &sync.Pool{
		New: newPuzzle,
	}
}

func getPuzzle(pool *sync.Pool) *Puzzle {
	return pool.Get().(*Puzzle)
}

func putPuzzle(pool *sync.Pool, puzzle *Puzzle) {
	pool.Put(puzzle)
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

func (puzzle *Puzzle) Reset(other *Puzzle) {
	puzzle.grid = other.grid
	puzzle.candidates = other.candidates
	puzzle.n_slot = other.n_slot
}

func (puzzle *Puzzle) GetSlot() (rx, ry int) {
	min_cdd := uint8(10) // 9 is the largest candidates count
	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			if puzzle.grid[x][y] != 0 {
				continue
			}
			cdd := getCandidateCount(puzzle.candidates[x][y])
			if cdd < min_cdd {
				rx, ry = x, y
				if cdd == 1 {
					return
				}
				min_cdd = cdd
			}
		}
	}
	return
}

func (puzzle *Puzzle) GetCandidates(result *[]uint8, x, y int) {
	*result = (*result)[:0]
	bit := puzzle.candidates[x][y]
	for i := uint8(1); i < 10; i++ {
		if (bit & (1 << i)) == 0 {
			*result = append(*result, i)
		}
	}
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

func (puzzle *Puzzle) Set(x, y int, value uint8) {
	if puzzle.grid[x][y] != 0 {
		panic("set value to non-zero slot!")
	}
	puzzle.n_slot -= 1
	puzzle.grid[x][y] = value
	or_value := uint16(1 << value)
	for xx := 0; xx < 9; xx++ {
		puzzle.candidates[xx][y] |= or_value
	}
	for yy := 0; yy < 9; yy++ {
		puzzle.candidates[x][yy] |= or_value
	}
	x_base := x / 3 * 3
	y_base := y / 3 * 3
	for xx := x_base; xx < x_base+3; xx++ {
		for yy := y_base; yy < y_base+3; yy++ {
			puzzle.candidates[xx][yy] |= or_value
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

type stack struct {
	top   uint
	items []interface{}
}

func newStack() *stack {
	return &stack{
		items: make([]interface{}, 10240),
	}
}

func (s *stack) Push(item interface{}) {
	s.items[s.top] = item
	s.top++
}

func (s *stack) Pop() interface{} {
	if s.top == 0 {
		panic("stackunderflow!")
	}
	s.top--
	v := s.items[s.top]
	s.items[s.top] = nil
	return v
}

func resolve(puzzle *Puzzle) []*Puzzle {
	results := make([]*Puzzle, 0, 1024)
	if puzzle.n_slot == 0 {
		results = append(results, puzzle)
		return results
	}

	// use copy version of input argument
	syncPool := newPool()
	puzzleCopy := getPuzzle(syncPool)
	puzzleCopy.Reset(puzzle)

	candidatesResult := make([]uint8, 0, 9)

	stack := newStack()
	stack.Push(puzzleCopy)
	for stack.top != 0 {
		current, ok := stack.Pop().(*Puzzle)
		if !ok {
			log.Fatal("Pop invalid")
		}
		x, y := current.GetSlot()
		current.GetCandidates(&candidatesResult, x, y)
		for _, c := range candidatesResult {
			next := getPuzzle(syncPool)
			next.Reset(current)
			next.Set(x, y, c)
			if next.n_slot == 0 {
				results = append(results, next)
			} else {
				stack.Push(next)
			}
		}
		putPuzzle(syncPool, current)
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
