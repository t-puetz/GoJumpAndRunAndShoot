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
	AnimationDuration uint8
	CurrentFrame      uint8
	Paths             []string
	Images            []*sdl.Surface
	Textures          []*sdl.Texture
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
	entityToComponentMapOrdered := *ecsManager.EntityToComponentMapOrdered

	for el := entityToComponentMapOrdered.Front(); el != nil; el = el.Next() {
		components := el.Value.([]uint16)
		entityID := el.Key.(uint64)

		if !ecsManager.HasComponent(components, sys.SystemID) {
			continue
		}

		pACD := sys.GetComponentData(entityID).(*AnimateComponentData)

		animationTypeMap := *pACD.AnimationData

		for animationName, _ := range animationTypeMap {
			pACDCore := animationTypeMap[animationName]

			noAnimationNeeded := ((len(pACDCore.Images) <= 1 || len(pACDCore.Textures) <= 1) ||
				(pACDCore.Images == nil || pACDCore.Textures == nil || pACDCore.Paths == nil))

			if noAnimationNeeded {
				continue
			}

			sliceWithComponentData := make([]interface{}, 2, 2)
			sliceWithComponentData[0] = nil

			sliceOtherParametersUpdateComponent := make([]interface{}, 4, 4)
			sliceOtherParametersUpdateComponent[0] = animationName
			sliceOtherParametersUpdateComponent[1] = animationTypeMap
			sliceOtherParametersUpdateComponent[2] = entityID

			sys.UpdateComponent(delta, sliceWithComponentData, sliceOtherParametersUpdateComponent)
		}
	}
}

/*
func (sys *AnimateSystem) RunUnordered(delta float64) {
	ecsManager := sys.ECSManager
	entityToComponentMap := *ecsManager.EntityToComponentMap

	for entityID, components := range entityToComponentMap {

		if !ecsManager.HasComponent(components, sys.SystemID) {
			continue
		}

		pACD := sys.GetComponentData(entityID).(*AnimateComponentData)

		animationTypeMap := *pACD.AnimationData

		for animationName, _ := range animationTypeMap {
			pACDCore := animationTypeMap[animationName]

			noAnimationNeeded := ((len(pACDCore.Images) <= 1 || len(pACDCore.Textures) <= 1) ||
				(pACDCore.Images == nil || pACDCore.Textures == nil || pACDCore.Paths == nil))

			if noAnimationNeeded {
				continue
			}

			sliceWithComponentData := make([]interface{}, 2, 2)
			sliceWithComponentData[0] = nil

			sliceOtherParametersUpdateComponent := make([]interface{}, 4, 4)
			sliceOtherParametersUpdateComponent[0] = animationName
			sliceOtherParametersUpdateComponent[1] = animationTypeMap
			sliceOtherParametersUpdateComponent[2] = entityID

			sys.UpdateComponent(delta, sliceWithComponentData, sliceOtherParametersUpdateComponent)
		}
	}
}
*/


func (sys *AnimateSystem) UpdateComponent(delta float64, sliceWithComponentData []interface{}, sliceOtherParametersUpdateComponent []interface{}) {
	animationName := sliceOtherParametersUpdateComponent[0].(string)
	animationTypeMap := sliceOtherParametersUpdateComponent[1].(map[string]*AnimationComponentDataCore)
	entityID := sliceOtherParametersUpdateComponent[2].(uint64)
	pACDCore := animationTypeMap[animationName]

	if pACDCore.AnimationDuration > 0 && pACDCore.CurrentFrame%pACDCore.AnimationDuration == 0 {
		// Get the RenderComponentData to find out what the current drawn image is
		// and set it to the next image needed for a cyclic animation for the next
		// rendering iteration to present the correct animation image to the screen
		pRCD := sys.GetComponentData(entityID).(*RenderComponentData)
		pTCD := sys.GetComponentData(entityID).(*TransformComponentData)

		entityIsMoving := pTCD.Hspeed != 0

		for k, _ := range pACDCore.Images {
			if pACDCore.Images[k] == nil || pACDCore.Paths[k] == "" {
				// We are at an index where there is no image assigned
				continue
			}

			// TODO: Take care of ALL animation types

			if pRCD.Img == pACDCore.Images[entityID] && entityIsMoving {
				if k < len(pACDCore.Images)-1 && pACDCore.Images[k+1] != nil {
					pRCD.Img = pACDCore.Images[k+1]
					pRCD.Texture = pACDCore.Textures[k+1]
				} else if k == len(pACDCore.Images)-1 || pACDCore.Images[k+1] == nil {
					pRCD.Img = pACDCore.Images[0]
					pRCD.Texture = pACDCore.Textures[0]
				}
			}
			pACDCore.CurrentFrame++
		}
	}
}
