package ecs

import "codeberg.org/alluneedistux/GoJumpRunShoot/statemachine"

type TransformComponentData struct {
	Posx          float64
	Posy          float64
	FlipImg       bool
	Hspeed        float64
	Vspeed        float64
	IsJumping     bool
	IsNotMoving   bool
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

		entiyHasDynamicComponent := ecsManager.HasNamedComponent(components, "DYNAMIC_COMPONENT")
		entiyHasTransformComponent := ecsManager.HasNamedComponent(components, "TRANSFORM_COMPONENT")

		if ! entiyHasDynamicComponent || ! entiyHasTransformComponent {
			continue
		}

		pTCD := sys.GetComponentData(entityID).(*TransformComponentData)

		sys.UpdateComponent(delta, pTCD)
	}
}

func (sys *TransformSystem) UpdateComponent(delta float64, essentialData ...interface{}) {
	pTCD := essentialData[0].(*TransformComponentData)
	pTCD.Posx += pTCD.Hspeed * delta
	pTCD.Posy -= pTCD.Vspeed
}