package main

import (
	"bufio"
	"constraints"
	"container/heap"
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
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

func solvePuzzle() map[Coord]*Route {
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

	return paths
}

type Number interface{
	constraints.Integer | constraints.Float
}

func remap[F, T Number](val, fromMin, fromMax F, toMin, toMax T) T {
	fromDiff := val - fromMin
	fromRange := fromMax - fromMin
	percent := float64(fromDiff) / float64(fromRange)

	toRange := toMax - toMin
	toDiff := percent * float64(toRange)

	return toMin + T(toDiff)
}

func hueToRgb(v1, v2, vH float64) uint8 {
	// remap vH from [0, 360) to [0, 1) to use with our formulas.
	vH = remap(vH, 0, 360, 0.0, 1.0)

	var ret float64
	if 6*vH < 1 {
		ret = v2 + (v1-v2)*6*vH
	} else if 2*vH < 1 {
		ret = v1
	} else if 3*vH < 2 {
		ret = v2 + (v1-v2)*(4 - 6*vH)
	} else {
		ret = v2
	}

	// Before we return, remap it to [0, 255)
	return remap(ret, 0, 1, uint8(0), uint8(255))
}

func hslToRgb(h, s, l float64) (r, g, b uint8) {
	if s == 0 {
		// No saturation means that it's purely grayscale.
		// Each RGB value is the same from 0 to 255.
		// We only need to remap to lightness.
		gray := remap(l, 0, 1, uint8(0), uint8(255))
		return gray, gray, gray
	}

	// These temporary variables make the formula easier to
	// understand.
	var v1, v2 float64
	if l < 0.5 {
		// At 0% saturation this will be a value between 0 and 0.5,
		// at 50% saturation this will be a value between 0
		// and 0.75, at 100% saturation this will be a value between
		// 0 and 1.
		v1 = l * (1 + s)
	} else {
		// At 0% saturation this will be a value between 0.5 and 1,
		// at 50% saturation this will be a value between 0.75 and 1,
		// at 100% saturation this value will always be 1.
		v1 = (l + s) - (s * l)
	}

	// At 0% lightness this value will always be 0.
	// At 50% lightness this value will range from 1 at 0% saturation,
	// to 0.25 at 50% saturation, and to 0 at 100% saturation.
	// At 100% lightness this value will always be 1
	v2 = 2*l - v1

	// RGB color channel values range from 0 to 255 for 120 degrees then
	// back to 0 for another 120 degrees. Then stay at 0 for 120 degrees
	// before going back around again. Each color channel is out of
	// phase by 120 degrees with one another. Red from -120 to 120,
	// Green from 0 to 240, and Blue from 120 to 360. These need to
	// be shifted so that each range starts at zero.
	rH := h + 120
	gH := h // No change required.
	bH := h - 120

	// These might wrap around the cirle.
	if rH > 360 {
		rH -= 360
	}
	if bH < 0 {
		bH += 360
	}

	// We then use a helper function to convert these variables to RGB channel
	// values.
	r = hueToRgb(v1, v2, rH)
	g = hueToRgb(v1, v2, gH)
	b = hueToRgb(v1, v2, bH)

	return r, g, b
}

type ImagePoint struct {
	Coord
	TotalRisk int
	PathCount int
}

func createImage(paths map[Coord]*Route) {
	imagePoints := make([]ImagePoint, 0, len(paths))
	coordImagePoints := make(map[Coord]*ImagePoint, len(paths))
	for coord, path := range paths {
		imagePoints = append(imagePoints, ImagePoint{
			Coord: coord,
			TotalRisk: path.TotalRisk,
		})
		coordImagePoints[coord] = &imagePoints[len(imagePoints)-1]
	}

	maxTotalRisk := 0
	for _, path := range paths {
		if path.TotalRisk > maxTotalRisk {
			maxTotalRisk = path.TotalRisk
		}

		// Walk the path and increment the path count for each point.
		route := path
		for route != nil {
			coordImagePoints[route.Node.Coord].PathCount++
			route = route.Prev
		}
	}

	maxPathCount := 0
	maxX, maxY := 0, 0
	for _, imgPoint := range imagePoints {
		if imgPoint.PathCount > maxPathCount {
			maxPathCount = imgPoint.PathCount
		}
		if imgPoint.X > maxX {
			maxX = imgPoint.X
		}
		if imgPoint.Y > maxY {
			maxY = imgPoint.Y
		}
	}

	// Now have a map containing each point with its total risk and the number
	// of paths that visit that point. From this, we can make a png image where
	// risk is mapped to a hue and path count is mapped to lightness.
	upLeft := image.Point{0, 0}
	lowRight := image.Point{maxX+1, maxY+1}
	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Most points have path counts closer to zero while the most have very many.
	// Ordering the points by path count I was able to see what looked like an
	// exponential function near the end of the distribution so for the visualization
	// I am scaling the path counts by taking their natural log in order to get a
	// much more gradual color gradient along the paths. If I don't do thes then
	// most points appear very dark (because their path count is so low) compared to
	// the few points with very high path counts.
	maxPathCountLog := math.Log(float64(maxPathCount))

	// Set color for each pixel.
	for x := 0; x <= maxX; x++ {
		for y := 0; y <= maxY; y++ {
			imgPoint := coordImagePoints[Coord{X: x, Y: y}]

			hue := remap(imgPoint.TotalRisk, 0, maxTotalRisk, 0.0, 360.0)
			lightness := remap(math.Log(float64(imgPoint.PathCount)), 0.0, maxPathCountLog, 0.0, 1.0)

			r, g, b := hslToRgb(hue, 1.0, lightness)

			rgbColor := color.RGBA{R: r, G: g, B: b, A: 255}
			img.Set(x, y, rgbColor)
		}
	}

	imgFile, err := os.Create("./visualization.png")
	if err != nil { log.Fatal(err) }
	defer imgFile.Close()

	if err := png.Encode(imgFile, img); err != nil {
		log.Fatal(err)
	}
}

func main() {
	paths := solvePuzzle()
	createImage(paths)
}
