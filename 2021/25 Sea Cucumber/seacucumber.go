package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

const inputFilename = "./INPUT"

func readInputLines() []string {
	inputFile, err := os.Open(inputFilename)
	if err != nil {
		panic(err)
	}
	defer inputFile.Close()

	var (
		line string
		lineNum int
		lines []string
	)

	bufReader := bufio.NewReader(inputFile)
	for {
		lineNum++
		line, err = bufReader.ReadString('\n')
		if err != nil {
			err = fmt.Errorf("unable to read line %d: %w", lineNum, err)
			break
		}
		lines = append(lines, strings.TrimSpace(line))
	}
	if !errors.Is(err, io.EOF) {
		panic(err)
	}

	return lines
}

type Grid struct {
	Cells [][]GridCell
	Length, Width int
	MoveBuf []*GridCell
}

func (g *Grid) Print() {
	for j := 0; j < g.Length; j++ {
		for i := 0; i < g.Width; i++ {
			cell := g.Cells[j][i]
			switch {
			case !cell.IsOccupied:
				fmt.Printf(".")
			case cell.IsEastFacing:
				fmt.Printf(">")
			default:
				fmt.Printf("v")
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

type GridCell struct {
	X, Y int
	IsOccupied bool
	IsEastFacing bool
}

func (c *GridCell) CanMove(g *Grid, eastward bool) bool {
	if !c.IsOccupied {
		return false
	}
	var targetCell *GridCell
	if eastward && c.IsEastFacing {
		// Try to move to X = c.X+1 % g.Width
		targetCell = &g.Cells[c.Y][(c.X+1) % g.Width]
	} else if !eastward && !c.IsEastFacing {
		// Try to move to Y = c.Y+1 % g.Length
		targetCell = &g.Cells[(c.Y+1) % g.Length][c.X]
	}
	if targetCell == nil || targetCell.IsOccupied {
		return false // Can't move.
	}
	// Can move!
	return true
}

func (c *GridCell) Move(g *Grid) {
	var targetCell *GridCell
	if c.IsEastFacing {
		// Try to move to X = c.X+1 % g.Width
		targetCell = &g.Cells[c.Y][(c.X+1) % g.Width]
	} else {
		// Try to move to Y = c.Y+1 % g.Length
		targetCell = &g.Cells[(c.Y+1) % g.Length][c.X]
	}
	// fmt.Printf("Moved eastward=%t (%d, %d) to (%d, %d)\n", c.IsEastFacing, c.X, c.Y, targetCell.X, targetCell.Y)
	targetCell.IsOccupied = true
	targetCell.IsEastFacing = c.IsEastFacing
	c.IsOccupied = false
	c.IsEastFacing = false
}

func (g *Grid) Step() bool {
	var numMoves int
	g.MoveBuf = g.MoveBuf[:0]
	// First iterate across every row from west to east to try to move each cucumber if possible.
	for j := 0; j < g.Length; j++ {
		for i := 0; i < g.Width; i++ {
			if g.Cells[j][i].CanMove(g, true) {
				g.MoveBuf = append(g.MoveBuf, &g.Cells[j][i])
			}
		}
	}

	for _, cell := range g.MoveBuf {
		cell.Move(g)
		numMoves++
	}

	g.MoveBuf = g.MoveBuf[:0]
	// Next iterate across every column fro north to south to try to move each cucumber if possible.
	for i := 0; i < g.Width; i++ {
		for j := 0; j < g.Length; j++ {
			if g.Cells[j][i].CanMove(g, false) {
				g.MoveBuf = append(g.MoveBuf, &g.Cells[j][i])
			}
		}
	}

	for _, cell := range g.MoveBuf {
		cell.Move(g)
		numMoves++
	}

	return numMoves > 0
}

func loadGrid() *Grid {
	lines := readInputLines()
	grid := &Grid{
		Length: len(lines),
		Width:  len(lines[0]),
	}
	grid.MoveBuf = make([]*GridCell, 0, grid.Length*grid.Width)
	grid.Cells = make([][]GridCell, grid.Length)
	for j, line := range lines {
		grid.Cells[j] = make([]GridCell, grid.Width)
		for i, rawCell := range strings.Split(line, "") {
			var isOccupied, isEastFacing bool
			switch rawCell {
			case ">":
				isEastFacing = true
				fallthrough
			case "v":
				isOccupied = true
			}
			grid.Cells[j][i] = GridCell{
				X: i, Y: j,
				IsOccupied: isOccupied,
				IsEastFacing: isEastFacing,
			}
		}
	}
	return grid
}

func main() {
	grid := loadGrid()
	grid.Print()

	step := 1
	for grid.Step() {
		// fmt.Printf("After %d steps:\n", step)
		// grid.Print()
		step++
	}
	// grid.Print()
	fmt.Printf("First step on which no sea cucumbers move: %d\n", step)
}
