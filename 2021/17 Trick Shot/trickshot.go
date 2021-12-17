package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"math"
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

type TargetArea struct {
	MinX, MaxX int
	MinY, MaxY int
}

func loadTargetArea() TargetArea {
	rawLines := readInputLines()
	if len(rawLines) != 1 {
		log.Fatalf("expected 1 line but got %d\n", len(rawLines))
	}

	rawLine := rawLines[0]

	ranges := rawLine[strings.Index(rawLine, "x="):]
	parts := strings.Split(ranges, ", ")
	xRange, yRange := parts[0][2:], parts[1][2:]
	xRangeParts := strings.Split(xRange, "..")
	yRangeParts := strings.Split(yRange, "..")

	xMin, err := strconv.Atoi(xRangeParts[0])
	if err != nil { log.Fatal(err) }
	xMax, err := strconv.Atoi(xRangeParts[1])
	if err != nil { log.Fatal(err) }
	yMin, err := strconv.Atoi(yRangeParts[0])
	if err != nil { log.Fatal(err) }
	yMax, err := strconv.Atoi(yRangeParts[1])
	if err != nil { log.Fatal(err) }

	return TargetArea{
		MinX: xMin, MaxX: xMax,
		MinY: yMin, MaxY: yMax,
	}
}

func main() {
	targetArea := loadTargetArea()

	// Brute force method.
	// The minimum Y velocity to begin would be minY
	// maximum y velocity would be -(minY-1).
	// The minimum X velocity would be ceil(((8*minX +1)^0.5 - 1)/2) before it stops in the range
	// The maximum X velocity would be maxX
	// Iterate with all pairs in these ranges, simulating a probe launch until it goes beyond
	// the target range.
	globalMaxY := 0
	viableOptionsCount := 0
	minX0 := int(math.Ceil((math.Sqrt(8.0*float64(targetArea.MinX)+1.0)-1.0)/2.0))
	for dx0 := minX0; dx0 <= targetArea.MaxX; dx0++ {
		for dy0 := targetArea.MinY; dy0 <= -targetArea.MinY-1; dy0++ {
			dx, dy := dx0, dy0
			x, y := 0, 0
			maxY := 0
			for x <= targetArea.MaxX && y >= targetArea.MinY {
				if targetArea.MinX <= x && x <= targetArea.MaxX && targetArea.MinY <= y && y <= targetArea.MaxY {
					if maxY > globalMaxY {
						globalMaxY = maxY
					}
					viableOptionsCount++
					// fmt.Printf("%d,%d\n",dx0, dy0)
					break
				}
				x += dx
				y += dy
				if dx > 0 { dx-- }
				if dx < 0 { dx++ }
				dy--

				if y > maxY {
					maxY = y
				}
			}
		}
	}
	fmt.Printf("global max y is %d\n", globalMaxY)
	fmt.Printf("viable options count is %d\n", viableOptionsCount)
}
