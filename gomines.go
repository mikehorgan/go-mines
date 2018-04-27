package main

import (
	"go-mines/msgame"
	"os"
	"time"
)

func main() {
	game := msgame.New(time.Now().UnixNano())

	game.RunConsole(os.Stdin, os.Stdout)
}
