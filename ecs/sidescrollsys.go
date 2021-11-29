package ecs

import (
	"github.com/t-puetz/GoJumpAndRunAndShoot/statemachine"
	"log"
)

type SideScrollComponentData struct {
	sidescrolled *float64
	hspeed float64
}

type SideScrollSystem struct {
	*CommonSystemData
}

func NewSideScrollSystem(e *ECSManager) *SideScrollSystem {
	return &SideScrollSystem{
		CommonSystemData: NewCommonSystemData("SIDE_SCROLL_COMPONENT", e),
	}
}

func (sys *SideScrollSystem) Run(delta float64, statemachine *statemachine.StateMachine) {
	sys.UpdateComponent(delta, "Hello")
}

func (sys *SideScrollSystem) UpdateComponent(delta float64, essentialData ...interface{}) {
	ecsManager := sys.ECSManager
	entityToComponentMapOrdered := ecsManager.EntityToComponentMap

	for el := entityToComponentMapOrdered.Front(); el != nil; el = el.Next() {
		entityID := el.Key.(uint64)
		components := el.Value.([]uint16)

		entityHasSideScrollComponent := ecsManager.HasNamedComponent(components, "SIDE_SCROLL_COMPONENT")

		playersTransformComponentData := sys.ECSManager.GetComponentDataByName(1, "TRANSFORM_COMPONENT").(*TransformComponentData)
		log.Println(playersTransformComponentData.PosX)

		pTCD := sys.ECSManager.GetComponentDataByName(entityID, "TRANSFORM_COMPONENT").(*TransformComponentData)
		pSCD := sys.GetComponentData(entityID).(*SideScrollComponentData)

		if !entityHasSideScrollComponent {
			// Only the player should be skipped here (Entity 1 in normal levels)
			playersTransformComponentData = sys.ECSManager.GetComponentDataByName(entityID, "TRANSFORM_COMPONENT").(*TransformComponentData)
			continue
		}

		if playersTransformComponentData.PosX > 450 && !playersTransformComponentData.IsNotMoving {
			pTCD.PosX -= pSCD.hspeed * delta
		}
	}
}
