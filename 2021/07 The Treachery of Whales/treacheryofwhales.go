package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
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

func readCrabPositions() ([]int, error) {
	rawLines, err := readInputLines()
	if err != nil {
		return nil, fmt.Errorf("unable to read input lines: %w", err)
	}

	if len(rawLines) != 1 {
		return nil, fmt.Errorf("should only be one input line but got %d", len(rawLines))
	}

	rawVals := strings.Split(rawLines[0], ",")
	positions := make([]int, len(rawVals))
	for i, rawVal := range rawVals {
		val, err := strconv.Atoi(rawVal)
		if err != nil {
			return nil, fmt.Errorf("unable to parse value %q: %w", rawVal, err)
		}
		positions[i] = val
	}

	return positions, nil
}

func determineMinFuelUsage(crabPositions []int) {
	sort.Ints(crabPositions)
	medPos := crabPositions[len(crabPositions)/2]

	var medPosFuel int
	for _, pos := range crabPositions {
		if pos < medPos {
			medPosFuel += medPos - pos
		} else {
			medPosFuel += pos - medPos
		}
	}

	minPos, maxPos := crabPositions[0], crabPositions[len(crabPositions)-1]
	idealPos := minPos
	idealFuel := totalFuelNeeded(crabPositions, idealPos)
	for pos := minPos; pos <= maxPos; pos++ {
		totalFuel := totalFuelNeeded(crabPositions, pos)
		if totalFuel < idealFuel {
			idealPos = pos
			idealFuel = totalFuel
		}
	}

	fmt.Printf("%d fuel to median position %d\n", medPosFuel, medPos)
	fmt.Printf("%d fuel to ideal position %d\n", idealFuel, idealPos)
}

func fuelNeeded(posA, posB int) int {
	diff := posA - posB
	if diff < 0 {
		diff = -diff
	}

	return diff*(diff+1)/2
}

func totalFuelNeeded(positions []int, targetPos int) int {
	var totalFuel int
	for _, pos := range positions {
		totalFuel += fuelNeeded(pos, targetPos)
	}
	return totalFuel
}

func main() {
	positions, err := readCrabPositions()
	if err != nil {
		log.Fatal(err)
	}

	determineMinFuelUsage(positions)
}
