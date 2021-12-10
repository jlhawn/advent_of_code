package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
)

const inputFilename = "./INPUT"

func readInputLanternfish() (map[int]int, error) {
	rawLines, err := readInputLines()
	if err != nil {
		return nil, fmt.Errorf("unable to read input lines: %w", err)
	}

	if len(rawLines) != 1 {
		return nil, fmt.Errorf("there should be 1 input line but got %d", len(rawLines))
	}

	rawInts := strings.Split(rawLines[0], ",")

	populationCounts := make(map[int]int)
	for _, rawInt := range rawInts {
		spawnTimer, err := strconv.Atoi(rawInt)
		if err != nil {
			return nil, fmt.Errorf("unable to parse input timer value %q: %w", rawInt, err)
		}
		populationCounts[spawnTimer]++
	}

	return populationCounts, nil
}

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

func simulatePopulation(populationCounts map[int]int, days int) {
	for day := 0; day < days; day++ {
		var totalPop int
		for _, population := range populationCounts {
			totalPop += population
		}
		fmt.Printf("Day %3d: %d\n", day, totalPop)

		newPopulationCounts := make(map[int]int)
		for spawnTimer, population := range populationCounts {
			if spawnTimer == 0 {
				newPopulationCounts[6] += population
				newPopulationCounts[8] += population
			} else {
				newPopulationCounts[spawnTimer-1] += population
			}
		}

		populationCounts = newPopulationCounts
	}
	var finalPop int
	for _, population := range populationCounts {
		finalPop += population
	}
	fmt.Printf("Day %3d: %d\n", days, finalPop)
}

func main() {
	populationCounts, err := readInputLanternfish()
	if err != nil {
		log.Fatal(err)
	}

	simulatePopulation(populationCounts, 256)
}
