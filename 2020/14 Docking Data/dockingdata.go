package main

import (
	"bufio"
	// "constraints"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
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

func parseMask(rawMask string) (mask, alwaysOn int64) {
	maxShift := len(rawMask)-1
	for i, char := range []byte(rawMask) {
		switch char {
		case '0':
			mask += 1 << (maxShift-i)
		case '1':
			mask += 1 << (maxShift-i)
			alwaysOn += 1 << (maxShift-i)
		case 'X':
			// Ignore
		}
	}
	return
}

func Part1() {
	mem := map[int64]int64{}
	var mask, alwaysOn int64
	for _, line := range readInputLines() {
		if strings.HasPrefix(line, "mask = ") {
			mask, alwaysOn = parseMask(strings.TrimPrefix(line, "mask = "))
			continue
		}
		line = strings.TrimPrefix(line, "mem[")
		rawAddr, rawVal, ok := strings.Cut(line, "] = ")
		if !ok { panic("unable to cut line") }
		addr, err := strconv.ParseInt(rawAddr, 10, 64)
		if err != nil { panic(err) }
		val, err := strconv.ParseInt(rawVal, 10, 64)
		if err != nil { panic(err) }
		mem[addr] = (val & ^mask) | alwaysOn
	}

	var sum int64
	for _, val := range mem {
		sum += val
	}
	fmt.Printf("Part 1 sum of memory values: %d\n", sum)
}

func parseMaskAndFloatOffsets(rawMask string) (mask int64, floatOffsets []int) {
	maxShift := len(rawMask)-1
	for i, char := range []byte(rawMask) {
		switch char {
		case '0':
			// Ignore.
		case '1':
			mask += 1 << (maxShift-i)
		case 'X':
			floatOffsets = append(floatOffsets, maxShift-i)
		}
	}
	return
}

func floatAddrs(addr, mask int64, floatOffsets []int) []int64 {
	addr |= mask // Apply first mask.
	// Then, zero out each of the floating bits.
	mask = 0
	for _, offset := range floatOffsets {
		mask += 1 << offset
	}
	addr &= ^mask
	fmt.Printf("unfloated addr is %036b (%[1]d)\n", addr)

	floated := make([]int64, 0, 1 << len(floatOffsets))
	floatedBuf := make([]int64, 0, cap(floated))
	floated = append(floated, addr)
	for _, offset := range floatOffsets {
		bit := int64(1 << offset)
		for _, floatedAddr := range floated {
			// Add the address without the bit set and with the bit set.
			// Note: the addrs always start with the bit unset.
			floatedBuf = append(floatedBuf, floatedAddr, floatedAddr+bit)
		}
		floated, floatedBuf = floatedBuf, floated[:0]
	}
	sort.Slice(floated, func(i, j int) bool { return floated[i] < floated[j] })
	return floated
}

func Part2() {
	mem := map[int64]int64{}
	var mask int64
	var floatOffsets []int
	for _, line := range readInputLines() {
		if strings.HasPrefix(line, "mask = ") {
			mask, floatOffsets = parseMaskAndFloatOffsets(strings.TrimPrefix(line, "mask = "))
			continue
		}
		line = strings.TrimPrefix(line, "mem[")
		rawAddr, rawVal, ok := strings.Cut(line, "] = ")
		if !ok { panic("unable to cut line") }
		addr, err := strconv.ParseInt(rawAddr, 10, 64)
		if err != nil { panic(err) }
		val, err := strconv.ParseInt(rawVal, 10, 64)
		if err != nil { panic(err) }
		for _, floatedAddr := range floatAddrs(addr, mask, floatOffsets) {
			fmt.Printf("writing %10d to memory addr %036b (%[2]d)\n", val, floatedAddr)
			mem[floatedAddr] = val
		}
	}

	var sum int64
	for _, val := range mem {
		sum += val
	}
	fmt.Printf("Part 2 sum of memory values: %d\n", sum)
}

func main() {
	Part1()
	Part2()
}
