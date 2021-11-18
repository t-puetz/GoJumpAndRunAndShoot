package ecs

import (
	"codeberg.org/alluneedistux/GoJumpRunShoot/input"
	"codeberg.org/alluneedistux/GoJumpRunShoot/statemachine"
	"github.com/veandco/go-sdl2/sdl"
)

type ActiveControlComponentData struct {
}

type ActiveControlSystem struct {
	*CommonSystemData
	Keyboard         *input.Keyboard
}

func NewActiveControlSystem(e *ECSManager, k *input.Keyboard) *ActiveControlSystem {
	return &ActiveControlSystem{
		CommonSystemData: NewCommonSystemData("ACTIVE_CONTROL_COMPONENT", e),
		Keyboard:         k,
	}
}

func (sys *ActiveControlSystem) Run(delta float64, statemachine *statemachine.StateMachine) {
	ecsManager := sys.ECSManager
	entityToComponentMapOrdered := *ecsManager.EntityToComponentMapOrdered

	for el := entityToComponentMapOrdered.Front(); el != nil; el = el.Next() {
		components := el.Value.([]uint16)
		entityID := el.Key.(uint64)

		if !ecsManager.HasNamedComponent(components, "ACTIVE_CONTROL_COMPONENT") {
			continue
		}

		pTCD := sys.ECSManager.GetComponentDataByName(entityID, "TRANSFORM_COMPONENT")
		sliceWithComponentData := make([]interface{}, 2, 2)
		sliceWithComponentData[0] = pTCD

		sliceOtherParametersUpdateComponent := make([]interface{}, 1, 1)
		sliceOtherParametersUpdateComponent[0] = statemachine

		sys.UpdateComponent(delta, sliceWithComponentData, sliceOtherParametersUpdateComponent)
	}
}

func (sys *ActiveControlSystem) UpdateComponent(delta float64, sliceWithComponentData []interface{}, sliceOtherParametersUpdateComponent []interface{}) {
	pTCD := sliceWithComponentData[0].(*TransformComponentData)
	sm := sliceOtherParametersUpdateComponent[0].(*statemachine.StateMachine)

	if sm.CurrentState == statemachine.WELCOME_SCREEN {
		if sys.Keyboard.KeyHeldDown(sdl.Keycode('s')) {
			sm.DoTransition(statemachine.WELCOME_SCREEN, statemachine.GAME)
		}

		if sys.Keyboard.KeyHeldDown(sdl.Keycode('o')) {
			sm.DoTransition(statemachine.WELCOME_SCREEN, statemachine.OPTIONS_MENU)
		}

		if sys.Keyboard.KeyHeldDown(sdl.Keycode('e')) {
			sm.DoTransition(statemachine.WELCOME_SCREEN, statemachine.EXIT)
		}
	}

	if sm.CurrentState == statemachine.GAME {
		if sys.Keyboard.KeyHeldDown(sdl.Keycode('a')) {
			pTCD.FlipImg = true
			pTCD.Hspeed = -5.0
		}

		if sys.Keyboard.KeyHeldDown(sdl.Keycode('d')) {
			pTCD.FlipImg = false
			pTCD.Hspeed = 5.0
		}

		if sys.Keyboard.KeyJustPressed(sdl.Keycode(' ')) && !pTCD.IsJumping {
			pTCD.IsJumping = true
			pTCD.Vspeed = 31.0
		}

		playerStoppedMoving := !sys.Keyboard.KeyHeldDown(sdl.Keycode('d')) && !sys.Keyboard.KeyHeldDown(sdl.Keycode('a'))

		if playerStoppedMoving {
			pTCD.Hspeed = 0
		}
	}
}
