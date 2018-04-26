package main

import (
	"go-mines/msgame"
	"os"
)

func main() {
	game := msgame.New()

	game.RunConsole(os.Stdin, os.Stdout)
}
