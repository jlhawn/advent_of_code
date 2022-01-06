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

	"../../slices"
	// "../../streams"
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

func (p Point) InBounds(min, max Point) bool {
	return min.X <= p.X && p.X < max.X && min.Y <= p.Y && p.Y < max.Y
}

func (p Point) AdjacentPoints(min, max Point) []Point {
	points := make([]Point, 0, 8)
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			loc := Point{p.X+dx, p.Y+dy}
			if loc.InBounds(min, max) && loc != p {
				points = append(points, loc)
			}
		}
	}
	return points
}

func (p Point) LineOfSightSeats(g *Grid) []Point {
	points := make([]Point, 0, 8)
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}

			// Step in this direction until we find a seat
			// or step out of bounds.
			loc := Point{p.X+dx, p.Y+dy}
			for g.FloorSpaces[loc] {
				loc.X += dx
				loc.Y += dy
			}
			if loc.InBounds(g.Min, g.Max) {
				points = append(points, loc)
			}
		}
	}
	return points
}

type Grid struct {
	OccupiedSeats map[Point]bool
	FloorSpaces map[Point]bool
	Min, Max Point
}

func loadGrid() *Grid {
	floors := map[Point]bool{}
	lines := readInputLines()
	for j, line := range lines {
		for i, seatOrFlor := range []byte(line) {
			if seatOrFlor == '.' {
				floors[Point{i, j}] = true
			}
		}
	}
	return &Grid{
		OccupiedSeats: map[Point]bool{},
		FloorSpaces: floors,
		Max: Point{X: len(lines[0]), Y: len(lines)},
	}
}

func (g *Grid) Print() {
	var b strings.Builder
	for y := g.Min.Y; y < g.Max.Y; y++ {
		for x := g.Min.X; x < g.Max.X; x++ {
			loc := Point{x, y}
			if g.FloorSpaces[loc] {
				b.WriteByte('.')
			} else if g.OccupiedSeats[loc] {
				b.WriteByte('#')
			} else {
				b.WriteByte('L')
			}
		}
		b.WriteByte('\n')
	}
	fmt.Print(b.String())
}

func (g *Grid) SimulateRound1() bool {
	nextOccupiedSeats := make(map[Point]bool, 2*len(g.OccupiedSeats))

	stateChanged := false
	for x := g.Min.X; x < g.Max.X; x++ {
		for y := g.Min.Y; y < g.Max.Y; y++ {
			loc := Point{x, y}
			if g.FloorSpaces[loc] {
				continue // Floor spaces never change.
			}
			isOccupied := g.OccupiedSeats[loc]

			if isOccupied {
				// Assume it stays occupied.
				nextOccupiedSeats[loc] = true
			}

			occupiedAdjacentSeats := len(slices.Filter(func(p Point) bool {
				return g.OccupiedSeats[p]
			}, loc.AdjacentPoints(g.Min, g.Max)...))

			if !isOccupied && occupiedAdjacentSeats == 0 {
				// Becomes occupied.
				nextOccupiedSeats[loc] = true
				stateChanged = true
			}
			if isOccupied && occupiedAdjacentSeats >= 4 {
				// Becomes empty.
				delete(nextOccupiedSeats, loc)
				stateChanged = true
			}
		}
	}
	g.OccupiedSeats = nextOccupiedSeats
	return stateChanged
}

func (g *Grid) SimulateRound2() bool {
	nextOccupiedSeats := make(map[Point]bool, 2*len(g.OccupiedSeats))

	stateChanged := false
	for x := g.Min.X; x < g.Max.X; x++ {
		for y := g.Min.Y; y < g.Max.Y; y++ {
			loc := Point{x, y}
			if g.FloorSpaces[loc] {
				continue // Floor spaces never change.
			}
			isOccupied := g.OccupiedSeats[loc]

			if isOccupied {
				// Assume it stays occupied.
				nextOccupiedSeats[loc] = true
			}

			occupiedAdjacentSeats := len(slices.Filter(func(p Point) bool {
				return g.OccupiedSeats[p]
			}, loc.LineOfSightSeats(g)...))

			if !isOccupied && occupiedAdjacentSeats == 0 {
				// Becomes occupied.
				nextOccupiedSeats[loc] = true
				stateChanged = true
			}
			if isOccupied && occupiedAdjacentSeats >= 5 {
				// Becomes empty.
				delete(nextOccupiedSeats, loc)
				stateChanged = true
			}
		}
	}
	g.OccupiedSeats = nextOccupiedSeats
	return stateChanged
}

func main() {
	grid := loadGrid()

	var steps int
	for grid.SimulateRound1() {
		steps++
	}
	fmt.Printf("after %d steps, grid stabilized with %d occupied seats\n", steps, len(grid.OccupiedSeats))

	grid = loadGrid()
	steps = 0
	for grid.SimulateRound2() {
		steps++
	}
	fmt.Printf("after %d steps, grid stabilized with %d occupied seats\n", steps, len(grid.OccupiedSeats))
}
