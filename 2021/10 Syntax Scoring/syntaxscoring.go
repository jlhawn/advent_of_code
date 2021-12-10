package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	// "strconv"
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

var IncompleteLineErr = errors.New("Incomplete")

type CorruptedLineErr struct {
	Expected, Found byte
}

func (e *CorruptedLineErr) Error() string {
	return fmt.Sprintf("Expected %q, but found %q instead.", e.Expected, e.Found)
}

var illegalCharPoints = map[byte]int {
	')': 3,
	']': 57,
	'}': 1197,
	'>': 25137,
}

func (e *CorruptedLineErr) IllegalCharPoints() int {
	return illegalCharPoints[e.Found]
}

type Parser struct {
	Line string
	CompletionString string
	Index int
	Autocomplete bool
}

func (p *Parser) IsEnd() bool {
	return p.Index >= len(p.Line)
}

func (p *Parser) Peek() byte {
	if p.IsEnd() {
		return 0
	}

	return p.Line[p.Index]
}

func (p *Parser) Next() byte {
	if p.IsEnd() {
		return 0
	}

	c := p.Line[p.Index]
	p.Index++
	return c
}

func (p *Parser) CompleteWithChar(c byte) {
	p.Line += string(c)
	p.CompletionString += string(c)
	p.Index++
}

var completedCharPoints = map[byte]int {
	')': 1,
	']': 2,
	'}': 3,
	'>': 4,
}

func (p *Parser) AutocompleteScore() int {
	totalScore := 0
	for _, c := range []byte(p.CompletionString) {
		totalScore *= 5
		totalScore += completedCharPoints[c]
	}
	return totalScore
}

var (
	openClosePairs map[byte]byte
	closeBrackets map[byte]struct{}
)

func init() {
	openClosePairs = map[byte]byte{
		'(': ')',
		'[': ']',
		'{': '}',
		'<': '>',
	}
}

func (p *Parser) ParseChunks() error {
	open := p.Next()
	if open == 0 {
		return nil
	}
	// fmt.Printf("parser got %q\n", open)
	expected := openClosePairs[open]

	for {
		next := p.Peek()
		if next == 0 {
			if p.Autocomplete {
				p.CompleteWithChar(expected)
				return nil
			}
			return IncompleteLineErr
		}

		if _, isOpen := openClosePairs[next]; !isOpen {
			break
		}

		if err := p.ParseChunks(); err != nil {
			return err
		}
	}

	found := p.Next()
	// fmt.Printf("parser got %q\n", found)
	if expected != found {
		return &CorruptedLineErr{
			Expected: expected,
			Found: found,
		}
	}

	return nil
}

func main() {
	lines, err := readInputLines()
	if err != nil {
		log.Fatal(err)
	}

	var incompleteLines []string
	var corruptedLines []*CorruptedLineErr
	for i, line := range lines {
		parser := &Parser{Line: line}
		err := parser.ParseChunks()
		if errors.Is(err, IncompleteLineErr) {
			fmt.Printf("line %d is Incomplete.\n", i)
			incompleteLines = append(incompleteLines, line)
			continue
		}
		corruptedLine := err.(*CorruptedLineErr)
		fmt.Printf("line %d is corrputed: %s\n", i, corruptedLine)
		corruptedLines = append(corruptedLines, corruptedLine)
	}
	fmt.Printf("There are %d corrputed lines\n", len(corruptedLines))

	syntaxErrScore := 0
	for _, corrputedLine := range corruptedLines {
		syntaxErrScore += corrputedLine.IllegalCharPoints()
	}
	fmt.Printf("Total syntax error score: %d\n", syntaxErrScore)

	fmt.Printf("There are %d incomplete lines\n", len(incompleteLines))
	var parsers []*Parser
	for i, line := range incompleteLines {
		parser := &Parser{Line: line, Autocomplete: true}
		parsers = append(parsers, parser)
		err := parser.ParseChunks()
		if err != nil {
			fmt.Printf("error completing string on line %d: %#v\n", i, parser)
			log.Fatal(err)
		}
	}

	var autocompleteScores []int
	for _, parser := range parsers {
		autocompleteScores = append(autocompleteScores, parser.AutocompleteScore())
	}
	sort.Ints(autocompleteScores)
	median := autocompleteScores[len(autocompleteScores)/2]
	fmt.Printf("Median autocomplete score is %d\n", median)
}
