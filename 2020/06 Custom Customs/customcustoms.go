package main

import (
	"bufio"
	// "constraints"
	"errors"
	"fmt"
	"io"
	"os"
	// "strconv"
	"strings"
	// "time"

	// "../../slices"
	"../../streams"
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

type DeclarationForm map[byte]bool

func loadDeclarationFormsPart1() []DeclarationForm {
	var allForms []DeclarationForm
	currentForm := DeclarationForm{}
	for _, line := range readInputLines() {
		if len(line) == 0 {
			// End of a passenger group.
			allForms = append(allForms, currentForm)
			currentForm = DeclarationForm{}
			continue
		}
		// Each line list the questions that one member of a group
		// answered 'yes' to.
		for _, q := range []byte(line) {
			currentForm[q] = true
		}
	}
	return append(allForms, currentForm)
}

func loadDeclarationFormsPart2() []DeclarationForm {
	var allForms []DeclarationForm
	currentForm := DeclarationForm{}
	groupSize := 0
	for _, line := range readInputLines() {
		if len(line) == 0 {
			// End of a passenger group.
			allForms = append(allForms, currentForm)
			currentForm = DeclarationForm{}
			groupSize = 0
			continue
		}
		if groupSize == 0 {
			// First person in this group.
			for _, q := range []byte(line) {
				currentForm[q] = true
			}
		} else {
			// Subsequent members of the group. We only want the
			// intersection.
			intersection := DeclarationForm{}
			for _, q := range []byte(line) {
				if currentForm[q] {
					intersection[q] = true
				}
			}
			currentForm = intersection
		}
		groupSize++
	}
	return append(allForms, currentForm)
}

func main() {
	allForms := loadDeclarationFormsPart1()
	sumOfYesCounts := streams.SumFunc(streams.FromItems(allForms...), func(form DeclarationForm) int { return len(form) })
	fmt.Printf("Sum of Yes Counts for each group (Part 1): %d\n", sumOfYesCounts)

	allForms = loadDeclarationFormsPart2()
	sumOfYesCounts = streams.SumFunc(streams.FromItems(allForms...), func(form DeclarationForm) int { return len(form) })
	fmt.Printf("Sum of Yes Counts for each group (Part 2): %d\n", sumOfYesCounts)
}
