package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const inputFilename = "./INPUT"

type Point struct {
	X, Y int
}

type VentLine struct {
	Start, End Point
}

func (v VentLine) IsHorizontal() bool {
	return v.Start.Y == v.End.Y
}

func (v VentLine) IsVertical() bool {
	return v.Start.X == v.End.X
}

func (v VentLine) Walk(walkFunc func(Point)) {
	dX := 1
	if v.IsVertical() {
		dX = 0
	} else if v.Start.X > v.End.X {
		dX = -1
	}

	dY := 1
	if v.IsHorizontal() {
		dY = 0
	} else if v.Start.Y > v.End.Y {
		dY = -1
	}

	p := v.Start
	for {
		walkFunc(p)

		if p == v.End {
			break
		}

		p.X += dX
		p.Y += dY
	}
}

func readVentLines() ([]VentLine, error) {
	rawLines, err := readInputLines()
	if err != nil {
		return nil, fmt.Errorf("unable to read input lines: %w", err)
	}

	ventLines := make([]VentLine, len(rawLines))
	for i, rawLine := range rawLines {
		parts := strings.Split(rawLine, " -> ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("unable to determine start and end of line %d: %q", i, rawLine)
		}

		rawStart, rawEnd := parts[0], parts[1]
		start, err := parsePoint(rawStart)
		if err != nil {
			return nil, fmt.Errorf("unable to parse start point %q: %w", rawStart, err)
		}

		end, err := parsePoint(rawEnd)
		if err != nil {
			return nil, fmt.Errorf("unable to parse end point %q: %w", rawEnd, err)
		}

		ventLines[i] = VentLine{
			Start: start,
			End: end,
		}
	}

	return ventLines, nil
}

func parsePoint(rawPoint string) (Point, error) {
	parts := strings.Split(rawPoint, ",")
	if len(parts) != 2 {
		return Point{}, fmt.Errorf("unable to split point x,y: %q", rawPoint)
	}

	rawX, rawY := parts[0], parts[1]
	x, err := strconv.Atoi(rawX)
	if err != nil {
		return Point{}, fmt.Errorf("unable to parse x coord %q: %w", rawX, err)
	}

	y, err := strconv.Atoi(rawY)
	if err != nil {
		return Point{}, fmt.Errorf("unable to parse y coord %q: %w", rawY, err)
	}

	return Point{X: x, Y: y}, nil
}

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

func mapVents(ventLines []VentLine) {
	ventCounts := make(map[Point]int)
	for _, ventLine := range ventLines {
		ventLine.Walk(func(p Point) {
			ventCounts[p] += 1
		})
	}

	pointsWithMultipleVents := 0
	for _, ventCount := range ventCounts {
		if ventCount > 1 {
			pointsWithMultipleVents++
		}
	}

	fmt.Printf("There are %d points with multiple vents\n", pointsWithMultipleVents)
}

func main() {
	ventLines, err := readVentLines()
	if err != nil {
		log.Fatal(err)
	}

	mapVents(ventLines)
}
