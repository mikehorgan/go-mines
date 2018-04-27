/*

	Game.go - minesweeper game logic loop

	mike@pocomotech.com

*/

// Package msgame -- Game logic/play implemention for Go Minesweeper
package msgame

import (
	"bufio"
	"fmt"
	"go-mines/msboard"
	"io"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// Game : main minesweeper game runner class
type Game struct {
	start     time.Time
	turnCount int
	randSeed  int64
}

//New -- init a new Game object with given random seed for testing
func New(seed int64) *Game {
	retval := new(Game)
	retval.start = time.Now()
	retval.randSeed = seed

	return retval
}

// RunConsole -- run a game loop using Console rendering to the provided input/output objects
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
	rand.Seed(g.randSeed)
	// output seed on stderr for potential replay in debugger
	fmt.Fprintf(os.Stderr, "{ starting with random seed %d }\n\n", g.randSeed)

	// buffered reader and writer
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

		// have to init board before displaying initial blank board; re-init after user chooses safe square
		board.Initialize(msboard.NewLocation(0, 0))
		board.ConsoleRender(out)

		gameInit := false
		for !board.MineHit() && board.SafeRemaining() > 0 {

			if !gameInit {
				fmt.Fprint(out, "\nChoose starting cell location:  ")
			} else {
				fmt.Fprint(out, "\nChoose command (s,f) & location :  ")
			}
			out.Flush()

			cmd, location, err := readNextMove(in)
			if err != nil {
				fmt.Fprintln(os.Stderr, "readNextmove() failure: cmd ", cmd, " location ", location, " err ", err)
				continue
			}
			fmt.Fprintln(out, location)

			// sanity check
			if !board.ValidLocation(location) {
				fmt.Fprint(out, "Invalid board location selected, please retry: ", location)
				continue
			}

			if !gameInit {
				// game starts now with user's 'safe' square
				board.Initialize(location)
				gameInit = true
			}

			switch cmd {
			case "s":
				board.Click(location)
			case "f":
				board.ToggleFlag(location)
			default:
				fmt.Fprintf(out, "Invalid command selection %q\n", cmd)
			}

			board.ConsoleRender(out)
		}

	}

game_over:
	return nil
}

// readNextMove -- read and parse an input line into a cell location
func readNextMove(in *bufio.Scanner) (string, msboard.Location, error) {
	/*
	   A move is picking a cell position, which are numbered for rows and letters for columns
	   The intent is to allow teh user to specify a row+column combo in whatever order they prefer
	   We'll gather the digits and letters separately to figure out the intended location
	*/

	inLine, err := readInput(in)
	if err != nil {
		return "", msboard.NewLocation(-1, -1), err
	}
	digits := ""
	letters := make([]rune, 0)
	inputRunes := []rune(inLine)
	for i := 0; i < len(inputRunes); i++ {
		if unicode.IsDigit(inputRunes[i]) {
			digits += string(inputRunes[i])
		} else {
			letters = append(letters, inputRunes[i])
		}
	}

	userRow, err := strconv.Atoi(digits)
	if err != nil {
		userRow = -1
	}
	userRow-- // offset because locations are 0 based, user sees 1 as first row

	userCol := -1
	if len(letters) > 0 {
		userCol = int(letters[0]) - int('a')
	}

	return "s", msboard.NewLocation(userRow, userCol), err
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
