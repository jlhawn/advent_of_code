package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
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

type IntReader interface {
	ReadInt() (int, error)
}

type SliceIntReader struct {
	Ints []int
	Index int
}

func (r *SliceIntReader) ReadInt() (int, error) {
	if r.Index >= len(r.Ints) {
		return 0, io.EOF
	}

	val := r.Ints[r.Index]
	r.Index++

	return val, nil
}

type Instruction struct {
	Op string
	Arg1, Arg2 *int
}

type ALU struct {
	Input IntReader
	Instructions []Instruction

	W, X, Y, Z int
}

func (a *ALU) PrintState() {
	fmt.Printf("W=%-3d X=%-3d Y=%-3d Z=%-3d\n", a.W, a.X, a.Y, a.Z)
}

func (a *ALU) Reset() {
	a.W, a.X, a.Y, a.Z = 0, 0, 0, 0
}

func Inp(input IntReader, a *int) {
	val, err := input.ReadInt()
	if err != nil { panic(err) }
	*a = val
}

func Add(a, b *int) {
	*a += *b
}

func Mul(a, b *int) {
	*a *= *b
}

func Div(a, b *int) {
	*a /= *b
}

func Mod(a, b *int) {
	*a %= *b
}

func Eql(a, b *int) {
	if *a == *b {
		*a = 1
	} else {
		*a = 0
	}
}

func (a *ALU) LoadInstructions() {
	lines := readInputLines()
	a.Instructions = make([]Instruction, len(lines))
	for i, line := range lines {
		parts := strings.Split(line, " ")
		
		// Get the op code.
		a.Instructions[i].Op = parts[0]

		// Get the first argument.
		switch parts[1] {
		case "w":
			a.Instructions[i].Arg1 = &a.W
		case "x":
			a.Instructions[i].Arg1 = &a.X
		case "y":
			a.Instructions[i].Arg1 = &a.Y
		case "z":
			a.Instructions[i].Arg1 = &a.Z
		default:
			panic("first arg must be one of w, x, y, or z")
		}

		// Get the second argument if there is one.
		if len(parts) > 2 {
			switch parts[2] {
			case "w":
				a.Instructions[i].Arg2 = &a.W
			case "x":
				a.Instructions[i].Arg2 = &a.X
			case "y":
				a.Instructions[i].Arg2 = &a.Y
			case "z":
				a.Instructions[i].Arg2 = &a.Z
			default:
				// Must be an integer literal.
				a.Instructions[i].Arg2 = new(int) // Important! Allocate a new int on the heap.
				var err error
				*a.Instructions[i].Arg2, err = strconv.Atoi(parts[2])
				if err != nil { panic(err) }
			}	
		}
	}
}

func (a *ALU) Execute(input []int) {
	a.Input = &SliceIntReader{Ints: input}

	for _, inst := range a.Instructions {
		switch inst.Op {
		case "inp":
			a.PrintState()
			Inp(a.Input, inst.Arg1)
		case "add":
			Add(inst.Arg1, inst.Arg2)
		case "mul":
			Mul(inst.Arg1, inst.Arg2)
		case "div":
			Div(inst.Arg1, inst.Arg2)
		case "mod":
			Mod(inst.Arg1, inst.Arg2)
		case "eql":
			Eql(inst.Arg1, inst.Arg2)
		}
	}
	a.PrintState()
}

type MysteryArgs struct {
	A, B, C int
}

func (alu *ALU) Mystery(args MysteryArgs) {
	digit, _ := alu.Input.ReadInt()

	if (alu.Z % 26) + args.B == digit {
		alu.Z = (alu.Z / args.A)
	} else {
		alu.Z = ((alu.Z / args.A) * 26) + digit + args.C
	}
}

// Mystery is the mystery function which we're trying to
// understand. It has been reduced to a previous state
// zPrev, and inputs: digit, A, B, and C. Returns the
// next state zNext.
func Mystery(zPrev, digit int, args MysteryArgs) int {
	if MysteryCondition(zPrev, digit, args.B) {
		return zPrev / args.A
	} else {
		return ((zPrev / args.A) * 26) + digit + args.C
	}
}

// MysteryCondition is the conditional test used in the
// Mystery function above.
func MysteryCondition(zPrev, digit, B int) bool {
	return (zPrev % 26) + B == digit
}

// Untrunc returns all the possible dividends which could be divided by the
// divisor to get the resulting quotient under truncated integer division.
func Untrunc(quotient, divisor int) []int {
	dividends := make([]int, divisor)
	min := quotient * divisor
	for i := 0; i < divisor; i++ {
		dividends[i] = min + i
	}
	return dividends
}

