package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	// "sort"
	"strconv"
	"strings"
)

const inputFilename = "./INPUT"

func readInputLines() []string {
	inputFile, err := os.Open(inputFilename)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to open input file: %w", err))
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
		log.Fatal(fmt.Errorf("unexpected error reading input: %w", err))
	}

	return lines
}

const (
	PacketTypeSum int64 = iota
	PacketTypeProduct
	PacketTypeMinimum
	PacketTypeMaximum
	PacketTypeLiteral
	PacketTypeGreaterThan
	PacketTypeLessThan
	PacketTypeEqualTo
)

var opNames map[int64]string
var hexMap map[byte]string

func init() {
	opNames = map[int64]string{
		PacketTypeSum: "sum",
		PacketTypeProduct: "product",
		PacketTypeMinimum: "min",
		PacketTypeMaximum: "max",
		PacketTypeLiteral: "literal",
		PacketTypeGreaterThan: "greaterThan",
		PacketTypeLessThan: "lessThan",
		PacketTypeEqualTo: "equalTo",
	}
	hexMap = map[byte]string{
		'0': "0000",
		'1': "0001",
		'2': "0010",
		'3': "0011",
		'4': "0100",
		'5': "0101",
		'6': "0110",
		'7': "0111",
		'8': "1000",
		'9': "1001",
		'A': "1010",
		'B': "1011",
		'C': "1100",
		'D': "1101",
		'E': "1110",
		'F': "1111",
	}
}

// output bytes are '0' or '1'
func loadBitSequence() *bytes.Buffer {
	rawLines := readInputLines()
	if len(rawLines) != 1 {
		log.Fatalf("expected exactly 1 line but got %d", len(rawLines))
	}

	rawLine := rawLines[0]

	var bitSequence []byte
	for _, hexVal := range []byte(rawLine) {
		bitSequence = append(bitSequence, []byte(hexMap[hexVal])...)
	}

	return bytes.NewBuffer(bitSequence)
}

type Packet struct {
	Version int64
	Type int64

	LiteralValue int64

	SubPackets []*Packet
}

func (p *Packet) VersionSum() int64 {
	sum := p.Version
	for _, subPacket := range p.SubPackets {
		sum += subPacket.VersionSum()
	}
	return sum
}

// I could use generics here but it's not strictly necessary and I don't want to
// have code that folks not using Go 1.18+ can't run.
func Reduce(reducer func(int64, int64) int64, vals []int64, initializer int64) int64 {
	reduced := initializer
	for _, val := range vals {
		reduced = reducer(reduced, val)
	}
	return reduced
}

func BoolToBit(b bool) int64 { if b { return 1 }; return 0 }

func (p *Packet) Eval() int64 {
	subVals := make([]int64, len(p.SubPackets))
	for i, subPacket := range p.SubPackets {
		subVals[i] = subPacket.Eval()
	}

	switch p.Type {
	case PacketTypeSum:
		return Reduce(func(a, b int64) int64 { return a + b }, subVals, 0)
	case PacketTypeProduct:
		return Reduce(func(a, b int64) int64 { return a * b }, subVals, 1)
	case PacketTypeMinimum:
		return Reduce(func(a, b int64) int64 { if a < b { return a }; return b }, subVals[1:], subVals[0])
	case PacketTypeMaximum:
		return Reduce(func(a, b int64) int64 { if a > b { return a }; return b }, subVals[1:], subVals[0])
	case PacketTypeLiteral:
		fallthrough
	default:
		return p.LiteralValue
	case PacketTypeGreaterThan:
		return BoolToBit(subVals[0] > subVals[1])
	case PacketTypeLessThan:
		return BoolToBit(subVals[0] < subVals[1])
	case PacketTypeEqualTo:
		return BoolToBit(subVals[0] == subVals[1])
	}
}

func (p *Packet) String() string {
	return p.IndentString("    ")
}

func (p *Packet) IndentString(indent string) string {
	if p.SubPackets == nil {
		return fmt.Sprintf("(v%d %s %d)", p.Version, opNames[p.Type], p.LiteralValue)
	}

	var subStr string
	for i, subPacket := range p.SubPackets {
		if i > 0 {
			subStr += ",\n"+indent
		}
		subStr += subPacket.IndentString(indent + "    ")
	}

	return fmt.Sprintf("(v%d %s\n%s%s)", p.Version, opNames[p.Type], indent, subStr)
}

func parsePacket(bitSequence *bytes.Buffer) *Packet {
	versionBits := bitSequence.Next(3)
	typeBits := bitSequence.Next(3)

	version, err := strconv.ParseInt(string(versionBits), 2, 64)
	if err != nil { panic(err) }

	packetType, err := strconv.ParseInt(string(typeBits), 2, 64)
	if err != nil { panic(err) }

	if packetType == PacketTypeLiteral {
		return &Packet{
			Version: version,
			Type: packetType,
			LiteralValue: parseLiteralValue(bitSequence),
		}
	}

	// This packet is an operation type.
	var subPackets []*Packet
	lengthTypeBit := bitSequence.Next(1)[0]
	if lengthTypeBit == '0' {
		lengthBits := bitSequence.Next(15)
		length, err := strconv.ParseInt(string(lengthBits), 2, 64)
		if err != nil { panic(err) }

		subPacketsBitSequence := bytes.NewBuffer(bitSequence.Next(int(length)))
		for subPacketsBitSequence.Len() > 0 {
			subPackets = append(subPackets, parsePacket(subPacketsBitSequence))
		}
	} else {
		numSubPacketsBit := bitSequence.Next(11)
		numSubPackets, err := strconv.ParseInt(string(numSubPacketsBit), 2, 64)
		if err != nil { panic(err) }

		subPackets = make([]*Packet, int(numSubPackets))
		for i := 0; i < int(numSubPackets); i++ {
			subPackets[i] = parsePacket(bitSequence)
		}
	}

	return &Packet{
		Version: version,
		Type: packetType,
		SubPackets: subPackets,
	}
}

func parseLiteralValue(bitSequence *bytes.Buffer) int64 {
	var valueBits []byte
	for {
		prefixBit := bitSequence.Next(1)[0]
		valueBits = append(valueBits, bitSequence.Next(4)...)

		if prefixBit == '0' {
			break
		}
	}

	value, err := strconv.ParseInt(string(valueBits), 2, 64)
	if err != nil { panic(err) }

	return value
}

func main() {
	bitSequence := loadBitSequence()
	fmt.Printf("bit sequence: %s\n", bitSequence)

	packet := parsePacket(bitSequence)
	fmt.Printf("%s\n", packet)
	fmt.Printf("version sum: %d\n", packet.VersionSum())
	fmt.Printf("evaluated expression: %d\n", packet.Eval())
}
