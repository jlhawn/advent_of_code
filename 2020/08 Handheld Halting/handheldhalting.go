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

	"../../slices"
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
	Op string
	Arg int

	ExecCount int
}

func (inst *Instruction) Exec(pc, acc *int) {
	switch inst.Op {
	case "jmp":
		*pc += inst.Arg
	case "acc":
		*acc += inst.Arg
		fallthrough
	default:
		*pc++
		// NOP
	}
	inst.ExecCount++
}

func (inst *Instruction) Reset() {
	inst.ExecCount = 0
}

func (inst *Instruction) String() string {
	return fmt.Sprintf("%s %+d", inst.Op, inst.Arg)
}

func parseInstruction(line string) *Instruction {
	op, rawArg, ok := strings.Cut(line, " ")
	if !ok {
		panic(fmt.Errorf("unable to cut raw instruction: %q", line))
	}

	arg, err := strconv.Atoi(rawArg)
	if err != nil { panic(err) }

	return &Instruction{Op: op, Arg: arg}
}

func loadInstructions() []*Instruction {
	return slices.Map(func(line string) *Instruction { return parseInstruction(line) }, readInputLines()...)
}

func halts(instructions []*Instruction) (acc int, halted bool) {
	slices.ForEach(func(inst *Instruction) { inst.Reset() }, instructions...)

	pc, acc := 0, 0
	for 0 <= pc && pc < len(instructions) {
		inst := instructions[pc]
		if inst.ExecCount > 0 {
			return acc, false
		}
		inst.Exec(&pc, &acc)
	}
	return acc, pc == len(instructions)
}

func main() {
	instructions := loadInstructions()
	acc, _ := halts(instructions)
	fmt.Printf("Acc = %d\n", acc)

	for _, inst := range instructions {
		originalOp := inst.Op
		switch inst.Op {
		case "jmp":
			inst.Op = "nop"
		case "nop":
			inst.Op = "jmp"
		default:
			continue
		}

		if acc, ok := halts(instructions); ok {
			fmt.Printf("Halted Acc = %d\n", acc)
			return
		}
		inst.Op = originalOp
	}
	fmt.Printf("unable to find boot code error\n")
}
