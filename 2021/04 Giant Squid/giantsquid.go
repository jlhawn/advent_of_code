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

type BingoGame struct {
	NumbersToCall []int

	Boards []BingoBoard
}

func (b *BingoGame) Play() {
	boardsInPlay := b.Boards
	for _, numberCalled := range b.NumbersToCall {
		var nonWinningBoards []BingoBoard
		for _, board := range boardsInPlay {
			board.MarkNum(numberCalled)
			if board.IsWin() {
				fmt.Printf("Winning board score: %d\n", board.Score(numberCalled))
			} else {
				nonWinningBoards = append(nonWinningBoards, board)
			}
		}
		boardsInPlay = nonWinningBoards
	}
}

type BingoBoard [][]BingoNumber

func (b BingoBoard) IsWin() bool {
	// Check for a row which is all marked.
	for i := 0; i < 5; i++ {
		if b.isRowWin(i) || b.isColWin(i) {
			return true
		}
	}
	return false
}

func (b BingoBoard) isRowWin(i int) bool {
	for _, num := range b[i] {
		if !num.Marked {
			return false
		}
	}
	return true
}

func (b BingoBoard) isColWin(j int) bool {
	for i := 0; i < 5; i++ {
		if !b[i][j].Marked {
			return false
		}
	}
	return true
}

func (b BingoBoard) MarkNum(val int) bool {
	for _, row := range b {
		for i := range row {
			if row[i].Number == val {
				row[i].Marked = true
				return true
			}
		}
	}
	return false
}

func (b BingoBoard) Score(lastNum int) int {
	unmarkedSum := 0
	for _, row := range b {
		for _, num := range row {
			if !num.Marked {
				unmarkedSum += num.Number
			}
		}
	}
	return unmarkedSum * lastNum
}

type BingoNumber struct {
	Number int
	Marked bool
}

func readInput() (*BingoGame, error) {
	inputFile, err := os.Open(inputFilename)
	if err != nil {
		return nil, fmt.Errorf("unable to open input file: %w", err)
	}
	defer inputFile.Close()

	lines, err := readLines(inputFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read input lines: %w", err)
	}

	// The first line should be the numbers to call.
	numbersToCallLine := lines[0]
	lines = lines[1:]
	numbersToCall, err := readNumbersToCall(numbersToCallLine)
	if err != nil {
		return nil, fmt.Errorf("unable to read input line of numbers to call: %w", err)
	}

	// This should be followed by a series of bingo boards.
	bingoBoards, err := readBingoBoards(lines)
	if err != nil {
		return nil, fmt.Errorf("unable to read bingo boards: %w", err)
	}

	return &BingoGame{
		NumbersToCall: numbersToCall,
		Boards: bingoBoards,
	}, nil
}

func readNumbersToCall(rawLine string) ([]int, error) {
	rawNums := strings.Split(rawLine, ",")

	numbers := make([]int, len(rawNums))
	for i := 0; i < len(numbers); i++ {
		number, err := strconv.Atoi(rawNums[i])
		if err != nil {
			return nil, fmt.Errorf("unable to parse number %q: %w", rawNums[i], err)
		}

		numbers[i] = number
	}

	return numbers, nil
}

func readBingoBoards(rawLines []string) ([]BingoBoard, error) {
	// These should be a blank line followed by five lines of five
	// numbers each, separated by spaces.
	if len(rawLines) % 6 != 0 {
		return nil, fmt.Errorf("number of lines should be divisible by 6")
	}

	var boards []BingoBoard
	for len(rawLines) > 0 {
		board, err := readBingoBoard(rawLines[:6])
		if err != nil {
			return nil, fmt.Errorf("unable to read board: %w", err)
		}

		boards = append(boards, board)
		rawLines = rawLines[6:]
	}

	return boards, nil
}

func readBingoBoard(rawLines []string) (BingoBoard, error) {
	if len(rawLines) != 6 {
		return nil, fmt.Errorf("board input should be 6 lines")
	}

	if rawLines[0] != "" {
		return nil, fmt.Errorf("board should be preceded by a blank line")		
	}
	rawLines = rawLines[1:]

	board := make([][]BingoNumber, 5)
	for i := 0; i < 5; i++ {
		rawLine := rawLines[i]
		replacer := strings.NewReplacer("  ", " ")
		rawLine = replacer.Replace(rawLine)

		board[i] = make([]BingoNumber, 5)
		rawNums := strings.Split(rawLine, " ")
		if len(rawNums) != 5 {
			return nil, fmt.Errorf("board row should have 5 numbers but has %d: %q", len(rawNums), rawLine)
		}
		for j, rawNum := range rawNums {
			num, err := strconv.Atoi(rawNum)
			if err != nil {
				return nil, fmt.Errorf("unable to parse number %q: %w", rawNum, err)
			}
			board[i][j] = BingoNumber{Number: num}
		}
	}

	return board, nil
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

func main() {
	bingoGame, err := readInput()
	if err != nil {
		log.Fatal(err)
	}

	bingoGame.Play()
}
