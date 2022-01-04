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

func loadExpenseReport() []int {
	var expenses []int
	for _, line := range readInputLines() {
		expense, err := strconv.Atoi(line)
		if err != nil { panic(err) }
		expenses = append(expenses, expense)
	}
	return expenses
}

func findPairProductForSum(sum int, vals ...int) int {
	for i := 0; i < len(vals); i++ {
		for j := i+1; j < len(vals); j++ {
			a, b := vals[i], vals[j]
			if a + b == sum {
				return a * b
			}
		}
	}
	return -1
}

func findTripleProductForSum(sum int, vals ...int) int {
	for i := 0; i < len(vals); i++ {
		for j := i+1; j < len(vals); j++ {
			for k := j+1; k < len(vals); k++ {
				a, b, c := vals[i], vals[j], vals[k]
				if a + b + c == sum {
					return a * b * c
				}
			}
		}
	}
	return -1
}

func main() {
	expenses := loadExpenseReport()
	fmt.Println("Pair Product: ", findPairProductForSum(2020, expenses...))
	fmt.Println("Triple Product: ", findTripleProductForSum(2020, expenses...))
}
