package ecs

import (
	"codeberg.org/alluneedistux/GoJumpRunShoot/statemachine"
	"github.com/veandco/go-sdl2/sdl"
)

type AnimationComponentDataCore struct {
	// 16 Images should suffice
	// see ecs.go at data initialization
	// Probably not smart to hardcode this
	// Especially for ALL entities
	NumberAnimations         uint8
	DefaultAnimationDuration uint8
	CurrentFrame             uint8
	Paths                    []string
	Images                   []*sdl.Surface
	Textures                 []*sdl.Texture
}

type AnimateComponentData struct {
	// The string key is the animation's name such as "Idle", "Walk" etc...
	AnimationData *map[string]*AnimationComponentDataCore
}

type AnimateSystem struct {
	*CommonSystemData
}

func NewAnimateSystem(e *ECSManager) *AnimateSystem {
	return &AnimateSystem{
		CommonSystemData: NewCommonSystemData("ANIMATE_COMPONENT", e),
	}
}

func (sys *AnimateSystem) Run(delta float64, statemachine *statemachine.StateMachine) {
	ecsManager := sys.ECSManager
	entityToComponentMapOrdered := *ecsManager.EntityToComponentMap

	for el := entityToComponentMapOrdered.Front(); el != nil; el = el.Next() {
		components := el.Value.([]uint16)
		entityID := el.Key.(uint64)

		if !ecsManager.HasComponent(components, sys.SystemID) {
			continue
		}

		if !sys.ECSManager.HasNamedComponent(components, "RENDER_COMPONENT") {
			continue
		}

		pACD := sys.GetComponentData(entityID).(*AnimateComponentData)
		pRCD := sys.ECSManager.GetComponentDataByName(entityID, "RENDER_COMPONENT").(*RenderComponentData)
		pTCD := sys.ECSManager.GetComponentDataByName(entityID, "TRANSFORM_COMPONENT").(*TransformComponentData)

		animationTypeMap := *pACD.AnimationData

		var animationName string

		if pTCD.IsNotMoving {
			animationName = "Idle"
		} else {
			animationName = "Walk"
		}

		if pTCD.IsJumping {
			animationName = "Jump"
		}

		if !sys.ECSManager.HasNamedComponent(components, "TRANSFORM_COMPONENT") {
			animationName = "Fallback"
		}

		pACDCore := animationTypeMap[animationName]
		sliceWithComponentData := make([]interface{}, 2, 2)
		sliceWithComponentData[0] = pRCD
		sliceWithComponentData[1] = pACDCore

		sliceOtherParametersUpdateComponent := make([]interface{}, 1, 1)
		sliceOtherParametersUpdateComponent[0] = animationName

		sys.UpdateComponent(delta, sliceWithComponentData, sliceOtherParametersUpdateComponent)
	}
}

func (sys *AnimateSystem) UpdateComponent(delta float64, sliceWithComponentData []interface{}, sliceOtherParametersUpdateComponent []interface{}) {
	pRCD := sliceWithComponentData[0].(*RenderComponentData)
	pACDCore := sliceWithComponentData[1].(*AnimationComponentDataCore)


	timeForNextImage := pACDCore.CurrentFrame%pACDCore.DefaultAnimationDuration == 0
	moreThanOneImage :=  pACDCore.NumberAnimations > 1

	if ! moreThanOneImage {
		pRCD.Image = pACDCore.Images[0]
		pRCD.Texture = pACDCore.Textures[0]
		pRCD.Path = pACDCore.Paths[0]
	}

	for i := 0; i < int(pACDCore.NumberAnimations) && moreThanOneImage && timeForNextImage; i++ {
		pRCD.Image = pACDCore.Images[i]
		pRCD.Texture = pACDCore.Textures[i]
		pRCD.Path = pACDCore.Paths[i]

		pACDCore.CurrentFrame++
	}

	if pACDCore.CurrentFrame == 240 {
		pACDCore.CurrentFrame = 0
	}

}
