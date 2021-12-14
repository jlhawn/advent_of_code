package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	// "sort"
	// "strconv"
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

// After solving part 1 I had to rewrite a lot of this because of
// the exponential growth in the number of elements in the polymer.
//
// Really, one pair that matches a rule creates two new pairs.
// CH -> B really is CH -> (CB, BH). We can keep track of the count
// of these pairs instead of the whole string.
//
// When we count elements later, only consider the first character
// in each pair (since the second element is also the first in a
// different pair) and consider the count of that pair. The same start
// element can occur in multiple pairs so add them together. Finally,
// add 1 for the final element in the original polymer (which we need
// need to keep track of separately).

type StringPair [2]string

func loadPolymerInstructions() (string, map[string]StringPair) {
	rawLines := readInputLines()

	polymer := rawLines[0]
	// fmt.Println(polymer)

	// Note: line 1 is empty

	insertionRules := map[string]StringPair{}
	for _, rawLine := range rawLines[2:] {
		parts := strings.Split(rawLine, " -> ")
		pair, out := parts[0], parts[1]

		out1, out2 := string(pair[0])+out, out+string(pair[1])
		// fmt.Printf("%s -> [%s, %s]\n", pair, out1, out2)
		insertionRules[pair] = StringPair{out1, out2}
	}

	return polymer, insertionRules
}

func makePairCounts(input string) map[string]int {
	asBytes := []byte(input)

	pairCounts := map[string]int{}
	for i := 0; i < len(asBytes)-1; i++ {
		pair := string(asBytes[i:i+2])
		pairCounts[pair]++
	}
	return pairCounts
}

func calculateElementCounts(pairCounts map[string]int, endElement byte) {
	countMap := map[byte]int{}

	for pair, count := range pairCounts {
		countMap[pair[0]] += count
	}
	countMap[endElement]++

	most, least := 0, -1
	for _, count := range countMap {
		if count > most {
			most = count
		}
		if count < least || least == -1 {
			least = count
		}
	}

	fmt.Printf("most common element appears %d times, least common appears %d times\nmost - least = %d\n", most, least, most - least)
}

func main() {
	polymer, rules := loadPolymerInstructions()

	endElement := polymer[len(polymer)-1]

	pairCounts := makePairCounts(polymer)

	for steps := 0; steps < 40; steps++ {
		didExpand := false

		newPairCounts := map[string]int{}

		for existingPair, existingCount := range pairCounts {
			if outPairs, ok := rules[existingPair]; ok {
				// This existing pair matches a rule.
				newPairCounts[outPairs[0]] += existingCount
				newPairCounts[outPairs[1]] += existingCount
				didExpand = true
			} else {
				// No rule for this pair.
				newPairCounts[existingPair] += existingCount
			}
		}

		pairCounts = newPairCounts

		if !didExpand {
			break
		}
	}
	calculateElementCounts(pairCounts, endElement)
}
