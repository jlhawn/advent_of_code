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
	"time"
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

func loadStartingPositions() (int, int) {
	rawLines := readInputLines()

	starts := []int{0, 0}
	for i, line := range rawLines {
		parts := strings.Split(line, " ")
		start, err := strconv.Atoi(parts[len(parts)-1])
		if err != nil { panic(err) }
		starts[i] = start
	}

	return starts[0]-1, starts[1]-1
}

type GameState struct {
	P1Position, P2Position int
	P1Score, P2Score int

	P1Turn bool
}

func NewGame(p1Start, p2Start int) *GameState {
	return &GameState{
		P1Position: p1Start,
		P2Position: p2Start,
		P1Turn: true,
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

func (g *GameState) PlayDeterministicDie(winScore int) {
	die := NewDeterministicDie(100)

	for g.P1Score < winScore && g.P2Score < winScore {
		rollVal := die.Roll() + die.Roll() + die.Roll()
		if g.P1Turn {
			// P1 Turn
			g.P1Position = (g.P1Position + rollVal) % 10
			g.P1Score += g.P1Position + 1
		} else {
			// P2 Turn
			g.P2Position = (g.P2Position + rollVal) % 10
			g.P2Score += g.P2Position + 1
		}
		g.P1Turn = !g.P1Turn
	}

	fmt.Printf("After %d deterministic die rolls, P1 score = %d, P2 score = %d\n", die.TotalRolls, g.P1Score, g.P2Score)
	lowerScore := g.P1Score
	if g.P2Score < lowerScore {
		lowerScore = g.P2Score
	}
	fmt.Printf("\tlowerScore * numRolls = %d\n", lowerScore * die.TotalRolls)
}

type WinCounts struct {
	P1, P2 int
}

func (a *WinCounts) Add(b WinCounts) {
	a.P1 += b.P1
	a.P2 += b.P2
}

func (a *WinCounts) Multiply(b int) {
	a.P1 *= b
	a.P2 *= b
}

type WinCountCacheKey struct {
	GameState
	rollSum int
}

// WinCountCache maps a GameState and rollSum to the number of wins in all
// universes which could branch from that game state using a Dirac Die where
// the three roll values summed to that rollSum.
type WinCountCache map[WinCountCacheKey]WinCounts

func (g *GameState) PlayDiracDie(winScore int) {
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
	cache := make(WinCountCache)
	wins := diracTurn(true, 0, winScore, *g, cache)
	fmt.Printf("With a dirac die, player 1 wins in %d universes, player 2 wins in %d universes\n", wins.P1, wins.P2)
}

func diracTurn(init bool, rollSum, winScore int, state GameState, cache WinCountCache) (wins WinCounts) {
	cacheKey := WinCountCacheKey{state, rollSum}

	if !init {
		if cached, ok := cache[cacheKey]; ok {
			// Another universe had this same game state and we
			// know the result already.
			return cached
		}

		if state.P1Turn {
			state.P1Position = (state.P1Position + rollSum) % 10
			state.P1Score += state.P1Position + 1
		} else {
			state.P2Position = (state.P2Position + rollSum) % 10
			state.P2Score += state.P2Position + 1
		}

		// Check for a win.
		if state.P1Score >= winScore || state.P2Score >= winScore {
			wins = WinCounts{1, 0}
			if state.P2Score > state.P1Score {
				wins = WinCounts{0, 1}
			}

			cache[cacheKey] = wins
			return wins
		}

		state.P1Turn = !state.P1Turn
	}

	// For the next turn, try all the different possible next roll outcomes.
	for _, combination := range [][]int{{3, 1}, {4, 3}, {5, 6}, {6, 7}, {7, 6}, {8, 3}, {9, 1}} {
		rollSum, ways := combination[0], combination[1]
		universeWins := diracTurn(false, rollSum, winScore, state, cache)
		universeWins.Multiply(ways)
		wins.Add(universeWins)
	}

	cache[cacheKey] = wins
	return wins
}

func main() {
	start := time.Now()
	p1Start, p2Start := loadStartingPositions()

	game := NewGame(p1Start, p2Start)
	game.PlayDeterministicDie(1000)

	game = NewGame(p1Start, p2Start)
	game.PlayDiracDie(21)

	runtime := time.Since(start)
	fmt.Println(runtime)
}
