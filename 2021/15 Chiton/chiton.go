package main

import (
	"bufio"
	"container/heap"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
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

type Coord struct {
	X, Y int
}

type Node struct {
	Coord
	Risk int
}

func (n *Node) Neighbors(g *Graph) []*Node {
	var neighbors []*Node
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if !(dx == 0 || dy == 0) {
				// Not a diagonal
				continue
			}
			coord := Coord{X: n.X+dx, Y: n.Y+dy}
			if neighbor, ok := g.Nodes[coord]; ok {
				neighbors = append(neighbors, neighbor)
			}
		}
	}
	return neighbors
}

type Route struct {
	*Node
	Prev *Route
	Graph *Graph
	TotalRisk int
}

func NewRoute(g *Graph) *Route{
	return &Route{
		Node: g.Start,
		Graph: g,
	}
}

func (r *Route) Options(visited map[Coord]*Route) []*Node {
	var options []*Node
	for _, node := range r.Node.Neighbors(r.Graph) {
		if _, alreadyFound := visited[node.Coord]; !alreadyFound {
			options = append(options, node)
		}
	}
	return options
}

func (r *Route) Next(node *Node) *Route {
	return &Route{
		Node: node,
		Prev: r,
		Graph: r.Graph,
		TotalRisk: r.TotalRisk + node.Risk,
	}
}

type RouteSlice []*Route

func (r RouteSlice) Len() int { return len(r) }
func (r RouteSlice) Less(i, j int) bool {
	return r[i].TotalRisk < r[j].TotalRisk
}
func (r RouteSlice) Swap(i, j int) { r[i], r[j] = r[j], r[i] }
func (r *RouteSlice) Push(x interface{}) {
	item := x.(*Route)
	*r = append(*r, item)
}
func (r *RouteSlice) Pop() interface{} {
	old := *r
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	*r = old[:n-1]
	return item
}

func SortMinPaths(minPaths map[Coord]*Route) []*Route {
	routes := make([]*Route, 0, len(minPaths))
	for _, route := range minPaths {
		routes = append(routes, route)
	}

	sort.Sort(RouteSlice(routes))
	return routes
}

type Graph struct {
	Nodes map[Coord]*Node
	Start, End *Node
	Length, Width int
}

func (g *Graph) DrawPath(route *Route, visited map[Coord]bool) {
	if visited != nil {
		visited[route.Node.Coord] = true
	}

	pathCoords := map[Coord]bool{}
	for route != nil {
		pathCoords[route.Node.Coord] = true
		route = route.Prev
	}

	for j := 0; j < g.Length; j++ {
		for i := 0; i < g.Width; i++ {
			coord := Coord{X: i, Y: j}
			risk := g.Nodes[coord].Risk
			if pathCoords[coord] {
				fmt.Printf("\033[1m %d \033[0m", risk)
			} else if visited[coord] {
				fmt.Printf("   ")
			} else {
				fmt.Printf(" %d ", risk)
			}
		}
		fmt.Println()
	}
}

func (g *Graph) Print() {
	for j := 0; j < g.Length; j++ {
		for i := 0; i < g.Width; i++ {
			coord := Coord{X: i, Y: j}
			fmt.Printf("%d", g.Nodes[coord].Risk)
		}
		fmt.Println()
	}
}

func loadGraph() *Graph {
	nodes := map[Coord]*Node{}
	rawLines := readInputLines()
	var start, end *Node
	for j, rawLine := range rawLines {
		for i, rawRisk := range []byte(rawLine) {
			risk, err := strconv.Atoi(string(rawRisk))
			if err != nil { log.Fatal(err) }

			// fmt.Printf("%d", risk)

			coord := Coord{X: i, Y: j}
			node := &Node{Coord: coord, Risk: risk}
			nodes[coord] = node

			if start == nil {
				start = node
			}
			end = node
		}
		// fmt.Println()
	}
	return &Graph{Nodes: nodes, Start: start, End: end, Length: len(rawLines), Width: len(rawLines[0])}
}

func (g *Graph) determineMinPaths() map[Coord]*Route {
	paths := &RouteSlice{NewRoute(g)}
	visited := map[Coord]*Route{}

	heap.Init(paths)

	for len(*paths) > 0 {
		path := heap.Pop(paths).(*Route)

		at := path.Node.Coord
		if _, alreadyFound := visited[at]; alreadyFound {
			// Found another route to this point earlier.
			// It will always have a lower risk so we need
			// not bother continuing to explore this route.
			continue
		}
		visited[at] = path

		for _, node := range path.Options(visited) {
			heap.Push(paths, path.Next(node))
		}
	}

	return visited
}

func (g *Graph) TileDownAcross(n int) {
	originalNodes := make([]*Node, 0, len(g.Nodes))
	for _, node := range g.Nodes {
		originalNodes = append(originalNodes, node)
	}

	for j := 0; j < n; j++ {
		yStart := j*g.Length
		for i := 0; i < n; i++ {
			if i == 0 && j == 0 {
				continue // We already have the first tile.
			}
			xStart := i*g.Width

			for _, originalNode := range originalNodes {
				newNode := &Node{
					Coord: Coord{X: xStart+originalNode.X, Y: yStart+originalNode.Y},
					Risk: ((originalNode.Risk + i + j - 1) % 9) + 1,
				}

				g.Nodes[newNode.Coord] = newNode
			}
		}
	}

	g.Length *= n
	g.Width *= n

	endCoord := Coord{X: g.Width-1, Y: g.Length-1}
	g.End = g.Nodes[endCoord]
}

func main() {
	graph := loadGraph()
	paths := graph.determineMinPaths()
	pathToEnd := paths[graph.End.Coord]

	animate := false
	if animate {
		sortedPaths := SortMinPaths(paths)
		visited := map[Coord]bool{}
		for _, path := range sortedPaths {
			graph.DrawPath(path, visited)
			time.Sleep(50*time.Millisecond)
			fmt.Print("\033[2J") //Clear screen
			fmt.Printf("\033[%d;%dH", 0, 0) // Set cursor position
		}
	} else {
		graph.DrawPath(pathToEnd, nil)
	}
	
	fmt.Printf("minimum risk path to end has total risk of %d\n", pathToEnd.TotalRisk)

	graph.TileDownAcross(5)
	paths = graph.determineMinPaths()
	pathToEnd = paths[graph.End.Coord]

	if animate {
		sortedPaths := SortMinPaths(paths)
		visited := map[Coord]bool{}
		for _, path := range sortedPaths {
			graph.DrawPath(path, visited)
			time.Sleep(50*time.Millisecond)
			fmt.Print("\033[2J") //Clear screen
			fmt.Printf("\033[%d;%dH", 0, 0) // Set cursor position
		}
	} else {
		graph.DrawPath(pathToEnd, nil)
	}
	
	fmt.Printf("minimum risk path to end has total risk of %d\n", pathToEnd.TotalRisk)
}
