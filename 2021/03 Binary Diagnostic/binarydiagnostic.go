package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const inputFilename = "./INPUT"

type Bit byte

const (
	ZeroBit Bit = '0'
	OneBit Bit = '1'
)

type BitArray []Bit

func (a BitArray) AsInt() int {
	bitWidth := len(a)
	var val int
	for bitPos, bit := range a {
		if bit == OneBit {
			val += 1 << (bitWidth - bitPos - 1)
		}
	}
	return val
}

func readInput() ([]BitArray, error) {
	inputFile, err := os.Open(inputFilename)
	if err != nil {
		return nil, fmt.Errorf("unable to open input file: %w", err)
	}
	defer inputFile.Close()

	lines, err := readLines(inputFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read input lines: %w", err)
	}

	bitArrays := make([]BitArray, len(lines))
	for i, line := range lines {
		bitArr := BitArray(line)
		for _, bit := range bitArr {
			switch bit {
			case ZeroBit, OneBit:
			default:
				return nil, fmt.Errorf("Invalid binary value on line %d: %q", i+1, line)
			}
		}
		bitArrays[i] = bitArr
	}

	return bitArrays, nil
}

func readLines(reader io.Reader) ([]string, error) {
	var (
		line string
		lineNum int
		lines []string
		err error
	)

	bufReader := bufio.NewReader(reader)
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

func getBitPosCounts(bitArrays []BitArray) map[int]int {
	bitPosCounts := make(map[int]int)
	for _, bitArray := range bitArrays {
		for i := 0; i < len(bitArray); i++ {
			bit := bitArray[i]
			switch bit {
			case ZeroBit:
				bitPosCounts[i] += 0
			case OneBit:
				bitPosCounts[i] += 1
			default:
				log.Fatalf("Unexpected bit value: %s", bit)
			}
		}
	}
	return bitPosCounts
}

func determineGammaEpsilonRates(bitArrays []BitArray) (gammaRate, epsilonRate int) {
	bitPosCounts := getBitPosCounts(bitArrays)
	bitWidth := len(bitPosCounts)

	bitPos := 0
	count, ok := bitPosCounts[bitPos]
	for ok {
		if count >= len(bitArrays)/2 {
			gammaRate += 1 << (bitWidth - bitPos - 1)
		} else {
			epsilonRate += 1 << (bitWidth - bitPos - 1)
		}

		bitPos++
		count, ok = bitPosCounts[bitPos]
	}
	return gammaRate, epsilonRate
}

func determineOxygenGeneratorRating(bitArrays []BitArray) int {
	filterBitPos := 0

	for len(bitArrays) > 0 && filterBitPos < len(bitArrays[0]) {
		bitPosCounts := getBitPosCounts(bitArrays)
		var filter func(BitArray) bool
		if bitPosCounts[filterBitPos] >= len(bitArrays)/2 {
			// 1 is most common in this position.
			filter = func(bitArray BitArray) bool { return bitArray[filterBitPos] == OneBit }
		} else {
			filter = func(bitArray BitArray) bool { return bitArray[filterBitPos] == ZeroBit }
		}
		bitArrays = filterBitArrays(bitArrays, filter)

		if len(bitArrays) == 1 {
			return bitArrays[0].AsInt()
		}

		filterBitPos++
	}

	log.Fatal("unable to filter bit arrays to determine oxygen generator rating")
	return -1
}

func determineCO2ScrubberRating(bitArrays []BitArray) int {
	filterBitPos := 0

	for len(bitArrays) > 0 && filterBitPos < len(bitArrays[0]) {
		bitPosCounts := getBitPosCounts(bitArrays)
		var filter func(BitArray) bool
		if bitPosCounts[filterBitPos] >= len(bitArrays)/2 {
			// 1 is most common in this position.
			filter = func(bitArray BitArray) bool { return bitArray[filterBitPos] == ZeroBit }
		} else {
			filter = func(bitArray BitArray) bool { return bitArray[filterBitPos] == OneBit }
		}
		bitArrays = filterBitArrays(bitArrays, filter)

		if len(bitArrays) == 1 {
			return bitArrays[0].AsInt()
		}

		filterBitPos++
	}
	
	log.Fatal("unable to filter bit arrays to determine CO2 scrubber rating")
	return -1
}

func filterBitArrays(bitArrays []BitArray, filterFunc func(BitArray) bool) []BitArray {
	filtered := make([]BitArray, 0, len(bitArrays))
	for _, bitArray := range bitArrays {
		if filterFunc(bitArray) {
			filtered = append(filtered, bitArray)
		}
	}
	return filtered
}

func main() {
	bitArrays, err := readInput()
	if err != nil {
		log.Fatal(err)
	}
	
	gammaRate, epsilonRate := determineGammaEpsilonRates(bitArrays)

	fmt.Printf("Gamma Rate: %d\n", gammaRate)
	fmt.Printf("Epsilon Rate: %d\n", epsilonRate)
	fmt.Printf("Gamma * Epsilon = %d\n", gammaRate * epsilonRate)

	oxygenGeneratorRating := determineOxygenGeneratorRating(bitArrays)
	co2ScrubberRating := determineCO2ScrubberRating(bitArrays)

	fmt.Printf("Oxygen Generator Rating: %d\n", oxygenGeneratorRating)
	fmt.Printf("CO2 Scrubber Rating: %d\n", co2ScrubberRating)
	fmt.Printf("Life Support Rating: %d\n", oxygenGeneratorRating * co2ScrubberRating)
}
