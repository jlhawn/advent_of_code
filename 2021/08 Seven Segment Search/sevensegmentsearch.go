package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	// "strconv"
	"strings"
)

const inputFilename = "./INPUT"

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

type SegmentReading byte

const ( // Segment Flags
	SegmentA SegmentReading = 1 << iota
	SegmentB
	SegmentC
	SegmentD
	SegmentE
	SegmentF
	SegmentG
)

const SegmentMask SegmentReading = SegmentA | SegmentB | SegmentC | SegmentD | SegmentE | SegmentF | SegmentG

func MakeSegmentReading(pattern string) SegmentReading {
	// Assume the pattern contains only lowercase values a through g.
	var reading SegmentReading
	for _, val := range pattern {
		switch val {
		case 'a':
			reading |= SegmentA
		case 'b':
			reading |= SegmentB
		case 'c':
			reading |= SegmentC
		case 'd':
			reading |= SegmentD
		case 'e':
			reading |= SegmentE
		case 'f':
			reading |= SegmentF
		case 'g':
			reading |= SegmentG
		}
	}
	return reading
}

/*

 Digit |  No. of Segments (* unique)
-------|-----------------------------
     0 | 6
     1 | 2 *
     2 | 5
     3 | 5
     4 | 4 *
     5 | 5
     6 | 6
     7 | 3 *
     8 | 7 *
     9 | 6

*/
func (r SegmentReading) IsOneFourSevenOrEight() bool {
	switch r.SegmentsLit() {
	case 2, 4, 3, 7: // Correspond to 1, 4, 7, 8
		return true
	}
	return false
}

func (r SegmentReading) SegmentsLit() int {
	bitCount := 0
	copiedReading := byte(r)
	for copiedReading != 0 {
		if copiedReading % 2 == 1 {
			bitCount++
		}
		copiedReading >>= 1
	}
	return bitCount
}

func (r SegmentReading) Intersect(other SegmentReading) SegmentReading {
	return r & other
}

func (r SegmentReading) Subtract(other SegmentReading) SegmentReading {
	common := r & other
	return r & (^common) & SegmentMask
}

func (r SegmentReading) Combine(others ...SegmentReading) SegmentReading {
	combined := r
	for _, other := range others {
		combined |= other
	}
	return combined
}

func (r SegmentReading) Not() SegmentReading {
	return (^r) & SegmentMask
}

type DisplayNote struct {
	SignalPatterns [10]SegmentReading
	OutputDigits  [4]SegmentReading

	OutputValue int
}

