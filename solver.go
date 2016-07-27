package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
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

func (puzzle *Puzzle) LoadFromFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	err = puzzle.ReadFrom(file)
	if err != nil {
		return err
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

func (puzzle *Puzzle) Set(x, y int, value uint8) bool {
	if puzzle.grid[x][y] != 0 {
		panic("set value to non-zero slot!")
	}
	puzzle.n_slot -= 1
	puzzle.grid[x][y] = value
	or_value := uint16(1 << value)
	for xx := 0; xx < 9; xx++ {
		puzzle.candidates[xx][y] |= or_value
		if puzzle.candidates[xx][y] == 1022 && puzzle.grid[xx][y] == 0 {
			return false
		}
	}
	for yy := 0; yy < 9; yy++ {
		puzzle.candidates[x][yy] |= or_value
		if puzzle.candidates[x][yy] == 1022 && puzzle.grid[x][yy] == 0 {
			return false
		}
	}
	x_base := x / 3 * 3
	y_base := y / 3 * 3
	for xx := x_base; xx < x_base+3; xx++ {
		for yy := y_base; yy < y_base+3; yy++ {
			puzzle.candidates[xx][yy] |= or_value
			if puzzle.candidates[xx][yy] == 1022 && puzzle.grid[xx][yy] == 0 {
				return false
			}
		}
	}
	return true
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
	items []*Puzzle
}

func newStack(preallocSize int) *stack {
	return &stack{
		items: make([]*Puzzle, 0, preallocSize),
	}
}

func (s *stack) Push(item *Puzzle) {
	s.items = append(s.items, item)
}

func (s *stack) Pop() *Puzzle {
	l := len(s.items)
	v := s.items[l-1]
	s.items[l-1] = nil
	s.items = s.items[:l-1]
	return v
}

type solver struct {
	sync.Mutex
	cond *sync.Cond
	wg   sync.WaitGroup

	c chan *Puzzle

	syncPool    *sync.Pool
	concurrency int
	results     []*Puzzle
	GRstarted   chan struct{}
}

func newSolver(concurrency int) *solver {
	s := &solver{
		c:           make(chan *Puzzle, 64),
		cond:        sync.NewCond(&sync.Mutex{}),
		syncPool:    newPool(),
		concurrency: concurrency,
		results:     make([]*Puzzle, 0, 64),
		GRstarted:   make(chan struct{}, concurrency),
	}
	for i := 0; i < s.concurrency; i++ {
		go s.workerSolve()
	}
	for i := 0; i < s.concurrency; i++ {
		<-s.GRstarted
	}
	return s
}

func (s *solver) workerSolve() {
	candidatesResult := make([]uint8, 0, 9)
	stack := newStack(64)
	var current *Puzzle
	var working bool = false
	s.GRstarted <- struct{}{}

	for {
		if !working {
			s.cond.L.Lock()
			s.cond.Wait()
			s.cond.L.Unlock()
			working = true
		}
		if len(stack.items) != 0 {
			current = stack.Pop()
		} else {
			select {
			case current = <-s.c:
			default:
				s.wg.Done()
				working = false
				continue
			}
		}
		x, y := current.GetSlot()
		current.GetCandidates(&candidatesResult, x, y)
		for _, c := range candidatesResult {
			next := getPuzzle(s.syncPool)
			next.Reset(current)
			if !next.Set(x, y, c) {
				putPuzzle(s.syncPool, next)
				continue
			}
			if next.n_slot == 0 {
				s.Lock()
				s.results = append(s.results, next)
				s.Unlock()
				continue
			}
			stack.Push(next)
		}
		putPuzzle(s.syncPool, current)
	}
}

func (s *solver) Solve(puzzle *Puzzle) []*Puzzle {
	s.results = s.results[:0]

	if puzzle.n_slot == 0 {
		s.results = append(s.results, puzzle)
		return s.results
	}

	candidatesResult := make([]uint8, 0, 9)
	x, y := puzzle.GetSlot()
	puzzle.GetCandidates(&candidatesResult, x, y)
	for _, c := range candidatesResult {
		next := getPuzzle(s.syncPool)
		next.Reset(puzzle)
		if !next.Set(x, y, c) {
			putPuzzle(s.syncPool, next)
			continue
		}
		s.c <- next
	}
	s.wg.Add(s.concurrency)
	s.cond.L.Lock()
	s.cond.Broadcast()
	s.cond.L.Unlock()
	s.wg.Wait()
	return s.results
}

func main() {
	cpuprofile := flag.String("cpuprofile", "", "write cpu profile to file")
	blockprofile := flag.String("blockprofile", "", "write block profile to file")
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
	if *blockprofile != "" {
		runtime.SetBlockProfileRate(1)
	}

	var puzzle Puzzle
	err := puzzle.LoadFromFile(*puzzleFile)
	if err != nil {
		panic(err)
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
	if *blockprofile != "" {
		file, err := os.Create(*blockprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.Lookup("block").WriteTo(file, 1)
		file.Close()
	}
}
