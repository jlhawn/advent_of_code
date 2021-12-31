package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"../../slices"
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

func loadRebootSteps() []ReactorCuboid {
	return slices.Map(parseRebootStep, readInputLines()...)
}

func parseRebootStep(line string) ReactorCuboid {
	onOrOff, allRanges, _ := strings.Cut(line, " ")
	ranges := strings.Split(allRanges, ",")

	xRange, yRange, zRange := parseRange(ranges[0]), parseRange(ranges[1]), parseRange(ranges[2])

	return ReactorCuboid{
		IsOn: onOrOff == "on",
		Range3D: Range3D{xRange, yRange, zRange},
	}
}

func parseRange(rawRange string) Range {
	// First two characters are x=, y=, or z= and can be ignored.
	rawRange = rawRange[2:]
	rawMin, rawMax, _ := strings.Cut(rawRange, "..")

	min, err := strconv.Atoi(rawMin)
	if err != nil { panic(err) }
	max, err := strconv.Atoi(rawMax)
	if err != nil { panic(err) }

	// Note: a range is a half-open interval.
	return Range{int64(min), int64(max+1)}
}

// Range is the range of values between [min, max)
type Range struct {
	Min, Max int64
}

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
	return Range{slices.Max(r.Min, s.Min), slices.Min(r.Max, s.Max)}
}

type Range3D struct {
	X, Y, Z Range
}

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

type ReactorCuboid struct {
	IsOn bool
	Range3D
}

func (c ReactorCuboid) IsInitStep() bool {
	initRange := Range{-50, 50}
	return c.SubsetOf(Range3D{initRange, initRange, initRange})
}

type Point struct {
	X, Y, Z int64
}

type ReactorGrid []Range3D

func (g *ReactorGrid) Execute(step ReactorCuboid) {
	oldGrid := *g
	newGrid := make(ReactorGrid, 0, len(oldGrid))
	for _, cuboid := range oldGrid {
		newGrid = append(newGrid, cuboid.Subtract(step.Range3D)...)
	}
	if step.IsOn {
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

type ReactorGrid2 map[Range3D]int64

func (g ReactorGrid2) Increment(cuboid Range3D, by int64) {
	g[cuboid] += by
}

func (g ReactorGrid2) Remove(cuboid Range3D) {
	delete(g, cuboid)
}

func (g ReactorGrid2) Update(updates ReactorGrid2) {
	for cuboid, count := range updates {
		g[cuboid] += count
		if g[cuboid] == 0 {
			delete(g, cuboid)
		}
	}
}

func (g ReactorGrid2) Execute(step ReactorCuboid) {
	updates := make(ReactorGrid2, len(g))
	for cuboid, count := range g {
		if cuboid.SubsetOf(step.Range3D) {
			g.Remove(cuboid)
			continue
		}

		intersection := cuboid.Intersect(step.Range3D)
		if intersection.IsEmpty() {
			continue
		}

		updates.Increment(intersection, -count)
	}

	if step.IsOn {
		updates.Increment(step.Range3D, 1)
	}

	g.Update(updates)
}

func (g ReactorGrid2) OnCubeCount() int64 {
	var sum int64
	for cuboid, count := range g {
		sum += count * cuboid.Volume()
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

	start := time.Now()
	grid = ReactorGrid{}
	for _, step := range steps {
		grid.Execute(step)
	}

	fmt.Printf("After full reboot procedure, %d cubes are on in %s.\n", grid.OnCubeCount(), time.Since(start))

	start = time.Now()
	grid2 := ReactorGrid2{}
	for _, step := range steps {
		grid2.Execute(step)
	}

	fmt.Printf("After full reboo2 procedure, %d cubes are on in %s.\n", grid2.OnCubeCount(), time.Since(start))
}
