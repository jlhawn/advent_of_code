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

func loadBusSchedule() (startTime int, busRoutes []int) {
	lines := readInputLines()
	startTime, err := strconv.Atoi(lines[0])
	if err != nil { panic(err) }

	rawRoutes := strings.Split(lines[1], ",")
	for _, rawRoute := range rawRoutes {
		if rawRoute == "x" {
			continue // Line is out of service.
		}
		route, err := strconv.Atoi(rawRoute)
		if err != nil { panic(err) }
		busRoutes = append(busRoutes, route)
	}

	return startTime, busRoutes
}

func DetermineNextBus(startTime int, routes []int) {
	earliestBus := routes[0]
	minWait := earliestBus - (startTime % earliestBus)
	if minWait == earliestBus {
		minWait = 0
	}
	for _, route := range routes[1:] {
		wait := route - (startTime % route)
		if wait == route {
			wait = 0
		}
		if wait < minWait {
			minWait = wait
			earliestBus = route
		}
	}
	fmt.Printf("The next bus is route %d in %d minutes. Product = %d\n", earliestBus, minWait, earliestBus*minWait)
}

func loadContestData() map[int]int64 {
	contestData := map[int]int64{}
	rawRoutes := strings.Split(readInputLines()[1], ",")
	for i, rawRoute := range rawRoutes {
		if rawRoute == "x" {
			continue
		}
		route, err := strconv.Atoi(rawRoute)
		if err != nil { panic(err) }
		contestData[i] = int64(route)
	}
	return contestData
}

// LeastPositiveResidue returns a positive value of a mod m
// for example, -2 mod 5 will return 3. -7 mod 3 will return 2
func LeastPositiveResidue(a, m int64) int64 {
	return ((a % m) + m) % m
}

func DivMod(a, b int64) (quotient, remainder int64) {
	return a/b, a%b
}

// Returns a*b (mod n) using a process that should not have intermediate
// values which are out of bounds, assuming that a, b, n are all less than
// MaxInt64/2.
func MultMod(a, b, n int64) int64 {
	// Need to stick with positive values.
	a = LeastPositiveResidue(a, n)
	b = LeastPositiveResidue(b, n)

	// Ensure a <= b
	if b < a {
		a, b = b, a
	}

	var product int64
	for a != 0 {
		if a % 2 == 1 {
			product = (product + b) % n
		}
		b = (b << 1) % n // b * 2 (mod n)
		a = a >> 1       // a / 2
	}
	return product
}

func MultManyMod(n int64, vals ...int64) int64 {
	return slices.Reduce(func(a, b int64) int64 { return MultMod(a, b, n) }, 1, vals...)
}

// Finds x and y such that: GCD(a, b) = ax + by. (By the extended euclidean algorithm)
//
// This implementation is based on
// https://en.wikibooks.org/wiki/Algorithm_Implementation/Mathematics/Extended_Euclidean_algorithm#Iterative_algorithm_3
func ExtendedGCD(a, b int64) (gcd, x, y int64) {
	if a == 0 {
		return b, 0, 1
	}

	q, r := DivMod(b, a)
	gcd, x1, y1 := ExtendedGCD(r, a)
	return gcd, y1 - q*x1, x1
}

// Represents an entry in the Extended Chinese Remainder Theorem
type CRTEntry struct {
	A, N int64
}

func (x0 CRTEntry) String() string {
	return fmt.Sprintf("x ≡ %d (mod %d)", x0.A, x0.N)
}

func (x0 CRTEntry) Solve(x1 CRTEntry) CRTEntry {
	a1, n1, a2, n2 := x0.A, x0.N, x1.A, x1.N

	gcd, m1, m2 := ExtendedGCD(n1, n2)
	if gcd != 1 {
		panic(fmt.Errorf("%d and %d are not coprime! got gcd = %d", n1, n2, gcd))
	}
	if m1*n1 + m2*n2 != gcd {
		panic(fmt.Errorf("expected %d*%d + %d*%d to be %d but it was %d", m1, n1, m2, n2, gcd, m1*n1 + m2*n2))
	}

	// Given our puzzle input, this should always fit in a 64-bit signed integer.
	n3 := n1*n2
	// But the following intermediate values might not:
	//   a3 = a1*m2*n2 + a2*m1*n1 (mod n3)
	// So we need to use a special binary multiplication
	// function which accounts for the modulus.
	a3 := MultManyMod(n3, a1, m2, n2) + MultManyMod(n3, a2, m1, n1)

	return CRTEntry{A: LeastPositiveResidue(a3, n3), N: n3}
}

func DetermineContestTime(contestData map[int]int64) {
	entries := make([]CRTEntry, 0, len(contestData))
	for i, n := range contestData {
		a := LeastPositiveResidue(int64(-i), n)
		entries = append(entries, CRTEntry{A: a, N: n})
	}

	// Sort the entries by their modulus value.
	sort.Slice(entries, func(i, j int) bool { return entries[i].N < entries[j].N })

	ans := slices.Reduce(func(x1, x2 CRTEntry) CRTEntry { return x1.Solve(x2) }, entries[0], entries[1:]...)
	fmt.Printf("Earliest timestamp such that all of the listed bus IDs depart at offsets matching their positions: %d\n", ans.A)
}

func main() {
	startTime, busRoutes := loadBusSchedule()
	DetermineNextBus(startTime, busRoutes)

	contestData := loadContestData()
	DetermineContestTime(contestData)
}
