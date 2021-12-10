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

type Direction int

const (
	Forward Direction = iota
	Up
	Down
)

type InvalidDirectionError struct {
	rawValue string
}

func (e *InvalidDirectionError) Error() string {
	return fmt.Sprintf("Invalid Direction: %q", e.rawValue)
}

func parseDirection(raw string) (Direction, error) {
	switch strings.ToLower(raw) {
	case "forward":
		return Forward, nil
	case "up":
		return Up, nil
	case "down":
		return Down, nil
	}
	return Direction(-1), &InvalidDirectionError{rawValue: raw}
}

type Command struct {
	Direction
	Magnitude int
}

type InvalidCommandError struct {
	rawValue string
	reason error
}

func (e *InvalidCommandError) Error() string {
	if e.reason != nil {
		return fmt.Sprintf("Invalid Command %q: %s", e.rawValue, e.reason)
	}
 	return fmt.Sprintf("Invalid Command %q", e.rawValue)
}

func (e *InvalidCommandError) Unwrap() error {
	return e.reason
}

func parseCommand(raw string) (*Command, error) {
	parts := strings.Fields(raw)

	// There should be 2 parts: a direction and a magnitude.
	if len(parts) != 2 {
		return nil, &InvalidCommandError{rawValue: raw}
	}

	direction, err := parseDirection(parts[0])
	if err != nil {
		return nil, &InvalidCommandError{rawValue: raw, reason: err}
	}

	magnitude, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, &InvalidCommandError{rawValue: raw, reason: err}
	}

	if magnitude <= 0 {
		return nil, &InvalidCommandError{rawValue: raw}
	}

	return &Command{
		Direction: direction,
		Magnitude: magnitude,
	}, nil
}

func readInput() ([]*Command, error) {
	inputFile, err := os.Open(inputFilename)
	if err != nil {
		return nil, fmt.Errorf("unable to open input file: %w", err)
	}
	defer inputFile.Close()

	lines, err := readLines(inputFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read input lines: %w", err)
	}

	var commands []*Command
	for i, line := range lines {
		command, err := parseCommand(line)
		if err != nil {
			return nil, fmt.Errorf("unable to parse command on line %d: %w", i+1, err)
		}
		commands = append(commands, command)
	}

	return commands, nil
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

func dive(commands []*Command) {
	var aim, depth, horizontalPos int
	for _, command := range commands {
		switch command.Direction {
		case Down:
			aim += command.Magnitude
		case Up:
			aim -= command.Magnitude
		case Forward:
			horizontalPos += command.Magnitude
			depth += aim * command.Magnitude
		}
	}
	fmt.Printf("Final Depth: %d\nFinal Distance: %d\n", depth, horizontalPos)
	fmt.Printf("Product: %d\n", depth*horizontalPos)
}

func main() {
	commands, err := readInput()
	if err != nil {
		log.Fatal(err)
	}

	dive(commands)
}
