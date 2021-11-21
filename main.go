package main

import (
	"codeberg.org/alluneedistux/GoJumpRunShoot/game"
	"codeberg.org/alluneedistux/GoJumpRunShoot/input"
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
