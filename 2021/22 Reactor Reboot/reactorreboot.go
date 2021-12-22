package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	// "sort"
	"strconv"
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

func loadRebootSteps() []RebootStep {
	rawLines := readInputLines()
	steps := make([]RebootStep, len(rawLines))
	for i, line := range rawLines {
		steps[i] = parseRebootStep(line)
	}
	return steps
}

func parseRebootStep(line string) RebootStep {
	parts := strings.Split(line, " ")

	target := parts[0] == "on"

	ranges := strings.Split(parts[1], ",")

	xRange, yRange, zRange := parseRange(ranges[0]), parseRange(ranges[1]), parseRange(ranges[2])

	return RebootStep{
		Target: target,
		Range3D: Range3D{xRange, yRange, zRange},
	}
}

func parseRange(rawRange string) Range {
	// First two characters are x=, y=, or z= and can be ignored.
	rawRange = rawRange[2:]
	parts := strings.Split(rawRange, "..")

	min, err := strconv.Atoi(parts[0])
	if err != nil { panic(err) }
	max, err := strconv.Atoi(parts[1])
	if err != nil { panic(err) }

	// Note: a range is a half-open interval.
	return Range{int64(min), int64(max+1)}
}

// Range is the range of values between [min, max)
type Range struct {
	Min, Max int64
}

var AnEmptyRange = Range{0, 0} // Min >= Max

func (r Range) Length() int64 {
	return r.Max - r.Min
}

func (r Range) SubsetOf(s Range) bool {
	return s.Min <= r.Min && r.Max <= s.Max
}

func (r Range) IsEmpty() bool {
	return r.Min >= r.Max
}

func (r Range) Intersect(s Range) Range {
	// Establish the invariant that r.Min <= s.Min
	if s.Min < r.Min {
		// Swap them if not.
		r, s = s, r
	}

	if r.Max < s.Min {
		return AnEmptyRange
	}

	minMax := r.Max
	if s.Max < minMax {
		minMax = s.Max
	}

	return Range{s.Min, minMax}
}

// Range3D is INCLUSIVE!
type Range3D struct {
	X, Y, Z Range
}

var AnEmptyRange3D = Range3D{AnEmptyRange, AnEmptyRange, AnEmptyRange}

func (r Range3D) Volume() int64 {
	return r.X.Length() * r.Y.Length() * r.Z.Length()
}

func (r Range3D) SubsetOf(s Range3D) bool {
	return r.X.SubsetOf(s.X) && r.Y.SubsetOf(s.Y) && r.Z.SubsetOf(s.Z)
}

func (r Range3D) IsEmpty() bool {
	return r.X.IsEmpty() || r.Y.IsEmpty() || r.Z.IsEmpty()
}

func (r Range3D) Intersect(s Range3D) Range3D {
	return Range3D{
		X: r.X.Intersect(s.X),
		Y: r.Y.Intersect(s.Y),
		Z: r.Z.Intersect(s.Z),
	}
}

func (r Range3D) Subtract(s Range3D) []Range3D {
	s = r.Intersect(s)
	if s.IsEmpty() {
		return []Range3D{r}
	}

	// In 3 Dimensions, the intersecting cuboid range floats around somewhere
	// in the larger 3D cuboid range.
	// Let's rename them here.
	inner, outer := s, r

	// Let's consider the different sides of the inner cuboid independently:
	//   x (left, right)
	//   y (back, front)
	//	 z (bottom, top)
	left := Range{outer.X.Min, inner.X.Min}
	right := Range{inner.X.Max, outer.X.Max}
	back := Range{outer.Y.Min, inner.Y.Min}
	front := Range{inner.Y.Max, outer.Y.Max}
	bottom := Range{outer.Z.Min, inner.Z.Min}
	top := Range{inner.Z.Max, outer.Z.Max}

	cuboids := make([]Range3D, 0, 14)
	for _, cuboid := range []Range3D{
		// First we can carve off cuboids from the outer region into three regions
		// along the x-axis.
		{left, outer.Y, outer.Z},
		{right, outer.Y, outer.Z},
		// Next we can carve off cuboids from this region into three regions along
		// the y-axis.
		{inner.X, back, outer.Z},
		{inner.X, front, outer.Z},
		// Finally, carve off cuboids from this region along the z-axis.
		{inner.X, inner.Y, bottom},
		{inner.X, inner.Y, top},
	} {
		// Depending on how R and S intersect, some will be empty.
		// The only way that none are empty is if R completely contains
		// S with some buffer in all dimensions.
		if !cuboid.IsEmpty() {
			cuboids = append(cuboids, cuboid)
		}
	}

	return cuboids
}

type RebootStep struct {
	Target bool
	Range3D
}

func (s RebootStep) IsInitStep() bool {
	initRange := Range{-50, 50}
	return s.SubsetOf(Range3D{initRange, initRange, initRange})
}

type Point struct {
	X, Y, Z int64
}

type ReactorGrid []Range3D

func (g *ReactorGrid) Execute(step RebootStep) {
	oldGrid := *g
	newGrid := make(ReactorGrid, 0, len(oldGrid))
	for _, cuboid := range oldGrid {
		newGrid = append(newGrid, cuboid.Subtract(step.Range3D)...)
	}
	if step.Target {
		newGrid = append(newGrid, step.Range3D)
	}
	*g = newGrid
}

func (g ReactorGrid) OnCubeCount() int64 {
	var sum int64
	for _, cuboid := range g {
		sum += cuboid.Volume()
	}
	return sum
}

func main() {
	steps := loadRebootSteps()

	grid := ReactorGrid{}
	for _, step := range steps {
		if step.IsInitStep() {
			grid.Execute(step)
		}
	}

	fmt.Printf("After initialization procedure, %d cubes are on.\n", grid.OnCubeCount())

	grid = ReactorGrid{}
	for _, step := range steps {
		grid.Execute(step)
	}

	fmt.Printf("After full reboot procedure, %d cubes are on.\n", grid.OnCubeCount())
}
