package main

import (
	"bufio"
	// "constraints"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	// "time"

	"../../slices"
	// "../../streams"
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

func loadXMASData() []int {
	lines := readInputLines()
	vals := make([]int, len(lines))
	for i, line := range lines {
		val, err := strconv.Atoi(line)
		if err != nil {
			panic(err)
		}
		vals[i] = val
	}
	return vals
}

func firstInvalidXMASVal(vals []int, preambleSize int) (int, bool) {
	for i := preambleSize; i < len(vals); i++ {
		val := vals[i]
		prevs := vals[i-preambleSize:i]
		if !any2Equal(prevs, val) {
			return val, true
		}
	}
	return 0, false
}

func any2Equal(vals []int, targetSum int) bool {
	for i := 0; i < len(vals); i++ {
		for j := i+1; j < len(vals); j++ {
			if vals[i] + vals[j] == targetSum {
				return true
			}
		}
	}
	return false
}

func contiguousRangeWithSum(vals []int, targetSum int) []int {
	for i := 0; i < len(vals); i++ {
		rangeSum := vals[i]
		for j := i+1; j < len(vals); j++ {
			rangeSum += vals[j]
			if rangeSum == targetSum {
				return vals[i:j+1]
			}
			if rangeSum > targetSum {
				break
			}
		}
	}
	return nil
}

func main() {
	vals := loadXMASData()
	invalidVal, ok := firstInvalidXMASVal(vals, 25)
	fmt.Printf("first invalid XMAS value: %d %t\n", invalidVal, ok)

	rangeWithSum := contiguousRangeWithSum(vals, invalidVal)
	if rangeWithSum == nil {
		panic("no contiguousRangeWithSum")
	}

	min := slices.Min(rangeWithSum...)
	max := slices.Max(rangeWithSum...)

	fmt.Printf("encryption weakness is %d + %d = %d\n", min, max, min+max)
}
