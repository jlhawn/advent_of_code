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

func main() {
	depths, err := getDepthReadings()
	if err != nil {
		log.Fatal(err)
	}

	countSlidingWindowIncreases(depths, 3)
}

func countSlidingWindowIncreases(depths []int, windowSize int) {
	windows := make([]int, len(depths)-windowSize+1)
	var firstWindowSum int
	for i := 0; i < windowSize; i++ {
		firstWindowSum += depths[i]		
	}
	windows[0] = firstWindowSum

	for i := 0; i < len(windows)-1; i++ {
		prevSum := windows[i]
		outVal := depths[i]
		inVal := depths[i+windowSize]
		windows[i+1] = prevSum - outVal + inVal
	}

	countSingleIncreases(windows)
}

func countSingleIncreases(depths []int) {
	var increaseCount int
	for i := 1; i < len(depths); i++ {
		if depths[i] > depths[i-1] {
			increaseCount++
		}
	}
	fmt.Printf("number of increasing measurements: %d\n", increaseCount)
}

func getDepthReadings() ([]int, error) {
	inputFile, err := os.Open(inputFilename)
	if err != nil {
		return nil, fmt.Errorf("unable to open input file: %w", err)
	}
	defer inputFile.Close()

	reader := bufio.NewReader(inputFile)
	lineNum := 1
	reading, err := reader.ReadString('\n')
	if err != nil {
		return nil, fmt.Errorf("unable to read line %d: %w", lineNum, err)
	}

	depth, err := strconv.Atoi(strings.TrimSpace(reading))
	if err != nil {
		return nil, fmt.Errorf("unable to parse line %d: %w", lineNum, err)
	}

	depths := []int{depth}
	for {
		lineNum++
		reading, err = reader.ReadString('\n')
		if err != nil {
			err = fmt.Errorf("unable to read line %d: %w", lineNum, err)
			break
		}
		depth, err = strconv.Atoi(strings.TrimSpace(reading))
		if err != nil {
			err = fmt.Errorf("unable to parse line %d: %w", lineNum, err)
			break
		}
		depths = append(depths, depth)
	}
	if !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("unexpected error reading input: %w", err)
	}

	return depths, nil
}
