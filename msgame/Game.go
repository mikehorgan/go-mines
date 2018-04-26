/*

	Game.go - minesweeper game logic loop

*/

// Package msgame -- Game logic/play implemention for Go Minesweeper
package msgame

import (
	"bufio"
	"fmt"
	"go-mines/msboard"
	"io"
	"math/rand"
	"strings"
	"time"
)

// Game : main minesweeper game runner class
type Game struct {
	start     time.Time
	turnCount int
}

//New -- init a new Game object
func New() *Game {
	retval := new(Game)
	retval.start = time.Now()

	return retval
}

// RunConsole -- run a fame loop using Console rendering to the provided input/output objects
func (g *Game) RunConsole(cin io.Reader, cout io.Writer) error {

	/* Game loop:
	- Choose Game Type
	- Display Board
	- Choose Move
	- Update Board
	- Display board

	until board.HitMine() or board.SafeRemaining() == 0
	*/

	// get random
	rand.Seed(time.Now().UnixNano())

	// buffered input reader and writer
	in := bufio.NewScanner(cin)
	out := bufio.NewWriter(cout)

	// Outer loop
	for {
		fmt.Fprintln(cout, "Welcome to Minesweeper. Choose game type: [E]asy [M]edium [H]ard   or   [Q]uit")
		input, err := readOneCharacter(in)
		if err != nil {
			continue
		}

		boardType := "unknown"

		switch input {
		case "e":
			boardType = "easy"
		case "m":
			boardType = "medium"
		case "h":
			boardType = "hard"
		case "q":
			goto game_over
		default:
			continue
		}

		board := msboard.NewBoard(boardType)

		// have to init board before displaying initial blank board; re-init after user chooses first square
		board.Initialize(msboard.NewLocation(0, 0))
		board.ConsoleRender(cout)

		for !board.MineHit() && board.SafeRemaining() > 0 {
			fmt.Fprint(out, "Choose next move :  ")
			out.Flush()

			location, err := readNextMove(in)
			if err != nil {
				continue
			}
			fmt.Fprintln(out, location)
		}

	}

game_over:
	return nil
}

// readNextMove -- read and parse an input line to a cell location
func readNextMove(in *bufio.Scanner) (msboard.Location, error) {
	_, err := readInput(in)
	if err != nil {
		return msboard.NewLocation(-1, -1), err
	}

	return msboard.NewLocation(7, 6), nil
}

// readOneCharacter -- consume a line of input but return only the first non-whitespace character
func readOneCharacter(in *bufio.Scanner) (string, error) {
	inLine, err := readInput(in)
	if err != nil {
		return "", err
	}

	return inLine[0:1], nil
}

func readInput(in *bufio.Scanner) (string, error) {
	if !in.Scan() {
		return "", fmt.Errorf("Error or EOF during console read")
	}

	line := strings.Trim(in.Text(), " \n")
	line = strings.ToLower(line)
	return line, nil
}
