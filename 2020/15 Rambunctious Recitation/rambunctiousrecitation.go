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

func loadStartingNumbers() (startingNumbers []int) {
	lines := readInputLines()
	rawNums := strings.Split(lines[0], ",")
	startingNumbers = make([]int, len(rawNums))
	for i, rawNum := range rawNums {
		num, err := strconv.Atoi(rawNum)
		if err != nil { panic(err) }
		startingNumbers[i] = num
	}
	return startingNumbers
}

type IndexRecord struct {
	Index, Prev int
}

func OrdinalSuffix(n int) string {
	switch n % 10 {
	case 1:
		return "st"
	case 2:
		return "nd"
	case 3:
		return "rd"
	default:
		return "th"
	}
}

func DetermineNthNumber(N int) {
	startingNumbers := loadStartingNumbers()

	mem := map[int]IndexRecord{}
	var i, lastNum int
	for i, lastNum = range startingNumbers {
		if record, ok := mem[lastNum]; ok {
			mem[lastNum] = IndexRecord{
				Index: i,
				Prev: record.Index,
			}
		} else {
			mem[lastNum] = IndexRecord{
				Index: i,
				Prev: i,
			}
		}
	}
	for i := len(startingNumbers); i < N; i++ {
		record := mem[lastNum]
		lastNum = record.Index - record.Prev
		if record, ok := mem[lastNum]; ok {
			mem[lastNum] = IndexRecord{
				Index: i,
				Prev: record.Index,
			}
		} else {
			mem[lastNum] = IndexRecord{
				Index: i,
				Prev: i,
			}
		}
	}
	fmt.Printf("%d%s number spoken was %d\n", N, OrdinalSuffix(N), lastNum)
}

func main() {
	DetermineNthNumber(2020)
	DetermineNthNumber(30000000)
}
