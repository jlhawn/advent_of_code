package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
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

func loadPointSets() []*PointSet {
	rawLines := readInputLines()

	var pointSets []*PointSet
	for len(rawLines) > 0 {
		var pointSet *PointSet
		pointSet, rawLines = parsePointSet(rawLines)
		pointSets = append(pointSets, pointSet)
	}

	return pointSets
}

func parsePointSet(rawLines []string) (*PointSet, []string) {
	// First line is always a "--- scanner N ---" header.
	rawLines = rawLines[1:]

	pointSet := NewPointSet(10)
	for len(rawLines) > 0 {
		line := rawLines[0]
		rawLines = rawLines[1:]

		if line == "" {
			break // Empty line between point sets.
		}

		pointSet.Add(parsePoint(line))
	}

	return pointSet, rawLines
}

func parsePoint(line string) *Point {
	parts := strings.Split(line, ",")
	x, err := strconv.Atoi(parts[0])
	if err != nil { panic(err) }
	y, err := strconv.Atoi(parts[1])
	if err != nil { panic(err) }
	z, err := strconv.Atoi(parts[2])
	if err != nil { panic(err) }

	return NewPoint(x, y, z)
}

type Point struct {
	X, Y, Z int
}

// NewPoint creates a new point with the given coordinates.
func NewPoint(x, y, z int) *Point {
	return &Point{X: x, Y: y, Z: z}
}

// Clone creates a copy of this point and returns a pointer to it.
// Does not mutate this point.
func (p *Point) Clone() *Point {
	return NewPoint(p.X, p.Y, p.Z)
}

// Add adds the given point to this point, mutating itself.
// Returns itself to allow chaining.
func (p *Point) Add(o *Point) *Point {
	p.X += o.X
	p.Y += o.Y
	p.Z += o.Z
	return p
}

// Subtract subtracts the given point from this point, mutating itself.
// Returns itself to allow chaining.
func (p *Point) Subtract(o *Point) *Point {
	p.X -= o.X
	p.Y -= o.Y
	p.Z -= o.Z
	return p
}

// Rotate90X rotates the given point 90 degrees about the X axis, mutating itself.
// Returns itself to allow chaining.
func (p *Point) Rotate90X() *Point {
	p.Y, p.Z = -p.Z, p.Y
	return p
}

// Rotate90Y rotates the given point 90 degrees about the Y axis, mutating itself.
// Returns itself to allow chaining.
func (p *Point) Rotate90Y() *Point {
	p.X, p.Z = p.Z, -p.X
	return p
}

// Rotate90Z rotates the given point 90 degrees about the Z axis, mutating itself.
// Returns itself to allow chaining.
func (p *Point) Rotate90Z() *Point {
	p.X, p.Y = -p.Y, p.X
	return p
}

