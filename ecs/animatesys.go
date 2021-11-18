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

		if !sys.ECSManager.HasNamedComponent(components, "TRANSFORM_COMPONENT") {
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
			animationName = "Front"
		}

		if pTCD.IsJumping {
			animationName = "Jump"
		}

		pACDCore := animationTypeMap[animationName]

		log.Println(animationName)

		noAnimationNeeded := ((len(pACDCore.Images) <= 1 || len(pACDCore.Textures) <= 1) ||
			(pACDCore.Images == nil || pACDCore.Textures == nil || pACDCore.Paths == nil))

		if noAnimationNeeded {
			continue
		}

		sliceWithComponentData := make([]interface{}, 2, 2)
		sliceWithComponentData[0] = pRCD
		sliceWithComponentData[1] = pACDCore

		sliceOtherParametersUpdateComponent := make([]interface{}, 1, 1)
		sliceOtherParametersUpdateComponent[0] = entityID

		sys.UpdateComponent(delta, sliceWithComponentData, sliceOtherParametersUpdateComponent)

	}
}

func (sys *AnimateSystem) UpdateComponent(delta float64, sliceWithComponentData []interface{}, sliceOtherParametersUpdateComponent []interface{}) {
	//pRCD := sliceWithComponentData[0].(*RenderComponentData)
	//pACDCore := sliceWithComponentData[2].(*AnimationComponentDataCore)

	//entityID := sliceOtherParametersUpdateComponent[0].(uint64)

	//animationDuration := pACDCore.AnimationDuration
	//currentFrame := pACDCore.CurrentFrame


}
