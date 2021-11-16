package main

import (
	"codeberg.org/alluneedistux/GoJumpRunShoot/game"
	"codeberg.org/alluneedistux/GoJumpRunShoot/input"
	"codeberg.org/alluneedistux/GoJumpRunShoot/statemachine"
	"github.com/veandco/go-sdl2/sdl"
	_ "image/png"
)

func main() {
	g := &game.Game{}
	sm := statemachine.NewStateMachine()

	kboard := input.NewKeyboard()
	g.Keyboard = kboard

	g.InitializeSDL()
	g.PrepareBasicGameData(sm)
	g.LoadWelcomeScreen()

	runWelcomeScreen := true

	g.RunWelcomeScreen(runWelcomeScreen)

	sm.DoTransition(statemachine.WELCOME_SCREEN, statemachine.GAME)

	g.LoadFirstLevel()

	running := true

	defer sdl.Quit()
	defer g.Window.Destroy()

	g.Run(running)

}
