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

func loadPointSets() []PointSet {
	rawLines := readInputLines()

	var pointSets []PointSet
	for len(rawLines) > 0 {
		var pointSet PointSet
		pointSet, rawLines = parsePointSet(rawLines)
		pointSets = append(pointSets, pointSet)
	}

	return pointSets
}

func parsePointSet(rawLines []string) (PointSet, []string) {
	// First line is always a "--- scanner N ---" header.
	rawLines = rawLines[1:]

	pointSet := PointSet{}
	for len(rawLines) > 0 {
		line := rawLines[0]
		rawLines = rawLines[1:]

		if line == "" {
			break // Empty line between point sets.
		}

		pointSet[parsePoint(line)] = true
	}

	return pointSet, rawLines
}

func parsePoint(line string) Point {
	parts := strings.Split(line, ",")
	x, err := strconv.Atoi(parts[0])
	if err != nil { panic(err) }
	y, err := strconv.Atoi(parts[1])
	if err != nil { panic(err) }
	z, err := strconv.Atoi(parts[2])
	if err != nil { panic(err) }

	return Point{X: x, Y: y, Z: z}
}

type Point struct {
	X, Y, Z int
}

func (p Point) Add(o Point) Point {
	return Point{X: p.X+o.X, Y: p.Y+o.Y, Z: p.Z+o.Z}
}

func (p Point) Subtract(o Point) Point {
	return Point{X: p.X-o.X, Y: p.Y-o.Y, Z: p.Z-o.Z}
}

func (p Point) Scale(n int) Point {
	return Point{X: p.X*n, Y: p.Y*n, Z: p.Z*n}
}

func (p Point) Rotate90X() Point {
	return Point{X: p.X, Y: -p.Z, Z: p.Y}
}

func (p Point) Rotate90Y() Point {
	return Point{X: p.Z, Y: p.Y, Z: -p.X}
}

func (p Point) Rotate90Z() Point {
	return Point{X: -p.Y, Y: p.X, Z: p.Z}
}

func (p Point) ManhattanDistance(o Point) int {
	dx := p.X - o.X
	if dx < 0 {
		dx = -dx
	}
	dy := p.Y - o.Y
	if dy < 0 {
		dy = -dy
	}
	dz := p.Z - o.Z
	if dz < 0 {
		dz = -dz
	}
	return dx + dy + dz
}

type PointSet map[Point]bool

func (p PointSet) Remap(remapFunc func(Point) Point) PointSet {
	remapped := make(PointSet, len(p))
	for point := range p {
		remapped[remapFunc(point)] = true
	}
	return remapped
}

func (p PointSet) Clone() PointSet {
	// Clone is a remap with an identity function.
	return p.Remap(func(point Point) Point { return point })
}

func (p PointSet) Translate(v Point) PointSet {
	// Translate is a remap with an Add function.
	return p.Remap(func(point Point) Point { return point.Add(v) })
}

func (p PointSet) Rotate90X() PointSet {
	return p.Remap(func(point Point) Point { return point.Rotate90X() })
}

func (p PointSet) Rotate90Y() PointSet {
	return p.Remap(func(point Point) Point { return point.Rotate90Y() })
}

func (p PointSet) Rotate90Z() PointSet {
	return p.Remap(func(point Point) Point { return point.Rotate90Z() })
}

func (p PointSet) Orientations() []PointSet {
	// There are 24 possible orientations.
	orientations := make([]PointSet, 24)

	orientations[0] = p.Clone()	// Forward.
	orientations[1] = orientations[0].Rotate90Y()
	orientations[2] = orientations[1].Rotate90Y()
	orientations[3] = orientations[2].Rotate90Y()
	orientations[4] = orientations[3].Rotate90Y().Rotate90Z() // Left.
	orientations[5] = orientations[4].Rotate90Y()
	orientations[6] = orientations[5].Rotate90Y()
	orientations[7] = orientations[6].Rotate90Y()
	orientations[8] = orientations[7].Rotate90Y().Rotate90Z() // Back.
	orientations[9] = orientations[8].Rotate90Y()
	orientations[10] = orientations[9].Rotate90Y()
	orientations[11] = orientations[10].Rotate90Y()
	orientations[12] = orientations[11].Rotate90Y().Rotate90Z() // Right.
	orientations[13] = orientations[12].Rotate90Y()
	orientations[14] = orientations[13].Rotate90Y()
	orientations[15] = orientations[14].Rotate90Y()
	orientations[16] = orientations[15].Rotate90Y().Rotate90Z().Rotate90X() // Up.
	orientations[17] = orientations[16].Rotate90Y()
	orientations[18] = orientations[17].Rotate90Y()
	orientations[19] = orientations[18].Rotate90Y()
	orientations[20] = orientations[19].Rotate90Y().Rotate90X().Rotate90X() // Down.
	orientations[21] = orientations[20].Rotate90Y()
	orientations[22] = orientations[21].Rotate90Y()
	orientations[23] = orientations[22].Rotate90Y()

	return orientations
}

func (p PointSet) Union(o PointSet) PointSet {
	unioned := make(PointSet, len(p)+len(o))
	for point := range p {
		unioned[point] = true
	}
	for point := range o {
		unioned[point] = true
	}
	return unioned
}

func (p PointSet) Intersect(o PointSet) PointSet {
	maxCap := len(p)
	if len(o) > maxCap {
		maxCap = len(o)
	}
	intersection := make(PointSet, maxCap)
	for point := range p {
		if o[point] {
			intersection[point] = true
		}
	}
	return intersection
}

func (p PointSet) Points() []Point {
	points := make([]Point, 0, len(p))
	for point := range p {
		points = append(points, point)
	}
	return points
}

func tryMatchScannerData(a, b PointSet) (PointSet, Point) {
	for _, bOrientation := range b.Orientations() {
		bPoints :=  bOrientation.Points()
		for _, bPoint := range bPoints[:len(bPoints)-11] {
			for aPoint := range a {
				translation := aPoint.Subtract(bPoint)
				bTranslated := bOrientation.Translate(translation)
				if len(a.Intersect(bTranslated)) >= 12 {
					return a.Union(bTranslated), translation
				}
			}
		}		
	}
	return nil, Point{}
}

func reducePointSets(pointSets []PointSet, scannerPoints map[Point]bool) []PointSet {
	a := pointSets[0]
	pointSets = pointSets[1:]
	for i, b := range pointSets {
		merged, scannerPoint := tryMatchScannerData(a, b)
		if merged != nil {
			scannerPoints[scannerPoint] = true
			pointSets = append(pointSets[:i], pointSets[i+1:]...)
			return append([]PointSet{merged}, pointSets...)
		}
	}
	panic("could not reduce")
}

func main() {
	pointSets := loadPointSets()
	fmt.Printf("loaded %d scanners\n", len(pointSets))
	for i, pointSet := range pointSets {
		fmt.Printf("scanner %d has %d beacons\n", i, len(pointSet))
	}

	scannerPoints := map[Point]bool{
		Point{}: true,
	}
	for len(pointSets) > 1 {
		pointSets = reducePointSets(pointSets, scannerPoints)
		fmt.Printf("reduced to %d sets\n", len(pointSets))
	}
	finalPointSet := pointSets[0]

	fmt.Printf("count of unique beacons: %d\n", len(finalPointSet))

	maxScannerDistance := 0
	for a := range scannerPoints {
		for b := range scannerPoints {
			distance := a.ManhattanDistance(b)
			if distance > maxScannerDistance {
				maxScannerDistance = distance
			}
		}
	}
	fmt.Printf("max scanner distace: %d\n", maxScannerDistance)
}
