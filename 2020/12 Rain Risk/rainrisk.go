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

type Instruction struct {
	Direction string
	Amount int
}

func parseInstruction(raw string) Instruction {
	dir := raw[:1]
	amount, err := strconv.Atoi(raw[1:])
	if err != nil { panic(err) }

	return Instruction{Direction: dir, Amount: amount}
}

func loadInstructions() []Instruction {
	lines := readInputLines()
	instructions := make([]Instruction, len(lines))
	for i, line := range lines {
		instructions[i] = parseInstruction(line)
	}
	return instructions
}

type Coord struct {
	X, Y int
}

func (c Coord) ManhattanDistance() int {
	dx := c.X
	if dx < 0 {
		dx = -dx
	}
	dy := c.Y
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

type CoordinateAndHeading struct {
	Coord
	Heading int
}

func (c *CoordinateAndHeading) Handle(inst Instruction) {
	switch inst.Direction {
	case "N":
		c.Y += inst.Amount
	case "S":
		c.Y -= inst.Amount
	case "E":
		c.X += inst.Amount
	case "W":
		c.X -= inst.Amount
	case "L":
		c.Heading = (c.Heading + inst.Amount) % 360
		if c.Heading < 0 {
			c.Heading += 360
		}
	case "R":
		c.Heading = (c.Heading - inst.Amount) % 360
		if c.Heading < 0 {
			c.Heading += 360
		}
	case "F":
		switch c.Heading {
		case 0: // Heading East.
			c.X += inst.Amount
		case 90: // Heading North.
			c.Y += inst.Amount
		case 180: // Heading West.
			c.X -= inst.Amount
		case 270: // Heading South.
			c.Y -= inst.Amount
		default:
			panic(fmt.Errorf("Heading is not a multiple of 90 degrees: %d", c.Heading))
		}
	default:
		panic(fmt.Errorf("unknown instruction: %#v", inst))
	}
}

type CoordinateAndWaypoint struct {
	Coord
	WayPoint Coord
}

func (c *CoordinateAndWaypoint) Handle(inst Instruction) {
	switch inst.Direction {
	case "N":
		c.WayPoint.Y += inst.Amount
	case "S":
		c.WayPoint.Y -= inst.Amount
	case "E":
		c.WayPoint.X += inst.Amount
	case "W":
		c.WayPoint.X -= inst.Amount
	case "L":
		// Convert to a negative clockwise rotation.
		inst.Amount = -inst.Amount
		fallthrough
	case "R":
		// Normalize to positive rotation.
		inst.Amount %= 360
		if inst.Amount < 0 {
			inst.Amount += 360
		}
		if inst.Amount % 90 != 0 {
			panic(fmt.Errorf("Rotation is not a multiple of 90 degrees: %d", inst.Amount))
		}
		for inst.Amount > 0 {
			// Rotate waypoint 90 degrees clockwise.
			c.WayPoint.X, c.WayPoint.Y = c.WayPoint.Y, -c.WayPoint.X
			inst.Amount -= 90
		}
	case "F":
		for inst.Amount > 0 {
			c.X += c.WayPoint.X
			c.Y += c.WayPoint.Y
			inst.Amount--
		}
	default:
		panic(fmt.Errorf("unknown instruction: %#v", inst))
	}
}

func main() {
	instructions := loadInstructions()

	ship := CoordinateAndHeading{}
	for _, inst := range instructions {
		ship.Handle(inst)
	}
	fmt.Printf("Ship %#v\n", ship)
	fmt.Printf("ManhattanDistance: %d\n", ship.ManhattanDistance())

	shipAndWaypoint := CoordinateAndWaypoint{WayPoint: Coord{X: 10, Y: 1}}
	for _, inst := range instructions {
		shipAndWaypoint.Handle(inst)
	}
	fmt.Printf("Ship and WayPoint %#v\n", shipAndWaypoint)
	fmt.Printf("ManhattanDistance: %d\n", shipAndWaypoint.ManhattanDistance())
}
