package ecs

import (
	"codeberg.org/alluneedistux/GoJumpRunShoot/statemachine"
	"github.com/veandco/go-sdl2/sdl"
	"log"
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
	LastAnimation            string
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

		animationName := ""

		if sys.ECSManager.HasNamedComponent(components, "ACTIVE_CONTROL_COMPONENT") {
			// Check for Player

			if pTCD.IsNotMoving {
				animationName = "Idle"
			} else {
				animationName = "Walk"
			}

			if pTCD.IsJumping {
				animationName = "Jump"
			}
		}

		animationChanged := false

		if pACD.LastAnimation != animationName {
			animationChanged = true
		}

		pACD.LastAnimation = animationName

		pACDCore := animationTypeMap[animationName]

		sys.UpdateComponent(delta, pRCD, pACDCore, animationChanged)
	}
}

func (sys *AnimateSystem) UpdateComponent(delta float64, essentialData ...interface{}) {
	pRCD := essentialData[0].(*RenderComponentData)
	pACDCore := essentialData[1].(*AnimationComponentDataCore)
	pACDCore.CurrentFrame++
	animationChanged := essentialData[2].(bool)

	if animationChanged {
		pACDCore.CurrentFrame = 0
	}

	timeForNextImage := pACDCore.CurrentFrame%pACDCore.DefaultAnimationDuration == 0
	moreThanOneImage := pACDCore.NumberAnimations > 1


	if ! timeForNextImage || ! moreThanOneImage {
        return
	}

	for i, image := range pACDCore.Images {
		log.Println(i, image, pRCD.Image, pRCD.Path, pACDCore.CurrentFrame, pACDCore.DefaultAnimationDuration)
		var nextIndex int

		if i < len(pACDCore.Images) - 1 {
			nextIndex = i + 1
		} else if i == len(pACDCore.Images) - 1 {
			nextIndex = 0
		}

		pRCD.Image = pACDCore.Images[nextIndex]
		pRCD.Texture = pACDCore.Textures[nextIndex]
		pRCD.Path = pACDCore.Paths[nextIndex]

		break

	}
}
