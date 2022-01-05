package main

import (
	"bufio"
	// "constraints"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	// "time"

	// "../../slices"
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

func loadAdapters() []int {
	lines := readInputLines()
	adapters := make([]int, len(lines))
	for i, line := range lines {
		adapter, err := strconv.Atoi(line)
		if err != nil { panic(err) }
		adapters[i] = adapter
	}
	return adapters
}

type AdapterSet map[int]bool

func WaysToArrange(adapters AdapterSet, inJoltage, outJoltage int, cache map[int]int) int {
	if cachedWays, ok := cache[inJoltage]; ok {
		return cachedWays
	}

	if inJoltage == outJoltage {
		cache[inJoltage] = 1
		return 1
	}

	ways := 0
	if adapters[inJoltage+1] {
		ways += WaysToArrange(adapters, inJoltage+1, outJoltage, cache)
	}
	if adapters[inJoltage+2] {
		ways += WaysToArrange(adapters, inJoltage+2, outJoltage, cache)
	}
	if adapters[inJoltage+3] {
		ways += WaysToArrange(adapters, inJoltage+3, outJoltage, cache)
	}
	cache[inJoltage] = ways
	return ways
}

func main() {
	adapters := loadAdapters()
	sort.Ints(adapters)

	var d1Jolts, d3Jolts, prevJolts int
	for _, adapter := range adapters {
		switch adapter - prevJolts {
		case 1:
			d1Jolts++
		case 3:
			d3Jolts++
		}
		prevJolts = adapter
	}
	d3Jolts++ // Our device allows a 3-Jolt difference.
	fmt.Printf("number of 1-jolt differences multiplied by 3-jold differences: %d * %d = %d\n", d1Jolts, d3Jolts, d1Jolts*d3Jolts)

	maxAdapter := prevJolts
	adapterSet := make(AdapterSet, len(adapters))
	for _, adapter := range adapters {
		adapterSet[adapter] = true
	}

	ways := WaysToArrange(adapterSet, 0, maxAdapter, map[int]int{})
	fmt.Printf("number of ways to arrange adapters to connect to device: %d\n", ways)
}
