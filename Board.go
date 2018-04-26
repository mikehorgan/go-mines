/*

	Implementation for Board state management in go-minesweeper
	mike@pocomotech.com

*/

package gomines

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
)

type location struct {
	row, col int
}

// cell : manage state for a single cell on the board
type cell struct {
	hasMine  bool // cell holds mine
	score    int  // cache static score for this cell
	flagged  bool // user flag
	revealed bool // all cells start hidden
}

// BoardSaveState : Persistable board state object, read/written as JSON
type boardSaveState struct {
	initialized bool // board starts uninitialized, and then gets populated after player's first 'guaranteed safe' move
	rows        int
	cols        int
	mines       []location
}

// Board struct manages state of the Minesweeper board
type Board struct {
	boardSaveState          // persistable state
	cells          [][]cell // cells of initialized board
	safeRemaining  int      // cache number of non-mine cells remaining to be revealed
	mineCount      int      // number of mines defined for this board
}

/************************************\
** cell Methods
\************************************/

// HasMine : return Mine status for a cell
func (c *cell) HasMine() bool {
	if nil == c {
		return false
	}

	return c.hasMine
}

// Render : return a rune representing the current state of the cell
var scoreRunes = [...]rune{'_', '1', '2', '3', '4', '5', '6', '7', '8'}

func (c *cell) Render() rune {
	if nil == c {
		return '~'
	}

	if !c.revealed {
		return '.'
	} else if c.flagged {
		return '+'
	} else if c.hasMine {
		return '*'
	}

	return scoreRunes[c.score]
}

/************************************\
** Board Methods
\************************************/
type boardparams struct {
	difficulty            string
	rows, cols, mineCount int
}

// static map function of board difficulty parameters
var boardDefinitionsDict = func() map[string]boardparams {
	return map[string]boardparams{
		// name : difficulty, rows, cols, mines
		"easy":   {"easy", 9, 9, 6},
		"medium": {"medium", 16, 16, 18},
		"hard":   {"hard", 30, 16, 36},
	}
}

// NewBoard : allocate new, uninitialized board. Supported sizes are "easy" (9x9), "medium", (16x16) and "hard" (30x16)
func NewBoard(difficulty string) *Board {
	params, ok := boardDefinitionsDict()[difficulty]

	// unrecognized board types rejected
	if !ok {
		return nil
	}

	retval := new(Board)
	retval.rows, retval.cols, retval.mineCount = params.rows, params.cols, params.mineCount

	return retval
}

// Initialize : construct a new Board with consideratioon for user's selected 'safe' location
func (b *Board) Initialize(safespot location) error {

	// Create default cells, then loop over grid and place bombs randomly at 10% probbality until bomb supply exhausted
	b.cells = make([][]cell, b.rows)
	for row := range b.cells {
		b.cells[row] = make([]cell, b.cols)
	}
	b.safeRemaining = b.rows * b.cols

	minesToPlace := b.mineCount
	for minesToPlace > 0 {
		for row := range b.cells {
			for col := range b.cells[row] {
				if minesToPlace == 0 {
					continue
				}

				currloc := location{row, col}
				if currloc == safespot {
					continue // can't place mine at user's safe starting cell
				}
				mineshot := rand.Intn(100)

				if mineshot < 2 {
					currcell := b.getCell(currloc)
					if currcell.hasMine {
						continue // we already placed a mine here
					}
					// place and record mine at current location
					b.cells[row][col].hasMine = true
					b.mines = append(b.mines, currloc)
					minesToPlace--
					b.safeRemaining--
				}
			}
		}
	}

	// once mines are placed, go ahead and calculate cell scores
	initializeScores(b)

	b.initialized = true
	return nil
}

// initializeScores - calculate and set mine proximity scores for each cell
func initializeScores(b *Board) {

	for row := range b.cells {
		for col := range b.cells[row] {
			currloc := location{row, col}
			currcell := b.getCell(currloc)
			cellScore := 0
			// iterate over all neighbor cells
			for nrow := currloc.row - 1; nrow <= (currloc.row + 1); nrow++ {
				for ncol := currloc.col - 1; ncol <= (currloc.col + 1); ncol++ {
					neighborloc := location{nrow, ncol}
					// don't count yourself
					if currloc == neighborloc {
						continue
					}
					neighbor := b.getCell(neighborloc)
					if nil == neighbor { // invalid location outside grid
						continue
					}
					if neighbor.hasMine {
						cellScore++
					}
				}
			}
			currcell.score = cellScore
		}
	}

}

// Initialized : return board initilization status
func (b *Board) Initialized() bool {
	if nil == b {
		return false
	}
	return b.initialized
}

// GetCell : return a reference to a particular cell
func (b *Board) getCell(selected location) *cell {
	// bunch of preconditions
	if selected.row < 0 || selected.row >= b.rows || selected.col < 0 || selected.col >= b.cols {
		return nil
	}
	return &b.cells[selected.row][selected.col]
}

// SafeRemaining : report number of unrevealed non-mine cells remaining. Win condition is when this number reaches 0
func (b *Board) SafeRemaining() int {
	if nil == b || !b.initialized {
		return 0
	}
	return b.safeRemaining
}

// RevealAll : set all cells to revealed (for debugging or surrender); this is irreversible
func (b *Board) RevealAll() error {
	if nil == b || !b.initialized {
		return errors.New("called RevealAll() on an uninitialized board")
	}
	for row := range b.cells {
		for col := range b.cells[row] {
			b.cells[row][col].revealed = true
		}
	}

	return nil
}

// ConsoleRender -- render a console image of the board state
func (b *Board) ConsoleRender(cout io.Writer) error {

	if nil == b || !b.initialized {
		return errors.New("called Render() on an uninitialized board")
	}
	for row := range b.cells {
		nextLine := ""
		for col := range b.cells[row] {
			if col != 0 {
				nextLine += " "
			}
			nextLine += string(b.cells[row][col].Render())
		}
		fmt.Fprintln(cout, nextLine)
	}

	return nil
}
