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

	"../../streams"
	"../../slices"
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

type Range struct {
	Min, Max int // Both *Inclusive*.
}

func (r Range) Contains(x int) bool {
	return r.Min <= x && x <= r.Max
}

type Rule struct {
	Name string
	Ranges []Range
}

func (r Rule) IsValid(x int) bool {
	return streams.Any(streams.FromItems(r.Ranges...), func(r Range) bool { return r.Contains(x) })
}

func parseRules(rawRules []string) []Rule {
	rules := make([]Rule, len(rawRules))
	for i, rawRule := range rawRules {
		name, unsplitRanges, ok := strings.Cut(rawRule, ": ")
		if !ok { panic(fmt.Errorf("unable to cut raw rule %q", rawRule)) }

		rawRanges := strings.Split(unsplitRanges, " or ")
		ranges := make([]Range, len(rawRanges))
		for i, rawRange := range rawRanges {
			rawMin, rawMax, ok := strings.Cut(rawRange, "-")
			if !ok { panic(fmt.Errorf("unable to cut raw range %q", rawRange)) }

			min, err := strconv.Atoi(rawMin)
			if err != nil { panic(err) }
			max, err := strconv.Atoi(rawMax)
			if err != nil { panic(err) }

			ranges[i] = Range{Min: min, Max: max}
		}

		rules[i] = Rule{Name: name, Ranges: ranges}
	}

	return rules
}

func loadRulesAndTickets() (rules []Rule, yourTicket []int, allTickets [][]int) {
	lines := readInputLines()

	var rawRules, rawTickets []string
	var i int
	var line string
	for i, line = range lines {
		if len(line) == 0 {
			break
		}
		rawRules = append(rawRules, line)
	}

	// Next line is "your ticket:" label followed by the actual values.
	rawTickets = []string{lines[i+2]}
	// Next line is blank, followed by "nearby tickets:" label followed by the actual values.
	lines = lines[i+5:]
	for _, line := range lines {
		rawTickets = append(rawTickets, line)
	}

	tickets := make([][]int, len(rawTickets))
	for i, rawTicket := range rawTickets {
		rawVals := strings.Split(rawTicket, ",")
		ticket := make([]int, len(rawVals))
		for i, rawVal := range rawVals {
			val, err := strconv.Atoi(rawVal)
			if err != nil { panic(err) }
			ticket[i] = val
		}
		tickets[i] = ticket
	}

	return parseRules(rawRules), tickets[0], tickets
}

func FirstKey[K comparable, V any](m map[K]V) (first K) {
	for first = range m {
		return first
	}
	return first
}

func main() {
	rules, yourTicket, allTickets := loadRulesAndTickets()
	errorRate := 0
	var validTickets [][]int
	for _, ticket := range allTickets {
		ticketIsValid := true
		for _, val := range ticket {
			if !streams.Any(streams.FromItems(rules...), func(r Rule) bool { return r.IsValid(val) }) {
				errorRate += val
				ticketIsValid = false
			}
		}
		if ticketIsValid {
			validTickets = append(validTickets, ticket)
		}
	}
	fmt.Printf("ticket scanning error rate: %d\n", errorRate)

	// For each field index, build a set of rules which match that index of all tickets.
	// And for each rule, build a set of field indexs which match that rule.
	fieldRuleCandidates := make(map[int]map[string]bool, len(rules))
	rulesByName := make(map[string]Rule, len(rules))
	ruleFieldCandidates := make(map[string]map[int]bool, len(rules))
	for _, rule := range rules {
		rulesByName[rule.Name] = rule
		ruleFieldCandidates[rule.Name] = map[int]bool{}
	}

	for i := 0; i < len(rules); i++ {
		ticketVals := slices.Map(func(ticket []int) int { return ticket[i] }, validTickets...)

		fieldRuleCandidates[i] = map[string]bool{}
		for _, rule := range rules {
			if streams.All(streams.FromItems(ticketVals...), func(val int) bool { return rule.IsValid(val) }) {
				fieldRuleCandidates[i][rule.Name] = true
				ruleFieldCandidates[rule.Name][i] = true
			}
		}
	}

	eliminateCandidates := func(ruleName string, fieldIndex int) {
		// Need to delete this rule name from the set of rules not yet matched.
		delete(ruleFieldCandidates, ruleName)
		// Need to delete this field index from the set of fields with rule candidates.
		delete(fieldRuleCandidates, fieldIndex)
		// Need to delete this rule name from the set of candidate rules for
		// fields not yet matched.
		for _, candidateRules := range fieldRuleCandidates {
			delete(candidateRules, ruleName)
		}
		// Need to delete this field index from the set of candidate fields for
		// rules not yet matched.
		for _, validFields := range ruleFieldCandidates {
			delete(validFields, fieldIndex)
		}
	}

	// We can determine which rule corresponds to each field in one of two ways:
	// - A field is only valid for one rule.
	// - A rule is only valid for one field.
	knownRules := make(map[int]Rule, len(rules)) // Maps field position to the rule for that position.
	for len(knownRules) < len(rules) {
		found := false
		for fieldIndex, validRules := range fieldRuleCandidates {
			if len(validRules) == 1 {
				found = true
				ruleName := FirstKey(validRules)
				fmt.Printf("Matched field index %d to rule %q\n", fieldIndex,  ruleName)
				knownRules[fieldIndex] = rulesByName[ruleName]
				eliminateCandidates(ruleName, fieldIndex)
			}
		}
		for ruleName, validFields := range ruleFieldCandidates {
			if len(validFields) == 1 {
				found = true
				fieldIndex := FirstKey(validFields)
				fmt.Printf("Matched rule %q to field index %d\n", ruleName, fieldIndex)
				knownRules[fieldIndex] = rulesByName[ruleName]
				eliminateCandidates(ruleName, fieldIndex)
			}
		}
		if !found {
			for ruleName, validFields := range ruleFieldCandidates {
				fmt.Printf("candidates for rule %q: %#v\n", ruleName, validFields)
			}
			for fieldIndex, validRules := range fieldRuleCandidates {
				fmt.Printf("candidates for field %d: %#v\n", fieldIndex, validRules)
			}
			panic("not able to reduce")
		}
	}

	departureProduct := 1
	for fieldIndex, rule := range knownRules {
		if strings.HasPrefix(rule.Name, "departure ") {
			departureProduct *= yourTicket[fieldIndex]
		}
	}
	fmt.Printf("product of your six departure values: %d\n", departureProduct)
}
