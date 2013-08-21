package main

import (
	"os"
	"fmt"
	"strconv"
	list "container/list"
)

type Puzzle struct {
	grid       [9][9]uint8
	candidates [9][9]uint16
}

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

type Stack struct {
	count uint
	nodes list.List
}

func (stack *Stack) Init() {
	stack.nodes.Init()
}

func (stack *Stack) Push(item Puzzle) {
	stack.nodes.PushBack(item)
	stack.count++
}

func (stack *Stack) Pop() Puzzle {
	stack.count--
	poped := stack.nodes.Back()
	value := stack.nodes.Remove(poped)
	return value
}

func resolve(puzzle Puzzle) Puzzle {
	var stack Stack
	stack.Init()
	stack.Push(puzzle)
	result := stack.Pop()
	return result
}

func main() {
	var puzzle Puzzle
	puzzle.InitFromFile("puzzle1")
	puzzle.Print()
	result := resolve(puzzle)
	fmt.Println("result")
	result.Print()
}
