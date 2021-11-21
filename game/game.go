package game

import (
	"codeberg.org/alluneedistux/GoJumpRunShoot/ecs"
	"codeberg.org/alluneedistux/GoJumpRunShoot/input"
	"codeberg.org/alluneedistux/GoJumpRunShoot/statemachine"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"log"
	"os"
	"time"
)

type Game struct {
	Window            *sdl.Window
	Surface           *sdl.Surface
	Renderer          *sdl.Renderer
	Keyboard          *input.Keyboard
	ECSManager        *ecs.ECSManager
	AssetDescriptions *map[string]*AssetJSONConfig
	LvlDescription    *LevelJSONConfig
	StateMachine      *statemachine.StateMachine
}

func (g *Game) PrepareBasicGameData() {
	g.ECSManager = ecs.NewECSManager()

	g.ECSManager.Systems[0] = ecs.NewActiveControlSystem(g.ECSManager, g.Keyboard)
	g.ECSManager.Systems[1] = ecs.NewGravitySystem(g.ECSManager)
	g.ECSManager.Systems[2] = ecs.NewTransformSystem(g.ECSManager)
	g.ECSManager.Systems[3] = ecs.NewCollideSystem(g.ECSManager)
	g.ECSManager.Systems[4] = ecs.NewAnimateSystem(g.ECSManager)
	g.ECSManager.Systems[5] = ecs.NewRenderSystem(g.ECSManager, g.Renderer)

	g.StateMachine = statemachine.NewStateMachine()
}

func (g *Game) InitializeSDL() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}

	window, winCreationErr := sdl.CreateWindow("Alone outside of Space :'(", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		1368, 720, sdl.WINDOW_SHOWN)
	if winCreationErr != nil {
		panic(winCreationErr)
	}

	renderer, rendererCreationErr := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)

	if rendererCreationErr != nil {
		log.Fatalf("Failed to create renderer: %s\n", rendererCreationErr)
	}

	surface, err := window.GetSurface()
	if err != nil {
		panic(err)
	}

	g.Window = window
	g.Surface = surface
	g.Renderer = renderer

	// Init ttf system for rendering text
	ttf.Init()

}

func (g *Game) LoadFirstLevel() {
	LoadAssetDescriptions(g)
	LoadLvlConfig(g, "./game/lvlone.json")
	InitializeLevel(g)
}

func (g *Game) LoadWelcomeScreen() {
	LoadAssetDescriptions(g)
	LoadLvlConfig(g, "./game/welcomescreen.json")
	InitializeLevel(g)
}

func (g *Game) RunSystems(delta float64) {
	for i, system := range g.ECSManager.Systems {
		if i != 5  {
			system.Run(delta, g.StateMachine)
		} else {
			// 5 is the index of the RenderSystem
			go system.Run(delta, g.StateMachine)
		}

	}
}

func (g *Game) RunWelcomeScreen() {
	runWelcomeScreen := true

	for runWelcomeScreen {
		// Is decided in ActiveControlSystem by Pressing S:
		if g.StateMachine.CurrentState == statemachine.GAME {
			runWelcomeScreen = false
			g.StateMachine.DoTransition(statemachine.WELCOME_SCREEN, statemachine.GAME)
		}

		g.runBasicQuitKeyboardEventLoop(runWelcomeScreen)

		g.ECSManager.Systems[0].Run(1.0, g.StateMachine)
		g.ECSManager.Systems[5].Run(1.0, g.StateMachine)

		sdl.Delay(30)
	}
}

func (g *Game) renderGamePausedText() {
	var font *ttf.Font
	var text *sdl.Surface

	font, _ = ttf.OpenFont("./assets/SourceCodePro-Bold.ttf", 32)
	text, _ = font.RenderUTF8Blended("GAME PAUSED", sdl.Color{R: 255, G: 0, B: 0, A: 255})
	_ = text.Blit(nil, g.Surface, &sdl.Rect{X: 684 - (text.W / 2), Y: 364 - (text.H / 2), W: 0, H: 0})
	_ = g.Window.UpdateSurface()
}

func (g *Game) runBasicQuitKeyboardEventLoop(running bool) {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch t := event.(type) {
		case *sdl.QuitEvent:
			_ = g.Window.Destroy()
			running = false
			sdl.Quit()
			os.Exit(0)
		case *sdl.KeyboardEvent:
			g.Keyboard.OnEvent(t)
		}
	}
}

func (g *Game) decideGameOrPauseState(delta float64) {
	switch g.StateMachine.CurrentState {
	case statemachine.PAUSE:
		log.Println("Game Paused")
		if g.Keyboard.KeyHeldDown(sdl.Keycode(27)) || g.Keyboard.KeyHeldDown(sdl.Keycode(1073741896)) {
			g.StateMachine.DoTransition(statemachine.PAUSE, statemachine.GAME)
			time.Sleep(time.Millisecond * 100)
		}

		g.renderGamePausedText()
	case statemachine.GAME:
		if g.Keyboard.KeyHeldDown(sdl.Keycode(1073741896)) {
			g.StateMachine.DoTransition(statemachine.GAME, statemachine.PAUSE)
			time.Sleep(time.Millisecond * 100)
		}
		g.RunSystems(delta)
	}
}

func (g *Game) Run() {
	running := true

	var now time.Time
	var elapsedTime time.Duration

	lastTime := time.Now()
	timePerFrame := 1000.0 / 60.0

	for running {
		now = time.Now()
		elapsedTime = now.Sub(lastTime)
		lastTime = now
		delta := float64(elapsedTime.Milliseconds()) / timePerFrame

		g.Keyboard.ResetChangedStates()
		g.runBasicQuitKeyboardEventLoop(running)
		g.decideGameOrPauseState(delta)

		sdl.Delay(10)
	}
}
