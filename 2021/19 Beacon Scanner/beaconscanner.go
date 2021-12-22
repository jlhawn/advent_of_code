package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
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

func parsePoint(line string) Point {
	parts := strings.Split(line, ",")
	x, err := strconv.Atoi(parts[0])
	if err != nil { panic(err) }
	y, err := strconv.Atoi(parts[1])
	if err != nil { panic(err) }
	z, err := strconv.Atoi(parts[2])
	if err != nil { panic(err) }

	return Point{x, y, z}
}

type Point struct {
	X, Y, Z int
}

// Add returns a point which is the vector sum
// of this point and the given point.
func (p Point) Add(o Point) Point {
	return Point{p.X+o.X, p.Y+o.Y, p.Z+o.Z}
}

// Subtract returns a point which is the vector subtraction
// of this point and the given point.
func (p Point) Subtract(o Point) Point {
	return Point{p.X-o.X, p.Y-o.Y, p.Z-o.Z}
}

// Rotate90X returns this point rotated 90 degrees about the X axis.
func (p Point) Rotate90X() Point {
	return Point{X: p.X, Y: -p.Z, Z: p.Y}
}

// Rotate90Y returns this point rotated 90 degrees about the Y axis.
func (p Point) Rotate90Y() Point {
	return Point{X: p.Z, Y: p.Y, Z: -p.X}
}

// Rotate90Z returns this point rotated 90 degrees about the Z axis.
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

func (p Point) MagnitudeSquared() int {
	return p.X*p.X + p.Y*p.Y + p.Z*p.Z
}

type PointSet struct {
	Points []Point
	Set map[Point]bool

	// A cache of 24 different orientations of this PointSet.
	OrientationsCache []*PointSet

	// A set of point fingerprints for each point in this set.
	// Must call GenerateFingerprints() to initialize.
	Fingerprints map[Point]PointFingerprint
}

// NewPointSet creates a new empty PointSet with the given capacity hint.
// It is able to grow its capacity as needed.
func NewPointSet(cap int) *PointSet {
	return &PointSet{
		Points: make([]Point, 0, cap),
		Set: make(map[Point]bool, cap),
	}
}

// Size returns the number of elements in this set.
func (s *PointSet) Size() int {
	return len(s.Set)
}

func (s *PointSet) Contains(p Point) bool {
	return s.Set[p]
}

// Add inserts the given point into this point set if it is
// not in the set already.
func (s *PointSet) Add(p Point) {
	if s.Contains(p) {
		return // The point is already in the set.
	}

	s.Points = append(s.Points, p)
	s.Set[p] = true
}

// Clone creates a new PointSet which is a copy of this one.
// The cloned set will *NOT* have initialized
// OrientationsCache or Fingerprints values.
func (s *PointSet) Clone() *PointSet {
	cloned := NewPointSet(s.Size())
	for _, p := range s.Points {
		cloned.Add(p)
	}
	return cloned
}

// Remap modifies every point in this set with the given remap func.
// The remap function must be bijective, i.e., an invertible,
// one-to-one function which operates on a point and returns a new
// point. No two points must map to the same point.
// Returns itself to allow chaining.
func (s *PointSet) Remap(remapFunc func(Point) Point) *PointSet {
	for i, p := range s.Points {
		delete(s.Set, p)
		s.Points[i] = remapFunc(p)
	}

	for _, p := range s.Points {
		s.Set[p] = true
	}

	return s
}

// Translate is a remap with an Add function.
// Returns itself to allow chaining.
func (s *PointSet) Translate(v Point) *PointSet {
	return s.Remap(func(p Point) Point { return p.Add(v) })
}

// Rotate90X is a remap which rotates each point 90 degrees about the x axis.
// Returns itself to allow chaining.
func (s *PointSet) Rotate90X() *PointSet {
	return s.Remap(func(p Point) Point { return p.Rotate90X() })
}

// Rotate90Y is a remap which rotates each point 90 degrees about the y axis.
// Returns itself to allow chaining.
func (s *PointSet) Rotate90Y() *PointSet {
	return s.Remap(func(p Point) Point { return p.Rotate90Y() })
}

