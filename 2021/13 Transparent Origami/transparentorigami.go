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

func readInputLines() []string {
	inputFile, err := os.Open(inputFilename)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to open input file: %w", err))
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
		log.Fatal(fmt.Errorf("unexpected error reading input: %w", err))
	}

	return lines
}

type Point struct {
	X, Y int
}

func (p Point) IsPastFold(fold Point) bool {
	return (0 < fold.X && fold.X < p.X) || (0 < fold.Y && fold.Y < p.Y)
}

func (p Point) MirrorAcrossFold(fold Point) Point {
	if fold.X > 0 {
		// Fold Y value across x=fold.X
		return Point{X: 2*fold.X - p.X, Y: p.Y}
	}
	// Fold X value across y=fold.Y
	return Point{X: p.X, Y: 2*fold.Y - p.Y}
}

type DotGrid struct {
	Dots map[Point]bool

	MaxX, MaxY int

	Folds []Point
}

func loadDotGrid() *DotGrid {
	rawLines := readInputLines()

	parseDots := true
	maxX, maxY := 0, 0
	dots := map[Point]bool{}
	var folds []Point
	for _, rawLine := range rawLines {
		if rawLine == "" {
			// Empty line before fold instructions begin.
			parseDots = false
			continue
		}

		if parseDots {
			coords := strings.Split(rawLine, ",")
			xCoord, err := strconv.Atoi(coords[0])
			if err != nil {
				log.Fatal(err)
			}
			yCoord, err := strconv.Atoi(coords[1])
			if err != nil {
				log.Fatal(err)
			}

			dot := Point{X: xCoord, Y: yCoord}
			dots[dot] = true
			if dot.X > maxX {
				maxX = dot.X
			}
			if dot.Y > maxY {
				maxY = dot.Y
			}
		} else {
			rawFold := strings.Split(strings.TrimPrefix(rawLine, "fold along "), "=")
			val, err := strconv.Atoi(rawFold[1])
			if err != nil {
				log.Fatal(err)
			}
			axis := rawFold[0]
			if axis == "x" {
				folds = append(folds, Point{X: val})
			} else {
				folds = append(folds, Point{Y: val})
			}
		}
	}

	return &DotGrid{
		Dots: dots,
		MaxX: maxX,
		MaxY: maxY,
		Folds: folds,
	}
}

func (g *DotGrid) PerformFold() {
	fold := g.Folds[0]
	g.Folds = g.Folds[1:]

	for dot := range g.Dots {
		if !dot.IsPastFold(fold) {
			continue
		}

		mirroredDot := dot.MirrorAcrossFold(fold)
		delete(g.Dots, dot)
		g.Dots[mirroredDot] = true
	}

	if fold.X > 0 {
		g.MaxX = fold.X-1
	} else {
		g.MaxY = fold.Y-1
	}

	fmt.Printf("after fold, there are %d dots\n", len(g.Dots))
}

func (g *DotGrid) Print() {
	for y := 0; y <= g.MaxY; y++ {
		for x := 0; x <= g.MaxX; x++ {
			if g.Dots[Point{X: x, Y: y}] {
				fmt.Printf("█")
			} else {
				fmt.Printf(" ")
			}
		}
		fmt.Println()
	}
}

func main() {
	grid := loadDotGrid()
	for len(grid.Folds) > 0 {
		grid.PerformFold()
	}
	grid.Print()
}
