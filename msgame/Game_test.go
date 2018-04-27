package msgame

import (
	"os"
	"testing"
)

func TestRecordedGame(t *testing.T) {
	game := New(1995)

	gamefile := "testgame.easy.txt"
	infile, err := os.Open(gamefile)
	if infile == nil {
		t.Errorf("Failed to read game test script %q : %s", gamefile, err)
	}

	err = game.RunConsole(infile, os.Stdout)
}
