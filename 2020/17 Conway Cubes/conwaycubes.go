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

type Point2D struct {
	X, Y int
}

func (p Point2D) SetXY(x, y int) Point2D {
	p.X, p.Y = x, y
	return p
}

func (p Point2D) JoinRegion(pointSet map[Point2D]bool) {
	for i := 0; i < 9; i++ {
		dx := (i % 3) - 1
		dy := (i/3 % 3) - 1
		pointSet[Point2D{p.X+dx, p.Y+dy}] = true
	}
}

func (p Point2D) ActiveNeighbors(pointSet map[Point2D]bool) int {
	activeNeighbors := 0
	for i := 0; i < 9; i++ {
		dx := (i % 3) - 1
		dy := (i/3 % 3) - 1
		if dx == dy && dy == 0 {
			continue // Do not include own position.
		}
		if pointSet[Point2D{p.X+dx, p.Y+dy}] {
			activeNeighbors++
		}
	}
	return activeNeighbors
}

type Point3D struct {
	X, Y, Z int
}

func (p Point3D) SetXY(x, y int) Point3D {
	p.X, p.Y = x, y
	return p
}

func (p Point3D) JoinRegion(pointSet map[Point3D]bool) {
	for i := 0; i < 27; i++ {
		dx := (i % 3) - 1
		dy := (i/3 % 3) - 1
		dz := (i/9 % 3) - 1
		pointSet[Point3D{p.X+dx, p.Y+dy, p.Z+dz}] = true
	}
}

func (p Point3D) ActiveNeighbors(pointSet map[Point3D]bool) int {
	activeNeighbors := 0
	for i := 0; i < 27; i++ {
		dx := (i % 3) - 1
		dy := (i/3 % 3) - 1
		dz := (i/9 % 3) - 1
		if dx == dy && dy == dz && dz == 0 {
			continue // Do not include own position.
		}
		if pointSet[Point3D{p.X+dx, p.Y+dy, p.Z+dz}] {
			activeNeighbors++
		}
	}
	return activeNeighbors
}

type Point4D struct {
	X, Y, Z, W int
}

func (p Point4D) SetXY(x, y int) Point4D {
	p.X, p.Y = x, y
	return p
}

func (p Point4D) JoinRegion(pointSet map[Point4D]bool) {
	for i := 0; i < 81; i++ {
		dx := (i % 3) - 1
		dy := (i/3 % 3) - 1
		dz := (i/9 % 3) - 1
		dw := (i/27 % 3) -1
		pointSet[Point4D{p.X+dx, p.Y+dy, p.Z+dz, p.W+dw}] = true
	}
}

func (p Point4D) ActiveNeighbors(pointSet map[Point4D]bool) int {
	activeNeighbors := 0
	for i := 0; i < 81; i++ {
		dx := (i % 3) - 1
		dy := (i/3 % 3) - 1
		dz := (i/9 % 3) - 1
		dw := (i/27 % 3) -1
		if dx == dy && dy == dz && dz == dw && dw == 0 {
			continue // Do not include own position.
		}
		if pointSet[Point4D{p.X+dx, p.Y+dy, p.Z+dz, p.W+dw}] {
			activeNeighbors++
		}
	}
	return activeNeighbors
}

type Point[P comparable] interface {
	SetXY(x, y int) P
	JoinRegion(pointSet map[P]bool)
	ActiveNeighbors(pointSet map[P]bool) int

	comparable
}

type Grid[P Point[P]] struct {
	ActivePoints map[P]bool
}

func loadGrid[P Point[P]]() *Grid[P] {
	activePoints := map[P]bool{}
	for j, line := range readInputLines() {
		for i, char := range []byte(line) {
			if char == '#' {
				var p P
				activePoints[p.SetXY(i, j)] = true
			}
		}
	}
	return &Grid[P]{ActivePoints: activePoints}
}

func (g *Grid[P]) Cycle() bool {
	paddedRegion := make(map[P]bool, 27*len(g.ActivePoints))
	for point := range g.ActivePoints {
		point.JoinRegion(paddedRegion)
	}

	stateChanged := false
	nextCycle := make(map[P]bool, 8*len(g.ActivePoints))
	for point := range paddedRegion {
		isActive := g.ActivePoints[point]
		activeNeighbors := point.ActiveNeighbors(g.ActivePoints)
		if isActive && (activeNeighbors == 2 || activeNeighbors == 3) || !isActive && activeNeighbors == 3 {
			nextCycle[point] = true
		}
		stateChanged = stateChanged || isActive && !(activeNeighbors == 2 || activeNeighbors == 3) || !isActive && activeNeighbors == 3
	}

	g.ActivePoints = nextCycle
	return stateChanged
}

func (g *Grid[P]) Print(min, max Point2D) {
	var b strings.Builder
	for j := min.Y; j < max.Y; j++ {
		for i := min.X; i < max.X; i++ {
			var p P
			if g.ActivePoints[p.SetXY(i, j)] {
				b.WriteByte('#')
			} else {
				b.WriteByte(' ')
			}
		}
		b.WriteByte('\n')
	}
	fmt.Printf("\033[2J\033[H")
	fmt.Println(b.String())
}

func (g *Grid[P]) Size() int {
	return len(g.ActivePoints)
}

func main() {
	grid3D := loadGrid[Point3D]()
	for i := 0; i < 6; i++ {
		grid3D.Cycle()
	}

	fmt.Printf("number of 3D cubes active after boot process: %d\n", len(grid3D.ActivePoints))

	grid4D := loadGrid[Point4D]()
	for i := 0; i < 6; i++ {
		grid4D.Cycle()
	}

	fmt.Printf("number of 4D cubes active after boot process: %d\n", len(grid4D.ActivePoints))

	// grid2D := loadGrid[Point2D]()
	// min, max := Point2D{-25, -25}, Point2D{25, 25}
	// grid2D.Print(min, max)
	// for grid2D.Cycle() {
	// 	grid2D.Print(min, max)
	// 	time.Sleep(30*time.Millisecond)
	// }
}
