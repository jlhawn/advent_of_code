package main

import (
	"bufio"
	"bytes"
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

func parseSnailfishNumberPair(byteSequence *bytes.Buffer, parent *SnailfishNumber) *SnailfishNumber {
	number := &SnailfishNumber{
		Parent: parent,
	}

	if string(byteSequence.Next(1)) != "[" {
		panic("expected open bracket")
	}

	next := string(byteSequence.Next(1))
	if err := byteSequence.UnreadByte(); err != nil { panic(err) }

	if next == "[" {
		// Parse the left pair.
		number.Left = parseSnailfishNumberPair(byteSequence, number)
	} else {
		// Parse the left integer.
		number.Left = parseSnailfishNumberInteger(byteSequence, number, ',')
	}

	if string(byteSequence.Next(1)) != "," {
		panic("expected comma")
	}

	next = string(byteSequence.Next(1))
	if err := byteSequence.UnreadByte(); err != nil { panic(err) }

	if next == "[" {
		// Parse the right pair.
		number.Right = parseSnailfishNumberPair(byteSequence, number)
	} else {
		// Parse the right integer.
		number.Right = parseSnailfishNumberInteger(byteSequence, number, ']')
	}

	if string(byteSequence.Next(1)) != "]" {
		panic("expected open bracket")
	}

	return number
}

func parseSnailfishNumberInteger(byteSequence *bytes.Buffer, parent *SnailfishNumber, delim byte) *SnailfishNumber {
	rawInt, err := byteSequence.ReadString(delim)
	if err != nil { panic(err) }
	if err := byteSequence.UnreadByte(); err != nil { panic(err) }

	val, err := strconv.Atoi(rawInt[:len(rawInt)-1])
	if err != nil { panic(err) }

	return &SnailfishNumber{
		Parent: parent,
		Integer: val,
	}
}

func loadSnailfishNumbers() []*SnailfishNumber {
	rawLines := readInputLines()

	numbers := make([]*SnailfishNumber, len(rawLines))
	for i, rawLine := range rawLines {
		byteSequence := bytes.NewBufferString(rawLine)
		numbers[i] = parseSnailfishNumberPair(byteSequence, nil)
	}

	return numbers
}

type SnailfishNumber struct {
	Parent *SnailfishNumber

	Left *SnailfishNumber
	Right *SnailfishNumber

	Integer int
}

func (n *SnailfishNumber) Clone() *SnailfishNumber {
	clone := &SnailfishNumber{
		Parent: n.Parent,
		Integer: n.Integer,
	}

	if n.IsPair() {
		clone.Left = n.Left.Clone()
		clone.Right = n.Right.Clone()

		clone.Left.Parent = clone
		clone.Right.Parent = clone
	}

	return clone
}

func (n *SnailfishNumber) IsPair() bool {
	return !n.IsInteger()
}

func (n *SnailfishNumber) IsInteger() bool {
	return n.Left == nil && n.Right == nil	
}

func (n *SnailfishNumber) buildString(b *strings.Builder) {
	if n.IsInteger() {
		b.WriteString(strconv.Itoa(n.Integer))
		return
	}

	b.WriteString("[")
	n.Left.buildString(b)
	b.WriteString(",")
	n.Right.buildString(b)
	b.WriteString("]")
}

func (n *SnailfishNumber) String() string {
	var b strings.Builder
	n.buildString(&b)
	return b.String()
}

func (n *SnailfishNumber) Add(o *SnailfishNumber) *SnailfishNumber {
	p := &SnailfishNumber{
		Left: n,
		Right: o,
	}
	n.Parent = p
	o.Parent = p

	p.Reduce()
	return p
}

func (n *SnailfishNumber) Reduce() {
	for {
		if explodable := n.LeftmostPairAtDepth(4); explodable != nil {
			explodable.Explode()
			continue
		}

		if splitable := n.LeftmostIntegerGreaterThan(9); splitable != nil {
			splitable.Split()
			continue
		}

		break
	}
}

func (n *SnailfishNumber) LeftmostPairAtDepth(depth int) *SnailfishNumber {
	if n.IsInteger() {
		return nil
	}

	if depth == 0 {
		return n
	}

	if found := n.Left.LeftmostPairAtDepth(depth-1); found != nil {
		return found
	}

	if found := n.Right.LeftmostPairAtDepth(depth-1); found != nil {
		return found
	}

	return nil
}

func (n *SnailfishNumber) Explode() {
	// This must be a pair of integers.
	if n.IsInteger() {
		panic("cannot explode an integer value")
	}
	if n.Left.IsPair() || n.Right.IsPair()  {
		panic("cannot explode a pair which contains nested pairs")
	}
	// This must be a node with a parent.
	if n.Parent == nil {
		panic("cannot explode a pair without a parent")
	}

	// Add the left value to the next number to the left in the tree.
	n.AddNextLeftValue(n.Left.Integer)

	// Add the right value to the next number to the right in the tree.
	n.AddNextRightValue(n.Right.Integer)

	// Replace this node with a 0 integer value.
	n.Left = nil
	n.Right = nil
	n.Integer = 0
	
}

func (n *SnailfishNumber) AddNextLeftValue(val int) {
	// Go up until we find an ancestor with a left node whe can
	// traverse to.
	node := n
	for node.Parent != nil && node.Parent.Left == node {
		node = node.Parent
	}

	if node.Parent == nil {
		// No values to the left.
		return
	}

	// Step to the left sibling.
	node = node.Parent.Left

	// Go down to the right until we find an integer node.
	for node.IsPair() {
		node = node.Right
	}

	node.Integer += val
}

func (n *SnailfishNumber) AddNextRightValue(val int) {
	// Go up until we find an ancestor with a right node whe can
	// traverse to.
	node := n
	for node.Parent != nil && node.Parent.Right == node {
		node = node.Parent
	}

	if node.Parent == nil {
		// No values to the right.
		return
	}

	// Step to the right sibling.
	node = node.Parent.Right

	// Go down to the left until we find an integer node.
	for node.IsPair() {
		node = node.Left
	}

	node.Integer += val
}

func (n *SnailfishNumber) LeftmostIntegerGreaterThan(val int) *SnailfishNumber {
	if n.IsInteger() {
		if n.Integer > val {
			return n
		}
		return nil
	}

	if found := n.Left.LeftmostIntegerGreaterThan(val); found != nil {
		return found
	}

	if found := n.Right.LeftmostIntegerGreaterThan(val); found != nil {
		return found
	}

	return nil
}

func (n *SnailfishNumber) Split() {
	if n.IsPair() {
		panic("cannot split a pair node")
	}

	n.Left = &SnailfishNumber{
		Parent: n,
		Integer: n.Integer/2,
	}
	n.Right = &SnailfishNumber{
		Parent: n,
		Integer: n.Integer/2 + (n.Integer%2),
	}

	n.Integer = 0
}

func (n *SnailfishNumber) Magnitude() int {
	if n.IsInteger() {
		return n.Integer
	}

	return 3*n.Left.Magnitude() + 2*n.Right.Magnitude()
}

func main() {
	numbers := loadSnailfishNumbers()
	fmt.Printf("loaded %d snailfish numbers\n", len(numbers))

	a := numbers[0].Clone()
	for _, b := range numbers[1:] {
		a = a.Add(b.Clone())
	}
	fmt.Println(a)
	fmt.Printf("magnitude of sum of all numbers: %d\n", a.Magnitude())

	// Iterate through all pairs of numbers to find the max magnitude
	// when adding any two pairs of numbers.
	maxMagnitude := 0
	for i := 0; i < len(numbers); i++ {
		for j := i+1; j < len(numbers); j++ {
			a, b := numbers[i].Clone(), numbers[j].Clone()
			if aPlusB := a.Add(b).Magnitude(); aPlusB > maxMagnitude {
				maxMagnitude = aPlusB
			}

			a, b = numbers[i].Clone(), numbers[j].Clone()
			if bPlusA := b.Add(a).Magnitude(); bPlusA > maxMagnitude {
				maxMagnitude = bPlusA
			}
		}
	}
	fmt.Printf("max magnitude of adding any two pairs: %d\n", maxMagnitude)
}
