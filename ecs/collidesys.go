package ecs

import "codeberg.org/alluneedistux/GoJumpRunShoot/statemachine"

type CollideSystem struct {
	*CommonSystemData
}

func NewCollideSystem(e *ECSManager) *CollideSystem {
	return &CollideSystem{
		CommonSystemData: NewCommonSystemData("COLLIDE_COMPONENT", e),
	}
}

func (sys *CollideSystem) Run(delta float64, statemachine *statemachine.StateMachine) {
	ecsManager := sys.ECSManager
	entityToComponentMapOrdered := ecsManager.EntityToComponentMap

	for el := entityToComponentMapOrdered.Front(); el != nil; el = el.Next() {
	    components := el.Value.([]uint16)
	    entityID := el.Key.(uint64)

		if ! sys.ECSManager.HasNamedComponent(components, "DYNAMIC_COMPONENT") {
			continue
		}

		entityOneHasDynamicComponent := sys.ECSManager.HasNamedComponent(components, "DYNAMIC_COMPONENT")
		lengthOfEntityComponentMap := uint64(entityToComponentMapOrdered.Len())

		for j := entityID + 1; j < lengthOfEntityComponentMap; j++ {
			componentsEntityTwo, _ := entityToComponentMapOrdered.Get(j)

			if ! sys.ECSManager.HasNamedComponent(componentsEntityTwo.([]uint16),"COLLIDE_COMPONENT") {
				continue
			}

			entityTwoHasDynamicComponent := sys.ECSManager.HasNamedComponent(componentsEntityTwo.([]uint16), "DYNAMIC_COMPONENT")

			ent1 := entityID
			ent2 := j

			if (!entityOneHasDynamicComponent) && (!entityTwoHasDynamicComponent) {
				continue
			}

			pTCD1 := sys.ECSManager.GetComponentDataByName(ent1, "TRANSFORM_COMPONENT").(*TransformComponentData)
			pTCD2 := sys.ECSManager.GetComponentDataByName(ent2, "TRANSFORM_COMPONENT").(*TransformComponentData)

			collisionDirections := sys.CollideSystemCoreDetectAABB(sys.ECSManager, ent1, ent2, pTCD1, pTCD2)

			if collisionDirections == nil {
				continue
			}

			sliceWithComponentData := make([]interface{}, 2, 2)
			sliceWithComponentData[0] = pTCD1
			sliceWithComponentData[1] = pTCD2

			sliceOtherParametersUpdateComponent := make([]interface{}, 5, 5)
			sliceOtherParametersUpdateComponent[0] = entityOneHasDynamicComponent
			sliceOtherParametersUpdateComponent[1] = entityTwoHasDynamicComponent
			sliceOtherParametersUpdateComponent[2] = collisionDirections
			sliceOtherParametersUpdateComponent[3] = ent1
			sliceOtherParametersUpdateComponent[4] = ent2

			sys.UpdateComponent(delta, sliceWithComponentData, sliceOtherParametersUpdateComponent)
		}
	}
}

func (sys *CollideSystem) CollideSystemCoreDetectAABB(ecsManager *ECSManager, ent1, ent2 uint64, pTCD1, pTCD2 *TransformComponentData) map[string]bool {
	imgRectOne := ecsManager.GetEntityRect(ent1)
	imgRectTwo := ecsManager.GetEntityRect(ent2)

	collisionDirections := make(map[string]bool)

	// Formula for AABB collision detection:
	areColliding := (imgRectOne.X < imgRectTwo.X+imgRectTwo.W) &&
		(imgRectOne.X+imgRectOne.W > imgRectTwo.X) &&
		(imgRectOne.Y < imgRectTwo.Y+imgRectTwo.H) &&
		(imgRectOne.Y+imgRectOne.H > imgRectTwo.Y)

	if !areColliding {
		return nil
	}

	// Ent1 comes from left and hits left side of ent2. Facing does not matter.
	collisionDirections["right"] = imgRectOne.X+imgRectOne.W > imgRectTwo.X && (pTCD1.Hspeed > 0 || pTCD2.Hspeed < 0)

	// Ent1 comes from right and hits right side of ent2. Facing does not matter.
	collisionDirections["left"] = imgRectOne.X < imgRectTwo.X+imgRectTwo.W && (pTCD1.Hspeed < 0 || pTCD2.Hspeed > 0)

	// Ent1's top edge hits ent2's bottom edge. (e.g. Head hits bottom)
	collisionDirections["top"] = imgRectOne.Y < imgRectTwo.Y+imgRectTwo.H && (pTCD1.Vspeed > 0 || pTCD2.Vspeed < 0)

	// Ent1's bottom edge hits ent2's top edge. (e.g. Feet hit ground)
	collisionDirections["bottom"] = imgRectOne.Y+imgRectOne.H > imgRectTwo.Y && (pTCD1.Vspeed < 0 || pTCD2.Vspeed > 0)

	return collisionDirections
}

func (sys *CollideSystem) UpdateComponent(delta float64, sliceWithComponentData []interface{}, sliceOtherParametersUpdateComponent []interface{}) {
	pTCD1 := sliceWithComponentData[0].(*TransformComponentData)
	pTCD2 := sliceWithComponentData[1].(*TransformComponentData)

	entityOneHasDynamicComponent := sliceOtherParametersUpdateComponent[0].(bool)
	entityTwoHasDynamicComponent := sliceOtherParametersUpdateComponent[1].(bool)
	collisionDirections := sliceOtherParametersUpdateComponent[2].(map[string]bool)
	ent1 := sliceOtherParametersUpdateComponent[3].(uint64)
	ent2 := sliceOtherParametersUpdateComponent[4].(uint64)

	imgRectOne := sys.ECSManager.GetEntityRect(ent1)
	imgRectTwo := sys.ECSManager.GetEntityRect(ent2)

	if entityOneHasDynamicComponent {
		if collisionDirections["right"] {
			pTCD1.Posx -= float64(imgRectOne.X + imgRectOne.W - imgRectTwo.X)
		} else if collisionDirections["left"] {
			pTCD1.Posx += float64(imgRectTwo.X + imgRectTwo.W - imgRectOne.X)
		}

		if collisionDirections["bottom"] {
			pTCD1.Posy -= float64(imgRectOne.Y + imgRectOne.H - imgRectTwo.Y)
			pTCD1.IsJumping = false
			pTCD1.Vspeed = 0
		} else if collisionDirections["top"] {
			pTCD1.Posy += float64(imgRectTwo.Y + imgRectTwo.H - imgRectOne.Y)
			pTCD1.IsJumping = false
			pTCD1.Vspeed = 0
		}
	}

	if entityTwoHasDynamicComponent {
		if collisionDirections["right"] {
			pTCD2.Posx += float64(imgRectOne.X + imgRectOne.W - imgRectTwo.X)
		} else if collisionDirections["left"] {
			pTCD2.Posx -= float64(imgRectTwo.X + imgRectTwo.W - imgRectOne.X)
		}

		if collisionDirections["bottom"] {
			pTCD2.Posy += float64(imgRectOne.Y + imgRectOne.H - imgRectTwo.Y)
			pTCD2.IsJumping = false
			pTCD2.Vspeed = 0
		} else if collisionDirections["top"] {
			pTCD2.Posy -= float64(imgRectTwo.Y + imgRectTwo.H - imgRectOne.Y)
			pTCD2.IsJumping = false
			pTCD2.Vspeed = 0
		}
	}
}