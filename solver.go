package main

import "os"
import "fmt"

type Puzzle struct {
	grid [9][9]uint8
	candidates [9][9]uint16
}

func (puzzle Puzzle) InitFromFile(filename string) {
	input_file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer input_file.Close()
	buf := make([]byte, 10)
	for x := 0; x < 9; x ++ {
		n, _ := input_file.Read(buf)
		if n == 0 {
			break
		}
		for y := 0; y < 9; y++ {
			puzzle.candidates[x][y] = 0
			if buf[y] == '1' {
				puzzle.grid[x][y] = 1
			} else if buf[y] == '2' {
				puzzle.grid[x][y] = 2
			} else if buf[y] == '3' {
				puzzle.grid[x][y] = 3
			} else if buf[y] == '4' {
				puzzle.grid[x][y] = 4
			} else if buf[y] == '5' {
				puzzle.grid[x][y] = 5
			} else if buf[y] == '6' {
				puzzle.grid[x][y] = 6
			} else if buf[y] == '7' {
				puzzle.grid[x][y] = 7
			} else if buf[y] == '8' {
				puzzle.grid[x][y] = 8
			} else if buf[y] == '9' {
				puzzle.grid[x][y] = 9
			}
			// value is 0(default) if buf[y] == '_'
		}
	}
}

func (puzzle Puzzle) Print() {
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

func main() {
	var puzzle Puzzle
	puzzle.InitFromFile("puzzle1")
	puzzle.Print()
}
