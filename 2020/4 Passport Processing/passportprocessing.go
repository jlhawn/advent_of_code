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

var RequiredPassportFields = []string{
	"byr", "iyr", "eyr", "hgt", "hcl", "ecl", "pid",
}

var OptionalPassportFields = []string{
	"cid",
}

var ValidEyeColors map[string]bool

var PassportFieldValidators map[string]func(string)bool

func init() {
	ValidEyeColors = map[string]bool {
		"amb": true,
		"blu": true,
		"brn": true,
		"gry": true,
		"grn": true,
		"hzl": true,
		"oth": true,
	}
	PassportFieldValidators = map[string]func(string)bool {
		"byr": func(val string) bool {
			year, err := strconv.Atoi(val)
			if err != nil {
				return false
			}
			return 1920 <= year && year <= 2002
		},
		"iyr": func(val string) bool {
			year, err := strconv.Atoi(val)
			if err != nil {
				return false
			}
			return 2010 <= year && year <= 2020
		},
		"eyr": func(val string) bool {
			year, err := strconv.Atoi(val)
			if err != nil {
				return false
			}
			return 2020 <= year && year <= 2030
		},
		"hgt": func(val string) bool {
			unit := val[len(val)-2:]
			height, err := strconv.Atoi(val[:len(val)-2])
			if err != nil {
				return false
			}
			switch unit {
			case "cm":
				return 150 <= height && height <= 193
			case "in":
				return 59 <= height && height <= 76
			default:
				return false
			}
		},
		"hcl": func(val string) bool {
			if val[0] != '#' || len(val) != 7 {
				return false
			}
			_, err := strconv.ParseInt(val[1:], 16, 64)
			return err == nil
		},
		"ecl": func(val string) bool {
			return ValidEyeColors[val]
		},
		"pid": func(val string) bool {
			if len(val) != 9 {
				return false
			}
			_, err := strconv.Atoi(val)
			return err == nil
		},
	}
}

type Passport map[string]string

func (p Passport) IsValid(validateValue bool) bool {
	for _, field := range RequiredPassportFields {
		val, hasField := p[field]
		if !hasField || (validateValue && !PassportFieldValidators[field](val)) {
			return false
		}
	}
	return true
}

func loadPassports() []Passport {
	var passports []Passport
	passport := Passport{}
	for _, line := range readInputLines() {
		if len(line) == 0 {
			// End of current passport.
			passports = append(passports, passport)
			passport = Passport{} // Start a new Passport.
			continue
		}

		fields := strings.Split(line, " ")
		for _, field := range fields {
			key, val, ok := strings.Cut(field, ":")
			if !ok {
				panic(fmt.Errorf("unable to cut passport field: %q", field))
			}

			passport[key] = val
		}
	}
	return append(passports, passport)
}

func main() {
	passports := loadPassports()

	validCount := 0
	for _, passport := range passports {
		if passport.IsValid(false) {
			validCount++
		}
	}
	fmt.Printf("Valid Passport Count (Ignoring Values): %d\n", validCount)

	validCount = 0
	for _, passport := range passports {
		if passport.IsValid(true) {
			validCount++
		}
	}
	fmt.Printf("Valid Passport Count (Validating Values): %d\n", validCount)
}
