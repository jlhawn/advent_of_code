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

type PasswordPolicy struct {
	Min, Max int
	Char byte
}

func (p PasswordPolicy) IsAcceptableOld(password string) bool {
	var charCount int
	for _, b := range []byte(password) {
		if b == p.Char {
			charCount++
		}
	}

	return p.Min <= charCount && charCount <= p.Max
}

func (p PasswordPolicy) IsAcceptableNew(password string) bool {
	i := p.Min - 1
	j := p.Max - 1
	var iEqual, jEqual bool
	if i < len(password) {
		iEqual = password[i] == p.Char
	}
	if j < len(password) {
		jEqual = password[j] == p.Char
	}
	return iEqual != jEqual // XOR: only one must be true.
}

func parsePasswordPolicy(raw string) PasswordPolicy{
	minMaxRange, char, ok := strings.Cut(raw, " ")
	if !ok {
		panic(fmt.Errorf("unable to cut raw password policy: %q", raw))
	}

	rawMin, rawMax, ok := strings.Cut(minMaxRange, "-")
	if !ok {
		panic(fmt.Errorf("unable to cut raw minMaxRange: %q", minMaxRange))
	}

	min, err := strconv.Atoi(rawMin)
	if err != nil { panic(err) }
	max, err := strconv.Atoi(rawMax)
	if err != nil { panic(err) }

	return PasswordPolicy{min, max, char[0]}
}

type PasswordTest struct {
	Password string
	Policy PasswordPolicy
}

func loadPasswords() []*PasswordTest {
	var tests []*PasswordTest
	for _, line := range readInputLines() {
		rawPolicy, password, ok := strings.Cut(line, ": ")
		if !ok {
			panic(fmt.Errorf("unable to cut raw input line: %q", line))
		}

		tests = append(tests, &PasswordTest{
			Password: password,
			Policy: parsePasswordPolicy(rawPolicy),
		})
	}

	return tests
}

func main() {
	passwordTests := loadPasswords()

	var validPasswords int
	for _, test := range passwordTests {
		if test.Policy.IsAcceptableOld(test.Password) {
			validPasswords++
		}
	}
	fmt.Println("Count of valid passwords under old method: ", validPasswords)

	validPasswords = 0
	for _, test := range passwordTests {
		if test.Policy.IsAcceptableNew(test.Password) {
			validPasswords++
		}
	}
	fmt.Println("Count of valid passwords under new method: ", validPasswords)
}
