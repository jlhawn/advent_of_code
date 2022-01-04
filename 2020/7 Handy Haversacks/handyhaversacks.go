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

type BagColorCount struct {
	Color string
	Count int
}

type BagRule struct {
	Color string
	Contents []BagColorCount
}

type BagRuleSet map[string]BagRule

func parseBagRule(rawRule string) BagRule {
	rawRule = strings.TrimSuffix(rawRule, ".") // Each line ends in a period.
	color, contents, ok := strings.Cut(rawRule, " bags contain ")
	if !ok {
		panic(fmt.Errorf("unable to cut raw rule into color and contents: %q", rawRule))
	}

	if contents == "no other bags" {
		return BagRule{Color: color}
	}

	parts := strings.Split(contents, ", ")
	colorCounts := make([]BagColorCount, len(parts))
	for i, part := range parts {
		part = strings.TrimSuffix(part, " bag") // For single bag.
		part = strings.TrimSuffix(part, " bags") // For multiple bags.
		rawCount, color, ok := strings.Cut(part, " ")
		if !ok {
			panic(fmt.Errorf("unable to cut rule contents: %q", part))
		}
		count, err := strconv.Atoi(rawCount)
		if err != nil { panic(err) }
		colorCounts[i] = BagColorCount{Color: color, Count: count}
	}

	return BagRule{Color: color, Contents: colorCounts}
}

func loadBagRules() BagRuleSet {
	lines := readInputLines()

	ruleSet := make(BagRuleSet, len(lines))
	for _, line := range lines {
		rule := parseBagRule(line)
		ruleSet[rule.Color] = rule
	}

	return ruleSet
}

type ColorPair struct {
	A, B string
}

type CanContainCache map[ColorPair]bool

func (b BagRule) CanContain(color string, ruleSet BagRuleSet, cache CanContainCache) bool {
	cacheKey := ColorPair{b.Color, color}
	if cache[cacheKey] {
		return true
	}

	for _, content := range b.Contents {
		if content.Color == color {
			cache[cacheKey] = true
			return true
		}
	}

	for _, content := range b.Contents {
		if ruleSet[content.Color].CanContain(color, ruleSet, cache) {
			cache[cacheKey] = true
			return true
		}
	}

	return false
}

type InnerBagCountCache map[string]int

func (b BagRule) InnerBagCount(ruleSet BagRuleSet, cache InnerBagCountCache) int {
	cacheKey := b.Color
	if count, ok := cache[cacheKey]; ok {
		return count
	}

	count := 0
	for _, content := range b.Contents {
		// Multiply each sub-count times the type of bag itself plus the number of bags within it.
		count += content.Count * (1 + ruleSet[content.Color].InnerBagCount(ruleSet, cache))
	}

	cache[cacheKey] = count
	return count
}

func main() {
	ruleSet := loadBagRules()

	cache := CanContainCache{}
	for _, rule := range ruleSet {
		rule.CanContain("shiny gold", ruleSet, cache)
	}
	fmt.Printf("shiny gold bags can be contained in %d other colors of bag\n", len(cache))

	innerCount := ruleSet["shiny gold"].InnerBagCount(ruleSet, InnerBagCountCache{})
	fmt.Printf("shiny gold bags require %d inner bags\n", innerCount)
}
