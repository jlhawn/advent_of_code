package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	// "sort"
	"strconv"
	"strings"
)

const inputFilename = "./INPUT"

func readInputLines() ([]string, error) {
	inputFile, err := os.Open(inputFilename)
	if err != nil {
		return nil, fmt.Errorf("unable to open input file: %w", err)
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
		return nil, fmt.Errorf("unexpected error reading input: %w", err)
	}

	return lines, nil
}

func loadGrid() (*Grid, error) {
	rawLines, err := readInputLines()
	if err != nil {
		return nil, fmt.Errorf("unable to read lines: %w", err)
	}

	octos := map[Point]*Octopus{}
	for j, rawLine := range rawLines {
		rawInts := strings.Split(rawLine, "")
		if len(rawInts) != len(rawLines) {
			return nil, fmt.Errorf("expected %d ints but got %d", len(rawLines), len(rawInts))
		}
		for i, rawInt := range rawInts {
			energyLevel, err := strconv.Atoi(rawInt)
			if err != nil {
				return nil, fmt.Errorf("unable to parse energylevel %q: %w", rawInt, err)
			}
			point := Point{I: i, J: j}
			octos[point] = &Octopus{Point: point, EnergyLevel: energyLevel}
		}
	}

	return &Grid{Octopodes: octos, Length: len(rawLines), Width: len(rawLines[0])}, nil
}

type Point struct {
	I, J int
}

type Octopus struct {
	Point
	EnergyLevel int
}

type Grid struct {
	Length, Width int
	Octopodes map[Point]*Octopus
}

func (g *Grid) Size() int {
	return g.Length * g.Width
}

func (g *Grid) InBounds(p Point) bool {
	return 0 <= p.J && p.J < g.Length && 0 <= p.I && p.I < g.Width
}

func (g *Grid) Neighbors(o *Octopus) []*Octopus {
	// Get all 8 neigbors.
	// Up/Down is j-1 and j+1
	// Left/Right is i-1 and i+1
	var neighbors []*Octopus
	for j := o.J-1; j <= o.J+1; j++ {
		for i := o.I-1; i <= o.I+1; i++ {
			point := Point{I: i, J: j}
			if point == o.Point {
				continue // Skip self.
			}
			if !g.InBounds(point) {
				continue // skip out of bounds points.
			}
			neighbors = append(neighbors, g.Octopodes[point])
		}
	}
	return neighbors
}

func (g *Grid) SimulateRound() int {
	// Increment the energy level of each octopus.
	var toFlash []*Octopus
	for _, octo := range g.Octopodes {
		octo.EnergyLevel++
		if octo.EnergyLevel > 9 {
			toFlash = append(toFlash, octo)
		}
	}

	// Trigger flashes and neighbor increments.
	for len(toFlash) > 0 {
		popIdx := len(toFlash)-1
		popped := toFlash[popIdx]
		toFlash = toFlash[:popIdx]

		neighbors := g.Neighbors(popped)
		for _, octo := range neighbors {
			if octo.EnergyLevel > 9 {
				continue // Already flashed
			}

			octo.EnergyLevel++
			if octo.EnergyLevel > 9 {
				toFlash = append(toFlash, octo)
			}
		}
	}

	// Count flashes and reset.
	numFlashes := 0
	for _, octo := range g.Octopodes {
		if octo.EnergyLevel > 9 {
			numFlashes++
			octo.EnergyLevel = 0
		}
	}
	return numFlashes
}

func main() {
	grid, err := loadGrid()
	if err != nil {
		log.Fatal(err)
	}

	totalFlashes := 0
	maxSteps := 1000
	var i int
	for i = 0; i < maxSteps; i++ {
		newFlashes := grid.SimulateRound()
		totalFlashes += newFlashes
		if newFlashes == 100 {
			fmt.Printf("All octopuses flashed on step %d\n", i+1)
			break
		}
	}

	fmt.Printf("after %d steps there were %d flashes\n", i+1, totalFlashes)
}
