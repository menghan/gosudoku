package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	pprof "runtime/pprof"
	"strconv"
	"sync"
)

var candidateCountTable [512]uint8

func init() {
	initCandidateCountTable()
}

func initCandidateCountTable() {
	for v := 0; v < 512; v++ {
		count := uint8(0)
		for i := uint(0); i < 9; i++ {
			if (v & (1 << i)) == 0 {
				count += 1
			}
		}
		candidateCountTable[v] = count
	}
}

func getCandidateCount(val uint16) uint8 {
	return candidateCountTable[val>>1]
}

type Puzzle struct {
	grid       [9][9]uint8
	candidates [9][9]uint16
	n_slot     uint8
}

func (puzzle *Puzzle) ReadFrom(f io.Reader) error {
	buf := make([]byte, 10)
	for x := 0; x < 9; x++ {
		n, err := f.Read(buf)
		if err != nil || n != len(buf) {
			return fmt.Errorf("read puzzle from file failed: %v", err)
		}
		for y := 0; y < 9; y++ {
			puzzle.candidates[x][y] = 0
			if buf[y] != '_' {
				n, _ := strconv.ParseUint(string(buf[y]), 10, 8)
				puzzle.grid[x][y] = uint8(n)
			}
		}
	}
	puzzle.n_slot = puzzle.Slotcount()
	for x := 0; x < 9; x++ {
		for y := 0; y < 9; y++ {
			puzzle.candidates[x][y] = puzzle.CalculateCandidates(x, y)
		}
	}
	return nil
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
	for i, v := range []uint16{2, 4, 8, 16, 32, 64, 128, 256, 512} {
		if bit&v == 0 {
			*result = append(*result, uint8(i+1))
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
				r++
			}
		}
	}
	return
}

type stack struct {
	lock *sync.Mutex
	cond *sync.Cond

	C     chan *Puzzle
	top   uint64
	items []*Puzzle
}

func newStack() *stack {
	s := &stack{
		lock:  &sync.Mutex{},
		items: make([]*Puzzle, 10240),
	}
	s.cond = sync.NewCond(s.lock)
	return s
}

func (s *stack) Push(item *Puzzle) {
	s.lock.Lock()
	s.items[s.top] = item
	s.top++
	s.lock.Unlock()
	s.cond.Signal()
}

func (s *stack) Pop() *Puzzle {
	s.lock.Lock()
	for s.top == 0 {
		s.cond.Wait()
	}
	s.top--
	v := s.items[s.top]
	s.items[s.top] = nil
	s.lock.Unlock()
	return v
}

type solver struct {
	sync.Mutex

	wg sync.WaitGroup

	stack    *stack
	syncPool *sync.Pool
	results  []*Puzzle
}

func newSolver(concurrency int) *solver {
	s := &solver{
		stack:    newStack(),
		syncPool: newPool(),
		results:  make([]*Puzzle, 0, 64),
	}
	for i := 0; i < concurrency; i++ {
		go s.workerSolve()
	}
	return s
}

func (s *solver) workerSolve() {
	candidatesResult := make([]uint8, 0, 9)
	for {
		current := s.stack.Pop()
		x, y := current.GetSlot()
		current.GetCandidates(&candidatesResult, x, y)
		for _, c := range candidatesResult {
			next := getPuzzle(s.syncPool)
			next.Reset(current)
			next.Set(x, y, c)
			if next.n_slot == 0 {
				s.Lock()
				s.results = append(s.results, next)
				s.Unlock()
			} else {
				s.wg.Add(1)
				s.stack.Push(next)
			}
		}
		putPuzzle(s.syncPool, current)
		s.wg.Done()
	}
}

func (s *solver) Solve(puzzle *Puzzle) []*Puzzle {
	s.results = s.results[:0]

	if puzzle.n_slot == 0 {
		s.results = append(s.results, puzzle)
		return s.results
	}

	// use copy version of input argument
	puzzleCopy := getPuzzle(s.syncPool)
	puzzleCopy.Reset(puzzle)

	s.wg.Add(1)
	s.stack.Push(puzzleCopy)

	s.wg.Wait()

	return s.results
}

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	count := flag.Int("count", 100, "calculation count")
	concurrency := flag.Int("concurrency", 1, "concurrency")
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
	var puzzle Puzzle
	err = puzzle.ReadFrom(file)
	if err != nil {
		log.Fatal(err)
	}
	puzzle.Print()

	solver := newSolver(*concurrency)
	for i := 0; i < *count; i++ {
		solver.Solve(&puzzle)
	}
	fmt.Println("result")
	for _, result := range solver.results {
		result.Print()
	}
}
