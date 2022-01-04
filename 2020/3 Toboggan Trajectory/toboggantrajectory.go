package main

import (
	"bufio"
	// "constraints"
	"errors"
	"fmt"
	"io"
	"os"
	// "strconv"
	"strings"
	// "time"

	// "../../slices"
	"../../streams"
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

type Point struct {
	X, Y int
}

type Grid struct {
	Length, Width int
	Trees map[Point]bool
}

func loadGrid() *Grid {
	lines := readInputLines()
	trees := map[Point]bool{}
	for j, line := range lines {
		for i, location := range line {
			if location == '#' {
				trees[Point{i, j}] = true
			}
		}
	}
	return &Grid{Length: len(lines), Width: len(lines[0]), Trees: trees}
}

func (g *Grid) CountTrees(slope Point) int {
	treesEncountered := 0

	at := Point{0, 0}
	for at.Y < g.Length {
		if g.Trees[at] {
			treesEncountered++
		}
		at.Y += slope.Y
		at.X = (at.X + slope.X) % g.Width
	}

	return treesEncountered
}

func main() {
	grid := loadGrid()

	fmt.Printf("Trees encountered from (0, 0) with slope (3, 1): %d\n", grid.CountTrees(Point{3, 1}))

	product := streams.Reduce(
		streams.Map(
			streams.FromItems([]Point{{1, 1}, {3, 1}, {5, 1}, {7, 1}, {1, 2}}...),
			func(slope Point) int { return grid.CountTrees(slope) },
		),
		1,
		func(a, b int) int { return a * b },
	)
	fmt.Printf("Product of number of trees encountered on each of the listed slopes: %d\n", product)
}
