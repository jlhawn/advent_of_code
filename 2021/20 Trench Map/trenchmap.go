package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	// "sort"
	"strconv"
	"strings"
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

func loadImage() *Image {
	rawLines := readInputLines()

	// First line is the enhancement algorithm.
	enhancementAlgorithm := make(map[int]bool, 512)
	for i, bit := range rawLines[0] {
		if bit == '#' {
			enhancementAlgorithm[i] = true
		}
	}

	// Second line is empty.

	litPixels := map[Point]bool{}

	// Following lines are the initial image.
	rawLines = rawLines[2:]
	for y, line := range rawLines {
		for x, pixel := range []byte(line) {
			if pixel == '#' {
				litPixels[Point{x, y}] = true
			}
		}
	}

	topLeft := Point{-1, -1}
	bottomRight := Point{len(rawLines[0]), len(rawLines)}

	return &Image{
		LitPixels: litPixels,
		TopLeft: topLeft,
		BottomRight: bottomRight,
		EnhancementAlgorithm: enhancementAlgorithm,
	}
}

type Point struct {
	X, Y int
}

func (p Point) Get3x3() []Point {
	points := make([]Point, 0, 9)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			points = append(points, Point{p.X+dx, p.Y+dy})
		}
	}
	return points
}

type Image struct {
	LitPixels map[Point]bool
	TopLeft, BottomRight Point

	EnhancementAlgorithm map[int]bool
	FringeOn bool
}

func (img *Image) Print() {
	for y := img.TopLeft.Y; y <= img.BottomRight.Y; y++ {
		for x := img.TopLeft.X; x <= img.BottomRight.X; x++ {
			p := Point{x, y}
			if img.LitPixels[p] {
				fmt.Printf("#")
			} else {
				fmt.Printf(".")
			}
		}
		fmt.Println()
	}
	fmt.Println()
}

func (img *Image) IsWithinBounds(p Point) bool {
	return img.TopLeft.X <= p.X && p.X <= img.BottomRight.X && img.TopLeft.Y <= p.Y && p.Y <= img.BottomRight.Y
}

func (img *Image) Convert3x3toInt(points []Point) int {
	var b strings.Builder
	b.Grow(9)
	for _, point := range points {
		if img.IsWithinBounds(point) {
			if img.LitPixels[point] {
				b.WriteByte('1')
			} else {
				b.WriteByte('0')
			}
		} else {
			if img.FringeOn {
				b.WriteByte('1')
			} else {
				b.WriteByte('0')
			}
		}
	}

	val, err := strconv.ParseInt(b.String(), 2, 32)
	if err != nil { panic(err) }

	return int(val)
}

func (img *Image) Enhance() {
	enhanced := make(map[Point]bool, len(img.LitPixels)*2)
	for y := img.TopLeft.Y-1; y <= img.BottomRight.Y+1; y++ {
		for x := img.TopLeft.X-1; x <= img.BottomRight.X+1; x++ {
			p := Point{x, y}
			index := img.Convert3x3toInt(p.Get3x3())
			if img.EnhancementAlgorithm[index] {
				enhanced[p] = true
			}
		}
	}

	img.TopLeft = Point{img.TopLeft.X-1, img.TopLeft.Y-1}
	img.BottomRight = Point{img.BottomRight.X+1, img.BottomRight.Y+1}

	if img.EnhancementAlgorithm[0] {
		// The first bit in the algorithm being set means that the infinite
		// expanse may change between lit and unlit with each enhancement step.

		if !img.FringeOn {
			img.FringeOn = true
		} else {
			img.FringeOn = img.EnhancementAlgorithm[511]
		}
	}

	img.LitPixels = enhanced
}

func main() {
	img := loadImage()

	for i := 0; i < 2; i++ {
		img.Enhance()
	}
	img.Print()

	fmt.Printf("lit pixels after 2 enhance steps: %d\n", len(img.LitPixels))

	for i := 0; i < 48; i++ {
		img.Enhance()
	}
	img.Print()

	fmt.Printf("lit pixels after 50 enhance steps: %d\n", len(img.LitPixels))
}