func (p *Point) ManhattanDistance(o *Point) int {
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

type PointSet struct {
	Points []*Point
	Set map[Point]bool

	// A cache of 24 different orientations of this PointSet.
	OrientationsCache []*PointSet
}

// NewPointSet creates a new empty PointSet with the given capacity hint.
// It is able to grow its capacity as needed.
func NewPointSet(cap int) *PointSet {
	return &PointSet{
		Points: make([]*Point, 0, cap),
		Set: make(map[Point]bool, cap),
	}
}

// Size returns the number of elements in this set.
func (p *PointSet) Size() int {
	return len(p.Set)
}

func (p *PointSet) Contains(point *Point) bool {
	return p.Set[*point]
}

// Add inserts the given point into this point set if it is
// not in the set already.
func (p *PointSet) Add(point *Point) {
	if p.Contains(point) {
		return // The point is already in the set.
	}

	p.Points = append(p.Points, point)
	p.Set[*point] = true
}

// Clone creates a new PointSet which is a copy of this one.
func (p *PointSet) Clone() *PointSet {
	cloned := NewPointSet(len(p.Points))
	for _, point := range p.Points {
		cloned.Add(point.Clone())
	}
	return cloned
}

// Remap modifies every point in this set with the given remap func.
// The remap function should be bijective, i.e., an invertible,
// one-to-one function which mutates the given point.
// Returns itself to allow chaining.
func (p *PointSet) Remap(remapFunc func(*Point)) *PointSet {
	for i := range p.Points {
		delete(p.Set, *p.Points[i])
		remapFunc(p.Points[i])
	}

	for i := range p.Points {
		p.Set[*p.Points[i]] = true
	}

	return p
}

// Translate is a remap with an Add function.
// Returns itself to allow chaining.
func (p *PointSet) Translate(v *Point) *PointSet {
	return p.Remap(func(point *Point) { point.Add(v) })
}

// Rotate90X is a remap which rotates each point 90 degrees about the x axis.
// Returns itself to allow chaining.
func (p *PointSet) Rotate90X() *PointSet {
	return p.Remap(func(point *Point) { point.Rotate90X() })
}

// Rotate90Y is a remap which rotates each point 90 degrees about the y axis.
// Returns itself to allow chaining.
func (p *PointSet) Rotate90Y() *PointSet {
	return p.Remap(func(point *Point) { point.Rotate90Y() })
}

// Rotate90Z is a remap which rotates each point 90 degrees about the z axis.
// Returns itself to allow chaining.
func (p *PointSet) Rotate90Z() *PointSet {
	return p.Remap(func(point *Point) { point.Rotate90Z() })
}

// Union adds every point from the given set to this set.
// Returns itself to allow chaining.
func (p *PointSet) Union(o *PointSet) *PointSet {
	for _, point := range o.Points {
		p.Add(point)
	}
	return p
}

// IntersectionSize returns the number of points in the given point set which
// are also elements of this point set.
func (p *PointSet) IntersectionSize(o *PointSet) int {
	largerSet, smallerSet := p, o
	if smallerSet.Size() > largerSet.Size() {
		largerSet, smallerSet = o, p
	}

	count := 0
	for _, point := range smallerSet.Points {
		if largerSet.Contains(point) {
			count++
		}
	}
	return count
}

type OrientationIterator struct {
	*PointSet
	Iteration int
}

func (i *OrientationIterator) IsEnd() bool {
	// There are 24 possible orientations.
	return i.Iteration == 24
}

func (i *OrientationIterator) Next() {
	i.Iteration++

	// Always rotate 90 degrees around the y-axis.
	i.PointSet.Rotate90Y()

	if i.Iteration % 4 == 0 {
		// We're back to our original rotation around the y-axis.
		switch i.Iteration {
		case 4, 8, 12:
			i.PointSet.Rotate90Z() // Face left, back, right.
		case 16:
			i.PointSet.Rotate90Z().Rotate90X() // Face up.
		case 20:
			i.PointSet.Rotate90X().Rotate90X() // Face down.
		}
	}
}

// Orientations returns a slice of the 24 orientations for this PointSet.
// These orientations will be cached so that subsequent calls return the
// same 24 orientation PointSets. This PointSet is not mutated at all.
func (p *PointSet) Orientations() []*PointSet {
	if p.OrientationsCache != nil {
		return p.OrientationsCache
	}

	for it := (&OrientationIterator{PointSet: p.Clone()}); !it.IsEnd(); it.Next() {
		p.OrientationsCache = append(p.OrientationsCache, it.PointSet.Clone())
	}

	return p.OrientationsCache
}

type ThreadSafeMap[K, V any] struct {
	sync.Map
}

func (m *ThreadSafeMap[K, V]) Delete(key K) {
	m.Map.Delete(key)
}

func (m *ThreadSafeMap[K, V]) Load(key K) (V, bool) {
	val, ok := m.Map.Load(key)
	if ok {
		return val.(V), ok
	}

	var zero V
	return zero, ok
}

func (m *ThreadSafeMap[K, V]) Store(key K, value V) {
	m.Map.Store(key, value)
}

type TranslationCacheKey struct {
	*PointSet
	Point
}

type TranslationCache = ThreadSafeMap[TranslationCacheKey, *PointSet]

// TryMatchOrientations will attempt to patch the given candidate PointSet with
// the given reduced PointSet. If there is a translation of any orientation of
// the candidate PointSet which contains at least 12 of the same points then it
// is a match and that orientation PointSet and translation vector will be
// returned. If there is no match, returns (nil, nil).
//
// This function does not mutate the given PoinSets.
func TryMatchOrientations(reduced, candidate *PointSet, cache *TranslationCache) (match *PointSet, origin *Point) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	for _, orientation := range candidate.Orientations() {
		wg.Add(1)
		go func(orientation *PointSet) {
			defer wg.Done()

			// We don't need to check the last 11 points in the candidate orientation because
			// if there is a match then there would be a match with at least 12 points. If
			// there are 11 left then one of the previous points would have been a match.
			for _, b := range orientation.Points[:orientation.Size()-11] {
				for _, a := range reduced.Points {
					select {
					case <-ctx.Done():
						return
					default:
					}

					translation := a.Clone().Subtract(b)

					// The translated orientation may be cached from a previous attempt to
					// match this point in A to the point in B. The orientation pointers
					// are cached for every candidate, but that logic is internal to PointSet.
					cacheKey := TranslationCacheKey{orientation, *translation}
					bTranslated, ok := cache.Load(cacheKey)
					if !ok {
						bTranslated = orientation.Clone().Translate(translation)
						cache.Store(cacheKey, bTranslated)
					}

					if reduced.IntersectionSize(bTranslated) >= 12 {
						match, origin = bTranslated, translation
						cancel()
						return
					}
				}
			}
		}(orientation)	
	}

	wg.Wait()
	return
}

