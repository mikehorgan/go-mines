/*

	Implementation for Board state management in go-minesweeper
	mike@pocomotech.com

*/

package msboard

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
)

// Location : zero-based cell location, {0,0} is upper left
type Location struct {
	row, col int
}

// NewLocation -- public interface to create a Location struct
func NewLocation(row, col int) Location {
	retval := Location{row, col}
	return retval
}

// cell : manage state for a single cell on the board
type cell struct {
	location Location // cell position in grid, zero based, {0,0} is upper left
	hasMine  bool     // cell holds mine
	score    int      // cache static score for this cell
	flagged  bool     // user flag
	revealed bool     // all cells start hidden
}

// BoardSaveState : Persistable board state object, read/written as JSON
type boardSaveState struct {
	initialized      bool // board starts uninitialized, and then gets populated after player's first 'guaranteed safe' move
	difficulty       string
	rows             int
	cols             int
	mines            []Location
	explosionOccured bool
}

// Board struct manages state of the Minesweeper board
type Board struct {
	boardSaveState           // persistable state
	cells          [][]*cell // cells of initialized board
	safeRemaining  int       // cache number of non-mine cells remaining to be revealed
	mineCount      int       // number of mines defined for this board
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
		"easy":   {"easy", 9, 9, 10},
		"medium": {"medium", 16, 16, 30},
		"hard":   {"hard", 30, 16, 72},
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
	retval.difficulty, retval.rows, retval.cols, retval.mineCount = difficulty, params.rows, params.cols, params.mineCount

	return retval
}

// Initialize : construct a new Board with consideratioon for user's selected 'safe' Location
func (b *Board) Initialize(safespot Location) error {

	// Create default cells, then loop over grid and place bombs randomly at 10% probbality until bomb supply exhausted
	b.cells = make([][]*cell, b.rows)
	for row := range b.cells {
		b.cells[row] = make([]*cell, b.cols)
		for col := range b.cells[row] {
			b.cells[row][col] = new(cell)
			b.cells[row][col].location = NewLocation(row, col)
		}
	}
	b.safeRemaining = b.rows * b.cols

	minesToPlace := b.mineCount
	for minesToPlace > 0 {
		for row := range b.cells {
			for col := range b.cells[row] {
				if minesToPlace == 0 {
					continue
				}

				currloc := Location{row, col}
				if currloc == safespot {
					continue // can't place mine at user's safe starting cell
				}
				mineshot := rand.Intn(100)

				if mineshot < 2 {
					currcell := b.getCell(currloc)
					if currcell.hasMine {
						continue // we already placed a mine here
					}
					// place and record mine at current Location
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
			currloc := Location{row, col}
			currcell := b.getCell(currloc)
			cellScore := 0
			// iterate over all neighbor cells
			neighbors := b.getNeighborCells(currloc)
			if nil == neighbors {
				fmt.Fprintln(os.Stderr, "Board init failure for cell (this should not happen :() :  ", currloc)
			}

			for _, neighbor := range neighbors {
				if neighbor.hasMine {
					cellScore++
				}
			}
			currcell.score = cellScore
		}
	}

}

// GetNeighborCells - return array of pointers to all valid neighbor cells given a cell location
func (b *Board) getNeighborCells(loc Location) []*cell {
	// sanity check
	center := b.getCell(loc)
	if nil == center {
		return nil
	}

	retval := make([]*cell, 0, 8)

	// iterate over all potential neighbor cell position
	for nrow := loc.row - 1; nrow <= (loc.row + 1); nrow++ {
		for ncol := loc.col - 1; ncol <= (loc.col + 1); ncol++ {
			neighborloc := Location{nrow, ncol}
			// don't include center point
			if loc == neighborloc {
				continue
			}
			neighbor := b.getCell(neighborloc)
			if nil == neighbor { // invalid Location outside grid
				continue
			}
			retval = append(retval, neighbor)
		}
	}

	return retval
}

// Initialized : return board initilization status
func (b *Board) Initialized() bool {
	if nil == b {
		return false
	}
	return b.initialized
}

// GetCell : return a reference to a particular cell
func (b *Board) getCell(selected Location) *cell {
	// bunch of preconditions
	if selected.row < 0 || selected.row >= b.rows || selected.col < 0 || selected.col >= b.cols {
		return nil
	}
	return b.cells[selected.row][selected.col]
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

	// top line is header
	headingLine := ""
	switch b.difficulty {
	case "easy":
		headingLine = "    A  B  C  D  E  F  G  H  I"
	case "medium", "hard":
		headingLine = "    A  B  C  D  E  F  G  H  I  J  K  L  M  N  O  P"
	}
	fmt.Fprintln(cout, headingLine)

	for row := range b.cells {
		// index column along left side
		nextLine := fmt.Sprintf("%2d  ", row+1)

		for col := range b.cells[row] {
			if col != 0 {
				nextLine += "  "
			}
			nextLine += string(b.cells[row][col].Render())
		}
		fmt.Fprintln(cout, nextLine)
	}

	return nil
}

// Click -- Calculate and apply board state changes for a cell click event
func (b *Board) Click(l Location) {
	c := b.getCell(l)

	if nil == c {
		return
	}

	// flagged cells are protected from inadvertant clicks
	if c.flagged {
		return
	}

	// already revealed cells do not respond to clicks
	if c.revealed {
		return
	}

	// reveal cell
	c.revealed = true

	// Mine? Explode
	if c.hasMine {
		b.explosionOccured = true
		return
	}

	// non-zero score cells do not propagate (I think)
	if c.score == 0 {
		// propagate reveals for zero score cells
		b.PropagateReveals(c)
	}

}

// PropagateReveals -- clicking on a zero score cell reveals all connected zero score cells
func (b *Board) PropagateReveals(c *cell) {
	if nil == c {
		return
	}

	neighbors := b.getNeighborCells(c.location)
	// fmt.Fprintln(os.Stderr, "PropagateReveals: ", c.location, " has ", len(neighbors), " neighbors.")

	if nil == neighbors {
		fmt.Fprintln(os.Stderr, "PropogateReveals failure for cell (this should not happen :() :  ", c.location)
	}

	// reveal unrevealed neighbors and recurse for any zero-scored ones
	for _, n := range neighbors {
		if n.revealed {
			continue
		}

		n.revealed = true

		// debug
		// fmt.Fprintln(os.Stderr, "Revealing ", n.location, " (score = ", n.score, ") from ", c.location)

		if n.score == 0 {
			b.PropagateReveals(n)
		}
	}

}

// MineHit -- convenience function for game loop
func (b *Board) MineHit() bool {
	return b.explosionOccured
}

// ToggleFlag -- toggle flag status for a cell, ignored for non-hidden cells
func (b *Board) ToggleFlag(l Location) {
	c := b.getCell(l)

	if nil != c && c.revealed == false {
		c.flagged = !c.flagged
	}
}

// ValidLocation -- return true if selected location is valid for the board
func (b *Board) ValidLocation(l Location) bool {
	if l.row >= 0 && l.row < b.rows && l.col >= 0 && l.col < b.cols {
		return true
	}

	return false
}
