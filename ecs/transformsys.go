package ecs

import (
	"github.com/t-puetz/GoJumpAndRunAndShoot/statemachine"
	"math"
)

type TransformComponentData struct {
	LastPosX    int32
	LastPosY    int32
	LastSpeed   int32
	PosX        int32
	PosY        int32
	DX          int32
	DY          int32
	FlipImg     bool
	Hspeed      int32
	Vspeed      int32
	IsJumping   bool
	IsNotMoving bool
}

type TransformSystem struct {
	*CommonSystemData
}

func NewTransformSystem(e *ECSManager) *TransformSystem {
	return &TransformSystem{
		CommonSystemData: NewCommonSystemData("TRANSFORM_COMPONENT", e),
	}
}

func (sys *TransformSystem) Run(delta float64, statemachine *statemachine.StateMachine) {
	ecsManager := sys.ECSManager
	entityToComponentMapOrdered := ecsManager.EntityToComponentMap

	for el := entityToComponentMapOrdered.Front(); el != nil; el = el.Next() {
		entityID := el.Key.(uint64)
		components := el.Value.([]uint16)

		entityHasDynamicComponent := ecsManager.HasNamedComponent(components, "DYNAMIC_COMPONENT")
		entityHasTransformComponent := ecsManager.HasNamedComponent(components, "TRANSFORM_COMPONENT")

		if !entityHasDynamicComponent || !entityHasTransformComponent {
			continue
		}

		pTCD := sys.GetComponentData(entityID).(*TransformComponentData)

		sys.UpdateComponent(delta, pTCD)
	}
}

func (sys *TransformSystem) UpdateComponent(delta float64, essentialData ...interface{}) {
	pTCD := essentialData[0].(*TransformComponentData)
	pTCD.LastPosX = pTCD.PosX
	pTCD.LastPosY = pTCD.PosY
	pTCD.LastSpeed = pTCD.Hspeed
	pTCD.PosX += int32(float64(pTCD.Hspeed) * delta)
	pTCD.PosY -= pTCD.Vspeed
	pTCD.DX = int32(math.Abs(float64(pTCD.LastPosX - pTCD.PosX)))
	pTCD.DY = int32(math.Abs(float64(pTCD.LastPosY - pTCD.PosY)))
}
