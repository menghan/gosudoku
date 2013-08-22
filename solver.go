package main

import (
	list "container/list"
	"flag"
	"fmt"
	"os"
	pprof "runtime/pprof"
	"strconv"
)

type Puzzle struct {
	grid       [9][9]uint8
	candidates [9][9]uint16
	n_slot     uint8
}

// TODO: don't create then init, create directly from file, use factory method
func (puzzle *Puzzle) InitFromFile(filename string) {
	input_file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer input_file.Close()
	buf := make([]byte, 10)
	for x := 0; x < 9; x++ {
		n, _ := input_file.Read(buf)
		if n == 0 {
			break
		}
		for y := 0; y < 9; y++ {
			puzzle.candidates[x][y] = 0
			if buf[y] != '_' {
				n, _ := strconv.ParseUint(string(buf[y]), 10, 8)
				puzzle.grid[x][y] = uint8(n)
			}
		}
	}
	puzzle.n_slot = puzzle.slotcount()
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
	var min_cdd uint = 10 // 9 is the largest candidates count
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
	var i uint8 = 1
	for ; i < 10; i++ {
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
	var x_base int = x / 3 * 3
	var y_base int = y / 3 * 3
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
	var i uint = 1
	for ; i < 10; i++ {
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
	var x_base int = x / 3 * 3
	var y_base int = y / 3 * 3
	for xx := x_base; xx < x_base+3; xx++ {
		for yy := y_base; yy < y_base+3; yy++ {
			puzzle.candidates[xx][yy] = 0
		}
	}
}

func (puzzle *Puzzle) slotcount() (r uint8) {
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

func (stack *Stack) Init() {
	stack.nodes.Init()
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

func resolve(puzzle Puzzle) []Puzzle {
	var stack Stack
	var results []Puzzle
	stack.Init()
	stack.Push(puzzle)
	for stack.count != 0 {
		current, ok := stack.Pop().(Puzzle)
		if ! ok {
			fmt.Println("Pop invalid")
			os.Exit(1)
		}
		x, y := current.GetSlot()
		candidates := current.GetCandidates(x, y)
		for _, c := range candidates {
			next := Puzzle(current)
			next.Set(x, y, c)
			if next.n_slot == 0 {
				results = append(results, next)
			} else {
				stack.Push(next)
			}
		}
	}
	return results
}

func main() {
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	var puzzle Puzzle
	puzzle.InitFromFile("puzzle4")
	puzzle.Print()
	results := resolve(puzzle)
	fmt.Println("result")
	for _, result := range results {
		result.Print()
		fmt.Println()
	}
}
