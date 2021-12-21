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

func initGame() *GameState {
	rawLines := readInputLines()

	starts := []int{0, 0}
	for i, line := range rawLines {
		parts := strings.Split(line, " ")
		start, err := strconv.Atoi(parts[len(parts)-1])
		if err != nil { panic(err) }
		starts[i] = start
	}

	return &GameState{
		P1Start: starts[0]-1,
		P2Start: starts[1]-1,
	}
}

type DeterministicDie struct {
	Counter, Modulus int
	TotalRolls int
}

func NewDeterministicDie(modulus int) *DeterministicDie {
	return &DeterministicDie{Modulus: modulus}
}

func (d *DeterministicDie) Roll() int {
	d.Counter++
	result := d.Counter
	d.Counter %= d.Modulus
	d.TotalRolls++
	return result
}

type GameState struct {
	P1Start, P2Start int

	P1Space, P2Space int
	P1Score, P2Score int

	TurnNo int
}

func (g *GameState) Reset() {
	g.P1Space = g.P1Start
	g.P2Space = g.P2Start
	g.P1Score = 0
	g.P2Score = 0
	g.TurnNo = 0
}

func (g *GameState) PlayDeterministiclyToScore(winScore int) {
	g.Reset()
	die := NewDeterministicDie(100)

	for g.P1Score < winScore && g.P2Score < winScore {
		rollVal := die.Roll() + die.Roll() + die.Roll()
		if g.TurnNo % 2 == 0 {
			// P1 Turn
			g.P1Space = (g.P1Space + rollVal) % 10
			g.P1Score += g.P1Space + 1
		} else {
			// P2 Turn
			g.P2Space = (g.P2Space + rollVal) % 10
			g.P2Score += g.P2Space + 1
		}
		g.TurnNo++
	}

	fmt.Printf("After %d turns and %d deterministic die rolls, P1 score = %d, P2 score = %d\n", g.TurnNo, die.TotalRolls, g.P1Score, g.P2Score)
	lowerScore := g.P1Score
	if g.P2Score < lowerScore {
		lowerScore = g.P2Score
	}
	fmt.Printf("lowerScore * numRolls = %d\n", lowerScore * die.TotalRolls)
}

func (g *GameState) PlayDiracToScore(winScore int) {
	// It's a 3-sided die and we always roll 3x for each player's turn so we don't need to keep
	// track of dice state only the possible number of ways to roll different numbers.
	//
	//   3x Roll Sum | No. of Ways (3^3 = 27 Total)
	//   ------------|-----------------------------
	//          3    |  1     [1, 1, 1]
	//          4    |  3     [1, 1, 2], [1, 2, 1], [2, 1, 1]
	//          5    |  6     [1, 1, 3], [1, 3, 1], [3, 1, 1], [1, 2, 2], [2, 1, 2], [2, 2, 1]
	//          6    |  7     [1, 2, 3], [1, 3, 2], [2, 1, 3], [2, 3, 1], [3, 1, 2], [3, 2, 1], [2, 2, 2]
	//          7    |  6     [2, 2, 3], [2, 3, 2], [3, 2, 2], [1, 3, 3], [3, 1, 3], [3, 3, 1]
	//          8    |  3     [2, 3, 3], [3, 2, 3], [3, 3, 2]
	//          9    |  1     [3, 3, 3]
	//
	// So instead of exploding into 27 different universes for each turn, we only need to keep track of the
	// count of universes with the same outcome.
	// Also: Hey! What do you know? It's the trinomial coefficients!
	g.Reset()
	p1Wins, p2Wins := diracTurn(true, false, 0, 1, g.P1Start, g.P2Start, 0, 0, winScore)
	fmt.Printf("with a dirac die, player 1 wins in %d universes, player 2 wins in %d universes\n", p1Wins, p2Wins)
}

func diracTurn(init, p1Turn bool, rollSum, multiplier, p1Position, p2Position, p1Score, p2Score, winScore int) (p1Wins, p2Wins int) {
	if !init {
		if p1Turn {
			p1Position = (p1Position + rollSum) % 10
			p1Score += p1Position + 1
		} else {
			p2Position = (p2Position + rollSum) % 10
			p2Score += p2Position + 1
		}
	}

	// Check for a win.
	if p1Score >= winScore {
		return multiplier, 0
	}
	if p2Score >= winScore {
		return 0, multiplier
	}

	// For the next turn, try all the different possible next roll outcomes.
	for _, combination := range [][]int{{3, 1}, {4, 3}, {5, 6}, {6, 7}, {7, 6}, {8, 3}, {9, 1}} {
		rollSum, ways := combination[0], combination[1]
		p1, p2 := diracTurn(false, !p1Turn, rollSum, ways*multiplier, p1Position, p2Position, p1Score, p2Score, winScore)
		p1Wins, p2Wins = p1Wins+p1, p2Wins+p2
	}

	return p1Wins, p2Wins
}

func main() {
	game := initGame()
	game.PlayDeterministiclyToScore(1000)
	game.PlayDiracToScore(21)
}
