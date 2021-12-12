package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	// "sort"
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

type Graph struct {
	Caves map[string]bool
	LargeCaves map[string]bool
	Routes map[string]map[string]bool
}

func loadGraph() *Graph {
	rawLines, err := readInputLines()
	if err != nil {
		log.Fatal(err)
	}

	caves := map[string]bool{}
	largeCaves := map[string]bool{}
	routes := map[string]map[string]bool{}

	for _, rawLine := range rawLines {
		// Each line defines an edge.
		parts := strings.Split(rawLine, "-")
		a, b := parts[0], parts[1]

		caves[a] = true
		caves[b] = true

		if strings.ToLower(a) != a {
			largeCaves[a] = true
		}
		if strings.ToLower(b) != b {
			largeCaves[b] = true
		}

		if routes[a] == nil {
			routes[a] = map[string]bool{}
		}
		if routes[b] == nil {
			routes[b] = map[string]bool{}
		}
		routes[a][b] = true
		routes[b][a] = true
	}

	return &Graph{
		Caves: caves,
		LargeCaves: largeCaves,
		Routes: routes,
	}
}

type PathFinder struct{
	PathSoFar []string
	CavesSoFar map[string]bool
	SmallCaveTwice bool
}

func NewPathFinder() *PathFinder {
	return &PathFinder{PathSoFar: []string{"start"}, CavesSoFar: map[string]bool{"start": true}}
}

func (p *PathFinder) CanVisit(g *Graph, cave string) bool {
	if !p.CavesSoFar[cave] {
		return true // Haven't visited this cave yet.
	}

	// Can visit if it's a large cave or it's not "start" and we haven't yet visited any small cave twice.
	return g.LargeCaves[cave] || (cave != "start" && !p.SmallCaveTwice)
}

func (p *PathFinder) CurrentCave() string {
	return p.PathSoFar[len(p.PathSoFar)-1]
}

func (p *PathFinder) IsComplete() bool {
	return p.CurrentCave() == "end"
}

func (p *PathFinder) Options(g *Graph) []string {
	var options []string
	neighbors := g.Routes[p.CurrentCave()]
	for cave := range neighbors {
		if p.CanVisit(g, cave) {
			options = append(options, cave)
		}
	}
	return options
}

func (p *PathFinder) Next(g *Graph, nextCave string) *PathFinder {
	copiedPath := make([]string, len(p.PathSoFar), len(p.PathSoFar)+1)
	for i, cave := range p.PathSoFar {
		copiedPath[i] = cave
	}
	copiedPath = append(copiedPath, nextCave)

	copiedCaves := make(map[string]bool, len(p.CavesSoFar)+1)
	for cave := range p.CavesSoFar {
		copiedCaves[cave] = true
	}
	copiedCaves[nextCave] = true

	smallCaveTwice := p.SmallCaveTwice
	if !smallCaveTwice && !g.LargeCaves[nextCave] {
		// This is a small cave and we haven't yet visited a small cave twice.
		smallCaveTwice = p.CavesSoFar[nextCave]
	}

	return &PathFinder{PathSoFar: copiedPath, CavesSoFar: copiedCaves, SmallCaveTwice: smallCaveTwice}	
}

func findPaths(g *Graph) []*PathFinder {
	var completedPaths []*PathFinder

	stack := []*PathFinder{NewPathFinder()}
	for len(stack) > 0 {
		popIdx := len(stack)-1
		popped := stack[popIdx]
		stack = stack[:popIdx]

		if popped.IsComplete() {
			completedPaths = append(completedPaths, popped)
			continue
		}

		for _, nextCave := range popped.Options(g) {
			stack = append(stack, popped.Next(g, nextCave))
		}
	}

	fmt.Printf("There are %d complete paths.\n", len(completedPaths))
	return completedPaths
}

func main() {
	findPaths(loadGraph())
}
