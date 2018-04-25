/*
	Test functions for Minesweeper Board class

	mike@pocomotech.com
*/

package gomines

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
)

/*
	TestBoardCreation -- Board currently supports only 3 sizes
		9x9 (easy)
		16x16 (medium)
		30x16 (hard)
*/
func TestBoardCreation(t *testing.T) {
	var cases = []struct {
		difficulty string
		rows       int
		cols       int
		want       bool
	}{
		{"", 0, 0, false},
		{"easy", 9, 9, true},
		{"medium", 16, 16, true},
		{"hard", 30, 16, true},
		{"nightmare", 1024, 1024, false},
	}

	for _, testcase := range cases {
		got := NewBoard(testcase.difficulty)
		if (got != nil) != testcase.want {
			t.Errorf("NewBoard() failed for %q got %v", testcase.difficulty, got)
		}

		// check returned board shape
		if got != nil && (got.rows != testcase.rows || got.cols != testcase.cols) {
			t.Errorf("NewBoard() returned incorrect shape. Expected %dx%d, got %dx%d", testcase.rows, testcase.cols, got.rows, got.cols)
		}
	}
}

/*
	TestBoardInitialization -- board is initted after the user selects their first cell, which is guaranteed to be a non-mine by the rules
*/
func TestBoardInitialization(t *testing.T) {

	boardTypes := []boardparams{boardDefinitionsDict()["easy"], boardDefinitionsDict()["medium"], boardDefinitionsDict()["hard"]}

	for _, bt := range boardTypes {
		b := NewBoard(bt.difficulty)
		if b == nil {
			t.Errorf("Board Creation failed for difficulty %q", bt.difficulty)
			continue
		}

		// Initialize with random starting location
		startingLocation := location{rand.Intn(bt.rows), rand.Intn(bt.cols)}
		ok := b.Initialize(startingLocation)
		if ok != nil {
			t.Errorf("Board init for type %q failed with error %q.", bt.difficulty, ok)
			continue
		}

		safeWanted := (bt.rows * bt.cols) - bt.mineCount
		safeGot := b.SafeRemaining()
		if safeGot != safeWanted {
			t.Errorf("Board post-init SafeRemaining() count wrong. Game type %q wanted %d got %d", bt.difficulty, safeWanted, safeGot)
		}

		mineCountWanted := bt.mineCount
		mineCountGot := countMineCells(b)
		if mineCountWanted != mineCountGot {
			t.Errorf("Board post-init mine count verify failed. Game type %q wanted %d got %d", bt.difficulty, mineCountWanted, mineCountGot)

		}
	}
}

// Manually tally up the number of cells containing mines, for verification
func countMineCells(b *Board) int {
	if !b.Initialized() {
		return 0
	}

	retval := 0
	for r := 0; r < b.rows; r++ {
		for c := 0; c < b.rows; c++ {
			testcell := b.getCell(location{r, c})
			if testcell.HasMine() {
				retval++
			}
		}
	}

	return retval
}

func TestBoardRender(t *testing.T) {
	rand.Seed(1995) // want same test sequence each time

	boardTypes := []boardparams{boardDefinitionsDict()["easy"], boardDefinitionsDict()["medium"], boardDefinitionsDict()["hard"]}

	for _, bt := range boardTypes {
		b := NewBoard(bt.difficulty)
		if b == nil {
			t.Errorf("Board Creation failed for difficulty %q", bt.difficulty)
			continue
		}

		// Initialize with random starting location
		startingLocation := location{rand.Intn(bt.rows), rand.Intn(bt.cols)}
		ok := b.Initialize(startingLocation)
		if ok != nil {
			t.Errorf("Board init for type %q failed with error %q.", bt.difficulty, ok)
			continue
		}

		// capture output in a string buffer, which we will compare to a saved result
		buf := bytes.NewBufferString("")

		// render twice: once hidden, once revealed
		err := b.ConsoleRender(buf)
		if err != nil {
			t.Errorf("Error during ConsoleRender for game type %q: %s", bt.difficulty, err)
		}
		fmt.Fprintln(buf)

		b.RevealAll()
		err = b.ConsoleRender(buf)
		if err != nil {
			t.Errorf("Error during ConsoleRender for game type %q: %s", bt.difficulty, err)
		}

		// Now compare the render againsgt the expected output
		testfilename := fmt.Sprintf("testdata/render.%s.out", bt.difficulty)
		testdata, err := ioutil.ReadFile(testfilename)
		if err != nil {
			// ignore errors around reading test case data
			fmt.Fprintf(os.Stderr, "Could not read Render test data file %q, skipping render comparison", testfilename)
			continue
		}
		if string(testdata) != string(buf.Bytes()) {
			t.Errorf("Render test comparison failure.  Expected:\n%q\n\n Got:\n%q\n", string(testdata), string(buf.Bytes()))
		}
	}

}

func TestBoardClick(t *testing.T) {
}

func TestBoardWinCondition(t *testing.T) {

}
