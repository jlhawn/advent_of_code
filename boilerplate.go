package main

import (
	"bufio"
	"constraints"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

func Max[T constraints.Ordered](vals ...T) T {
	var max T
	if len(vals) == 0 {
		return max
	}
	max = vals[0]
	for _, val := range vals[1:] {
		if val > max {
			max = val
		}
	}
	return max
}

func Min[T constraints.Ordered](vals ...T) T {
	var min T
	if len(vals) == 0 {
		return min
	}
	min = vals[0]
	for _, val := range vals[1:] {
		if val < min {
			min = val
		}
	}
	return min
}

func Map[T1, T2 any](f func(T1) T2, vals ...T1) []T2 {
	mapped := make([]T2, len(vals))
	for i, val := range vals {
		mapped[i] = f(val)
	}
	return mapped
}

func Reduce[T any](f func(T, T) T, init T, vals ...T) T {
	reduced := init
	for _, val := range vals {
		reduced = f(reduced, val)
	}
	return reduced
}

func Sum[T constraints.Integer | constraints.Float](vals ...T) T {
	var zero T
	return Reduce(func(a, b T) T { return a + b }, zero, vals...)
}

func Filter[T any](f func(T) bool, vals ...T) []T {
	filtered := make([]T, 0, len(vals))
	for _, val := range vals {
		if f(val) {
			filtered = append(filtered, val)
		}
	}
	return filtered
}

func ForEach[T any](f func(T), vals ...T) {
	for _, val := range vals {
		f(val)
	}
}

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