// Rotate90Z is a remap which rotates each point 90 degrees about the z axis.
// Returns itself to allow chaining.
func (s *PointSet) Rotate90Z() *PointSet {
	return s.Remap(func(p Point) Point { return p.Rotate90Z() })
}

// Union adds every point from the given set to this set.
// Returns itself to allow chaining.
func (s *PointSet) Union(o *PointSet) *PointSet {
	for _, p := range o.Points {
		s.Add(p)
	}
	return s
}

// IntersectionSize returns the number of points in the given point set which
// are also elements of this point set.
func (s *PointSet) IntersectionSize(o *PointSet) int {
	largerSet, smallerSet := s, o
	if smallerSet.Size() > largerSet.Size() {
		largerSet, smallerSet = o, s
	}

	count := 0
	for _, p := range smallerSet.Points {
		if largerSet.Contains(p) {
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
func (s *PointSet) Orientations() []*PointSet {
	if s.OrientationsCache != nil {
		return s.OrientationsCache
	}

	for it := (&OrientationIterator{PointSet: s.Clone()}); !it.IsEnd(); it.Next() {
		s.OrientationsCache = append(s.OrientationsCache, it.PointSet.Clone())
	}

	return s.OrientationsCache
}

// A PointFingerprinter is used to compute a fingerprint for a reference
// point using the the distances from that point to a set of other points.
// The fingerprint can be used for matching a point with another point which
// was fingerprinted with a different origin.
type PointFingerprinter struct {
	Reference Point
	DistanceSquares map[Point]int
}

func NewPointFingerprinter(reference Point, cap int) *PointFingerprinter {
	return &PointFingerprinter{
		Reference: reference,
		DistanceSquares: make(map[Point]int, cap),
	}
}

func (f *PointFingerprinter) Add(p Point) {
	// Compute the vector from the reference point to p.
	refToP := p.Subtract(f.Reference)

	// Using the square of the distances means I don't have to do
	// any square roots which are more computationally expensive.
	f.DistanceSquares[p] = refToP.MagnitudeSquared()
}

type PointFingerprint []int

func (f *PointFingerprinter) Fingerprint() PointFingerprint {
	fingerprint := make([]int, 0, len(f.DistanceSquares))
	for _, distanceSquared := range f.DistanceSquares {
		fingerprint = append(fingerprint, distanceSquared)
	}
	sort.Ints(fingerprint)
	return PointFingerprint(fingerprint)
}

// Matches returns whether or not the given fingerprint is a match with this
// fingerprint with at least threshold many of the same distances from their
// reference point, which means they are the same point from a different
// reference origin.
func (f PointFingerprint) Matches(g PointFingerprint, threshold int) bool {
	var i, j, matches int
	for i < len(f) && j < len(g) {
		switch {
		case f[i] < g[j]:
			i++
		case f[i] > g[j]:
			j++
		default: // f[i] == g[j]
			i++; j++; matches++
			if matches == threshold {
				return true
			}
		}
	}
	return false
}

func (f PointFingerprint) Equals(g PointFingerprint) bool {
	if len(f) != len(g) {
		return false
	}

	for i := range f {
		if f[i] != g[i] {
			return false
		}
	}

	return true
}

func (s *PointSet) GenerateFingerprints() {
	s.Fingerprints = make(map[Point]PointFingerprint, s.Size())
	for i, p := range s.Points {
		fingerprinter := NewPointFingerprinter(p, s.Size()-1)
		for j, q := range s.Points {
			if i != j { // Don't include the point itself.
				fingerprinter.Add(q)
			}
		}
		s.Fingerprints[p] = fingerprinter.Fingerprint()
	}
}

func (s *PointSet) FindEqualPoint(fingerprint PointFingerprint) (Point, bool) {
	for _, q := range s.Points {
		if s.Fingerprints[q].Equals(fingerprint) {
			return q, true
		}
	}
	return Point{}, false
}

func FindMatchingPoints(setA, setB *PointSet) (pointA, pointB Point, found bool) {
	for _, pointA = range setA.Points {
		// We don't need to check the last 11 points in set B because if there is
		// a match then there would be a match with at least 12 points. If there
		// are 11 left then one of the previous points would have been a match.
		for _,  pointB = range setB.Points[:setB.Size()-11] {
			fingerprintA := setA.Fingerprints[pointA]
			fingerprintB := setB.Fingerprints[pointB]

			// If it's the same point then the fingerprints
			// should match with at least 11 other points.
			if fingerprintA.Matches(fingerprintB, 11) {
				return pointA, pointB, true
			}
		}
	}

	return Point{}, Point{}, false
}

// TryMatchOrientations will attempt to patch the given candidate PointSet with
// the given reduced PointSet. If there is a translation of any orientation of
// the candidate PointSet which contains at least 12 of the same points then it
// is a match and that orientation PointSet and translation vector will be
// returned. If there is no match, returns (nil, nil).
//
// This function does not mutate the given PoinSets.
func TryMatchOrientations(reduced, candidate *PointSet) (match *PointSet, origin Point) {
	reducedPoint, candidatePoint, found := FindMatchingPoints(reduced, candidate)
	if !found {
		// The reduced set and candidate set don't have any points in common.
		// Don't bother looking for a matching orientation and translation.
		return
	}

	var wg sync.WaitGroup
	for _, orientation := range candidate.Orientations() {
		wg.Add(1)
		go func(orientation *PointSet) {
			defer wg.Done()

			if orientation.Fingerprints == nil {
				// Orientations are cached so the fingerprints will be cached, too.
				orientation.GenerateFingerprints()
			}

			candidatePointFingerprint := candidate.Fingerprints[candidatePoint]
			orientationPoint, ok := orientation.FindEqualPoint(candidatePointFingerprint)
			if !ok { panic("unable to find equal fingerprint point between orientations") }

			// We know from the matching fingerprint that these *are* the same point but
			// we need to check whether we are in the correct orientation for the candidate.

			translation := reducedPoint.Subtract(orientationPoint)
			translated := orientation.Clone().Translate(translation)
			if reduced.IntersectionSize(translated) >= 12 {
				match, origin = translated, translation
			}
		}(orientation)
	}

	wg.Wait()
	return
}

func ReducePointSetsByMatchingOrientations(scanners []*PointSet) (reduced, origins *PointSet) {
	// Keep track of the origins of the scanners relative to our canonical scanner.
	// This corresponds to the scanner coordinates relative to the first scanner.
	origins = NewPointSet(len(scanners))

	reduced, scanners = scanners[0], scanners[1:]

	// The scanner from the first PointSet is our 'true' origin. All other scanner
	// locations are relative to this one.
	origins.Add(Point{0, 0, 0})

	// Generate fingerprints for all scanner data.
	for _, scanner := range scanners {
		scanner.GenerateFingerprints()
	}

	for len(scanners) > 0 {
		var  (
			matchingCandidateOrientation, candidate *PointSet
			origin Point
			i int
		)

		// Regenerate fingerprints for the reduced set of beacons since we have
		// added new points from the previous iteration.
		reduced.GenerateFingerprints()

		for i, candidate = range scanners {
			matchingCandidateOrientation, origin = TryMatchOrientations(reduced, candidate)
			if matchingCandidateOrientation != nil {
				break
			}
		}

		if matchingCandidateOrientation == nil {
			panic("could not reduce")	
		}

		reduced.Union(matchingCandidateOrientation)
		origins.Add(origin)

		// Splice element i out of the slice of scanners.
		scanners = append(scanners[:i], scanners[i+1:]...)
		fmt.Printf("reduced to %d sets\n", len(scanners)+1)
	}

	return reduced, origins
}

func main() {
	pointSets := loadPointSets()
	fmt.Printf("loaded %d scanners\n", len(pointSets))
	for i, pointSet := range pointSets {
		fmt.Printf("scanner %d has %d beacons\n", i, pointSet.Size())
	}

	beaconPointSet, scannerPointSet := ReducePointSetsByMatchingOrientations(pointSets)
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
