package ecs

import "codeberg.org/alluneedistux/GoJumpRunShoot/statemachine"

type GravityComponentData struct {}

type GravitySystem struct {
	*CommonSystemData
}

func NewGravitySystem(e *ECSManager) *GravitySystem {
	return &GravitySystem{
		CommonSystemData: NewCommonSystemData("GRAVITY_COMPONENT", e),
	}
}

func (sys *GravitySystem) Run(delta float64, statemachine *statemachine.StateMachine) {
	const GRAVITY = 0.981
	ecsManager := sys.ECSManager
	entityToComponentMapOrdered := ecsManager.EntityToComponentMap

	for el := entityToComponentMapOrdered.Front(); el != nil; el = el.Next() {
		components := el.Value.([]uint16)
		entityID := el.Key.(uint64)

		entityHasGravityComponent := ecsManager.HasNamedComponent(components, "GRAVITY_COMPONENT")
		entityHasTransformComponent := ecsManager.HasNamedComponent(components, "TRANSFORM_COMPONENT")

		if ! entityHasGravityComponent || ! entityHasTransformComponent {
			continue
		}

		pGCD := sys.GetComponentData(entityID).(*GravityComponentData)
		pTCD := sys.ECSManager.GetComponentDataByName(entityID, "TRANSFORM_COMPONENT").(*TransformComponentData)

		sliceWithComponentData := make([]interface{}, 2, 2)
		sliceWithComponentData[0] = pGCD
		sliceWithComponentData[1] = pTCD

		sliceOtherParametersUpdateComponent := make([]interface{}, 1, 1)
		sliceOtherParametersUpdateComponent[0] = GRAVITY

		sys.UpdateComponent(delta, sliceWithComponentData, sliceOtherParametersUpdateComponent)
	}
}

func (sys *GravitySystem) UpdateComponent(delta float64, sliceWithComponentData []interface{}, sliceOtherParametersUpdateComponent []interface{}) {
    pTCD := sliceWithComponentData[1].(*TransformComponentData)
	GRAVITY := sliceOtherParametersUpdateComponent[0].(float64)
	pTCD.Vspeed -= GRAVITY
}