// Unmult checks whether the product is a possible multiple of the
// given multiplier and, if so, returns the expected multiplicand.
// Specifically, the product is not possible if it is not evenly
// divided by the multiplier.
func Unmult(product, multiplier int) (int, bool) {
	if product % multiplier == 0 {
		return product / multiplier, true
	}
	return 0, false
}

// Wystery is the inverse of Mystery. It returns all the possible zPrev values
// which could have resulted in the given zNext value being returned by Mystery
// with matching arguments.
func Wystery(zNext, digit int, args MysteryArgs) []int {
	zPrevs := slices.Filter(func(zPrev int) bool {
		return MysteryCondition(zPrev, digit, args.B)
	}, Untrunc(zNext, args.A)...)

	if multiplicand, ok := Unmult(zNext - digit - args.C, 26); ok {
		zPrevs = append(zPrevs, slices.Filter(func(zPrev int) bool {
			return !MysteryCondition(zPrev, digit, args.B)
		}, Untrunc(multiplicand, args.A)...)...)
	}

	return zPrevs
}

type ModelNumber struct {
	Digit int
	Next *ModelNumber
}

func (n *ModelNumber) String() string {
	var b strings.Builder
	for n != nil {
		b.WriteString(strconv.Itoa(n.Digit))
		n = n.Next
	}
	return b.String()
}

func (n *ModelNumber) Less(o *ModelNumber) bool {
	for n != nil && o != nil {
		if n.Digit < o.Digit {
			return true
		}
		n, o = n.Next, o.Next
	}
	return false
}

func (n *ModelNumber) Ints() []int {
	ints := make([]int, 0, 14)
	for n != nil {
		ints = append(ints, n.Digit)
		n = n.Next
	}
	return ints
}

// FindValidModelNumbers returns a slice of valid model numbers.
func FindValidModelNumbers(zNext int, reversedArgs []MysteryArgs, modelNumberSoFar *ModelNumber) (foundModelNumbers []*ModelNumber, ok bool) {
	if len(reversedArgs) == 0 {
		// We've reached the end of the recursive search. zNext is really
		// the initial Z value and *must* be zero to be valid.
		return []*ModelNumber{modelNumberSoFar}, zNext == 0 // Signals whether this model number is valid.
	}

	args, reversedArgs := reversedArgs[0], reversedArgs[1:]

	for digit := 1; digit <= 9; digit++ {
		extendedModelNumber := &ModelNumber{Digit: digit, Next: modelNumberSoFar}

		prevOptions := Wystery(zNext, digit, args)
		for _, zPrev := range prevOptions {
			// Assert that we do in fact get this zNext value if we run the Mystery function.
			if Mystery(zPrev, digit, args) != zNext {
				panic(fmt.Errorf("Mystery(%d, %d, %#v) != %d", zPrev, digit, args, zNext))
			}

			modelNumbers, ok := FindValidModelNumbers(zPrev, reversedArgs, extendedModelNumber)
			if !ok {
				continue // No valid model numbers with this zPrev option.
			}

			for _, modelNumber := range modelNumbers {
				foundModelNumbers = append(foundModelNumbers, modelNumber)
			}

			// As soon as we find a valid model number for this digit value, we can break the loop
			// to continue to the next possible digit value.
			break
		}
	}

	return foundModelNumbers, len(foundModelNumbers) > 0
}

func main() {
	// These were found by looking at patterns in the input.
	mysteryArgs := []MysteryArgs{
		{1, 14, 1},
		{1, 15, 7},
		{1, 15, 13},
		{26, -6, 10},
		{1, 14, 0},
		{26, -4, 13},
		{1, 15, 11},
		{1, 15, 6},
		{1, 11, 1},
		{26, 0, 7},
		{26, 0, 11},
		{26, -3, 14},
		{26, -9, 4},
		{26, -9, 10},
	}

	slices.Reverse(mysteryArgs)
	foundModelNumbers, _ := FindValidModelNumbers(0, mysteryArgs, nil)
	slices.Reverse(mysteryArgs)

	fmt.Printf("Found %d valid model numbers.\n", len(foundModelNumbers))

	maxModelNumber := slices.MaxFunc(func(a, b *ModelNumber) bool { return a.Less(b) }, foundModelNumbers...)
	fmt.Printf("Largest model number: %s\n", maxModelNumber)		

	var alu ALU
	alu.LoadInstructions()
	alu.Execute(maxModelNumber.Ints())

	minModelNumber := slices.MinFunc(func(a, b *ModelNumber) bool { return a.Less(b) }, foundModelNumbers...)
	fmt.Printf("Smallest model number: %s\n", minModelNumber)

	alu.Reset()
	alu.Execute(minModelNumber.Ints())
}
