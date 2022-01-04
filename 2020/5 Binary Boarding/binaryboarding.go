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
	"sort"
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

func boardingPassToRowCol(boardingPass string) (row, col int64) {
	if len(boardingPass) != 10 {
		panic("boarding pass must have len = 10")
	}
	rowCode, colCode := boardingPass[:7], boardingPass[7:]
	rowCode = strings.ReplaceAll(rowCode, "F", "0")
	rowCode = strings.ReplaceAll(rowCode, "B", "1")
	colCode = strings.ReplaceAll(colCode, "L", "0")
	colCode = strings.ReplaceAll(colCode, "R", "1")

	row, err := strconv.ParseInt(rowCode, 2, 32)
	if err != nil { panic(err) }
	col, err = strconv.ParseInt(colCode, 2, 32)
	if err != nil { panic(err) }
	return row, col
}

func seatID(row, col int64) int64 {
	return row*8 + col
}

func main() {
	boardingPasses := readInputLines()
	seatIDs := make([]int64, len(boardingPasses))
	for i, boardingPass := range boardingPasses {
		row, col := boardingPassToRowCol(boardingPass)
		seatIDs[i] = seatID(row, col)
	}
	sort.Slice(seatIDs, func(i, j int) bool { return seatIDs[i] < seatIDs[j] })
	fmt.Printf("Max seat ID is %d\n", seatIDs[len(seatIDs)-1])
	for i := 0; i < len(seatIDs)-1; i++ {
		seatID := seatIDs[i]
		if seatIDs[i+1] != seatID+1 {
			fmt.Printf("Empty seat ID is %d\n", seatID+1)
			break
		}
	}
}