/*
 
      0:      1:      2:      3:      4:
     aaaa    ....    aaaa    aaaa    ....
    b    c  .    c  .    c  .    c  b    c
    b    c  .    c  .    c  .    c  b    c
     ....    ....    dddd    dddd    dddd
    e    f  .    f  e    .  .    f  .    f
    e    f  .    f  e    .  .    f  .    f
     gggg    ....    gggg    gggg    ....
    
      5:      6:      7:      8:      9:
     aaaa    aaaa    aaaa    aaaa    aaaa
    b    .  b    .  .    c  b    c  b    c
    b    .  b    .  .    c  b    c  b    c
     dddd    dddd    ....    dddd    dddd
    .    f  e    f  .    f  e    f  .    f
    .    f  e    f  .    f  e    f  .    f
     gggg    gggg    ....    gggg    gggg

*/
func (n *DisplayNote) DetermineOutputValue() int {
	// Determine which are signal patterns correspond to 1, 4, 7, and 8.
	// These each have a unique number of segments lit and can be known
	// with certainty.
	var pattern1, pattern4, pattern7, pattern8 SegmentReading
	// Determine which patterns could be either 2, 3, or 5.
	// These each have 5 segments lit.
	var patterns235 []SegmentReading
	// Determine which patterns could be either 0, 6, or 9.
	// These each have 6 segments lit.
	var patterns069 []SegmentReading

	for _, pattern := range n.SignalPatterns {
		switch pattern.SegmentsLit() {
		case 2:
			pattern1 = pattern
		case 3:
			pattern7 = pattern
		case 4:
			pattern4 = pattern
		case 5:
			patterns235 = append(patterns235, pattern)
		case 6:
			patterns069 = append(patterns069, pattern)
		case 7:
			pattern8 = pattern
		}
	}

	// We can determine the true A segment by subtracting the 1
	// pattern from the 7 pattern.
	trueSegmentA := pattern7.Subtract(pattern1)

	// We can determine the pattern for 2 by combining 4 and 7 and
	// doing a set subtraction from the candidates for 2, 3, and 5.
	// Whichever result has 2 segments lit must be the 2 pattern.
	var pattern2 SegmentReading
	var patterns35 []SegmentReading
	for _, candidate := range patterns235 {
		if candidate.Subtract(pattern4.Combine(pattern7)).SegmentsLit() == 2 {
			pattern2 = candidate
		} else {
			patterns35 = append(patterns35, candidate)
		}
	}

	// The true B segment must be whatever segment is not lit when
	// we combine patterns 1 and 2.
	trueSegmentB := pattern1.Combine(pattern2).Not()

	// The true F segment must be whatever segment is lit when we
	// combine pattern 2 with segment B and subtract it from pattern 8.
	trueSegmentF := pattern8.Subtract(pattern2.Combine(trueSegmentB))

	// The true C segment must be whatever segment is lit when we combine
	// true segments A and F and subtract it from pattern 7.
	trueSegmentC := pattern7.Subtract(trueSegmentA.Combine(trueSegmentF))

	// The true D segment must be whatever segment is lit when we combine
	// true segments B, C, and F and subtract it from pattern 4.
	trueSegmentD := pattern4.Subtract(trueSegmentB.Combine(trueSegmentC, trueSegmentF))

	// The true E segment must be whatever segment is not lit when we combine
	// the candidate patterns for 3 and 5.
	trueSegmentE := patterns35[0].Combine(patterns35[1]).Not()

	// The true G segment isn't necessary to determine.

	// Now that we already have patterns for 1, 2, 4, 7, and 8, we need to make 0, 3, 5, 6, and 9.
	pattern0 := pattern8.Subtract(trueSegmentD)
	pattern3 := pattern2.Subtract(trueSegmentE).Combine(trueSegmentF)
	pattern5 := pattern3.Subtract(trueSegmentC).Combine(trueSegmentB)
	pattern6 := pattern5.Combine(trueSegmentE)
	pattern9 := pattern5.Combine(trueSegmentC)

	digitMap := make(map[SegmentReading]int)
	digitMap[pattern0] = 0
	digitMap[pattern1] = 1
	digitMap[pattern2] = 2
	digitMap[pattern3] = 3
	digitMap[pattern4] = 4
	digitMap[pattern5] = 5
	digitMap[pattern6] = 6
	digitMap[pattern7] = 7
	digitMap[pattern8] = 8
	digitMap[pattern9] = 9

	mult := 1000
	for _, outputDigit := range n.OutputDigits {
		n.OutputValue += mult * digitMap[outputDigit]
		mult /= 10
	}

	fmt.Printf("Output value: %d\n", n.OutputValue)
	return n.OutputValue
}

func readDisplayNotes() ([]*DisplayNote, error) {
	rawInput, err := readInputLines()
	if err != nil {
		return nil, fmt.Errorf("unable to read input lines: %w", err)
	}

	displayNotes := make([]*DisplayNote, len(rawInput))
	for i, line := range rawInput {
		parts := strings.Split(line, " ")
		if len(parts) != 15 { // 10 unique signal patterns 
			return nil, fmt.Errorf("expected 15 parts in input line but got %d: %q", len(parts), line)
		}
		// Parts 0 through 9 are the 10 unique signal patterns.
		signalPatterns := parts[:10]
		// Parts 11 through 14 are the 4 output digits.
		outputDigits := parts[11:]

		displayNotes[i] = &DisplayNote{}

		for j, signalPattern := range signalPatterns {
			displayNotes[i].SignalPatterns[j] = MakeSegmentReading(signalPattern)
		}

		for j, outputDigit := range outputDigits {
			displayNotes[i].OutputDigits[j] = MakeSegmentReading(outputDigit)
		}
	}

	return displayNotes, nil
}

func determineNumberOfUniqueOutputDigits(displayNotes []*DisplayNote) {
	uniqueCount := 0
	for _, displayNote := range displayNotes {
		for _, outputValue := range displayNote.OutputDigits {
			if outputValue.IsOneFourSevenOrEight() {
				uniqueCount++
			}
		}
	}
	fmt.Printf("There are %d unique output values\n", uniqueCount)
}

func main() {
	displayNotes, err := readDisplayNotes()
	if err != nil {
		log.Fatal(err)
	}

	determineNumberOfUniqueOutputDigits(displayNotes)

	sum := 0
	for _, displayNote := range displayNotes {
		sum += displayNote.DetermineOutputValue()
	}
	fmt.Printf("Sum of output values: %d\n", sum)
}
