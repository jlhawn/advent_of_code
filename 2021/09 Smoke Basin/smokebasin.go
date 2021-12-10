package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
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

type Point struct{
	I, J int
}

type PointHeight struct {
	Point
	Height int
}

type Grid struct {
	Heights map[Point]int
	Length, Width int
}

func loadGrid() (Grid, error) {
	rawLines, err := readInputLines()
	if err != nil {
		return Grid{}, fmt.Errorf("unable to read lines: %w", err)
	}

	length := len(rawLines)
	width := len(rawLines[0])

	heights := make(map[Point]int)
	for j, rawLine := range rawLines {
		rawInts := strings.Split(rawLine, "")
		for i, rawInt := range rawInts {
			height, err := strconv.Atoi(rawInt)
			if err != nil {
				return Grid{}, fmt.Errorf("unable to parse int %q: %w", rawInt, err)
			}
			heights[Point{I: i, J: j}] = height
		}
	}

	return Grid{
		Heights: heights,
		Length: length,
		Width: width,
	}, nil
}

func (g Grid) Height(p Point) int {
	return g.Heights[p]
}

func (g Grid) InBounds(p Point) bool {
	return 0 <= p.I && p.I < g.Width && 0 <= p.J && p.J < g.Length
}

// NeighborHeights returns neighboring points and respective heights.
func (g Grid) NeighborHeights(p Point) []PointHeight {
	if !g.InBounds(p) {
		return nil
	}

	up := Point{p.I, p.J-1}
	down := Point{p.I, p.J+1}
	left := Point{p.I-1, p.J}
	right := Point{p.I+1, p.J}

	var neighbors []PointHeight
	for _, point := range []Point{up, down, left, right} {
		if g.InBounds(point) {
			neighbors = append(neighbors, PointHeight{Point: point, Height: g.Height(point)})
		}
	}

	return neighbors
}

// UpperBasinNeighbors are neighboring points whose value is
// greater than or equal to this point's value (but less than 9).
func (g Grid) UpperBasinNeighbors(p Point) []Point {
	neighborHeights := g.NeighborHeights(p)
	height := g.Height(p)

	var upperBasinNeighbors []Point
	for _, neighbor := range neighborHeights {
		if 9 > neighbor.Height  && neighbor.Height >= height {
			upperBasinNeighbors = append(upperBasinNeighbors, neighbor.Point)
		}
	}

	return upperBasinNeighbors
}

func (g Grid) IsLowPoint(p Point) bool {
	// Point is the low point if all of its neighbors are higher.
	height := g.Height(p)
	
	neighbors := g.NeighborHeights(p)
	for _, neighbor := range neighbors {
		if neighbor.Height <= height {
			return false
		}
	}

	return true
}

func sumLowPointRiskLevels(grid Grid) {
	sum := 0
	for point, height := range grid.Heights {
		if grid.IsLowPoint(point) {
			sum += height +1
		}
	}

	fmt.Printf("The sum of the low point risk levels is %d\n", sum)
}

type Basin map[Point]struct{}

func determineBasinSizes(grid Grid) {
	// Find the low points, create an initial basin for each.
	var lowPoints []Point
	for point := range grid.Heights {
		if grid.IsLowPoint(point) {
			lowPoints = append(lowPoints, point)
		}
	}

	fmt.Printf("There are %d low points\n", len(lowPoints))

	// Perform a depth first search from the initial point
	// to expand to all points in the basin, adding to the basin's set
	// of points.
	var basins []Basin
	for _, lowPoint := range lowPoints {
		basin := Basin{}
		basins = append(basins, basin)

		fringe := []Point{lowPoint}
		for len(fringe) > 0 {
			// Pop this point of the fringe.
			popIdx := len(fringe)-1
			popped := fringe[popIdx]
			fringe = fringe[:popIdx]

			if _, ok := basin[popped]; ok {
				// Point is already in the basin's set of points.
				continue
			}

			// Add this point to the basin.
			basin[popped] = struct{}{}

			// Add its neighbors to the fringe.
			neighbors := grid.UpperBasinNeighbors(popped)
			for _, neighbor := range neighbors {
				fringe = append(fringe, neighbor)
			}
		}
	}

	fmt.Printf("Done building basins\n")

	// Find the largest 3 basins and multiply their sizes.
	// Note: this sorts in reverse order by using a *greater* func.
	sort.Slice(basins, func(i, j int) bool { return len(basins[i]) > len(basins[j]) })

	product := len(basins[0])
	fmt.Printf("basin with size %d\n", product)
	for _, basin := range basins[1:3] {
		fmt.Printf("basin with size %d\n", len(basin))
		product *= len(basin)
	}

	fmt.Printf("product of 3 largest basins is %d\n", product)
}

func main() {
	grid, err := loadGrid()
	if err != nil {
		log.Fatal(err)
	}

	sumLowPointRiskLevels(grid)

	determineBasinSizes(grid)
}
