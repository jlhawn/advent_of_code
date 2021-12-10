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

func main() {

}
