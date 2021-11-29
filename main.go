package main

import (
	"github.com/t-puetz/GoJumpAndRunAndShoot/game"
	"github.com/t-puetz/GoJumpAndRunAndShoot/input"
	_ "image/png"
)

func main() {
	g := &game.Game{}

	g.Keyboard = input.NewKeyboard()
	g.InitializeSDL()
	g.PrepareBasicGameData()

	g.LoadWelcomeScreen()
	g.RunWelcomeScreen()

	g.LoadFirstLevel()
	g.Run()
}