func ReducePointSets(pointSets []*PointSet) (reduced, origins *PointSet) {
	// Keep track of the origins of the original point sets.
	// This corresponds to the scanner coordinates relative to the first scanner.
	origins = NewPointSet(len(pointSets))

	reduced, pointSets = pointSets[0], pointSets[1:]

	// The scanner from the first PointSet is our 'true' origin. All other scanner
	// locations are relative to this one.
	origins.Add(NewPoint(0, 0, 0))

	var cache TranslationCache

	for len(pointSets) > 0 {
		var  (
			matchingCandidateOrientation, candidate *PointSet
			origin *Point
			i int
		)

		for i, candidate = range pointSets {
			matchingCandidateOrientation, origin = TryMatchOrientations(reduced, candidate, &cache)
			if matchingCandidateOrientation != nil {
				break
			}
		}

		if matchingCandidateOrientation == nil {
			panic("could not reduce")	
		}

		reduced.Union(matchingCandidateOrientation)
		origins.Add(origin)

		// Splice element i out of the slice of point sets.
		pointSets = append(pointSets[:i], pointSets[i+1:]...)
		fmt.Printf("reduced to %d sets\n", len(pointSets)+1)
	}

	return reduced, origins
}

func main() {
	pointSets := loadPointSets()
	fmt.Printf("loaded %d scanners\n", len(pointSets))
	for i, pointSet := range pointSets {
		fmt.Printf("scanner %d has %d beacons\n", i, pointSet.Size())
	}

	beaconPointSet, scannerPointSet := ReducePointSets(pointSets)
	fmt.Printf("count of unique beacons: %d\n", beaconPointSet.Size())

	maxScannerDistance := 0
	for i, a := range scannerPointSet.Points[:scannerPointSet.Size()-1] {
		for _, b := range scannerPointSet.Points[i+1:] {
			distance := a.ManhattanDistance(b)
			if distance > maxScannerDistance {
				maxScannerDistance = distance
			}
		}
	}
	fmt.Printf("max scanner distace: %d\n", maxScannerDistance)
}
