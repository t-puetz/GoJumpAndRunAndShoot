package ecs

import (
	"codeberg.org/alluneedistux/GoJumpRunShoot/statemachine"
	"github.com/veandco/go-sdl2/sdl"
)

type CollisionCoreData struct {
	CollisionDirection  string
	IntersectRect       *sdl.Rect
	EnitytCollidingWith uint64
}

type CollisionComponentData struct {
    *CollisionCoreData
}

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
			pCCD1 := sys.GetComponentData(ent1).(*CollisionComponentData)
			pCCD2 := sys.GetComponentData(ent2).(*CollisionComponentData)

			collisionDirections := sys.CollideSystemCoreDetectAABB(sys.ECSManager, ent1, ent2, pTCD1, pTCD2, pCCD1, pCCD2)

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

			sys.UpdateComponent(delta, pTCD1, pTCD2, entityOneHasDynamicComponent, entityTwoHasDynamicComponent,
				collisionDirections, ent1, ent2, pCCD1, pCCD2)
		}
	}
}

func (sys *CollideSystem) CollideSystemCoreDetectAABB(ecsManager *ECSManager, ent1, ent2 uint64, pTCD1, pTCD2 *TransformComponentData, pCCD1, pCCD2 * CollisionComponentData) map[string]bool {
	imgRectOne := ecsManager.GetEntityRect(ent1)
	imgRectTwo := ecsManager.GetEntityRect(ent2)

	collisionDirections := make(map[string]bool)
	collideRect, areColliding := imgRectOne.Intersect(imgRectTwo)

	if !areColliding {
		return nil
	}

	pCCD1.EnitytCollidingWith = ent2
	pCCD2.EnitytCollidingWith = ent1
	pCCD1.IntersectRect = &collideRect
	pCCD2.IntersectRect = &collideRect

	// Ent1 comes from left and hits left side of ent2. Facing does not matter.
	if imgRectOne.X+imgRectOne.W > imgRectTwo.X && (pTCD1.Hspeed > 0 || pTCD2.Hspeed < 0) {
		pCCD1.CollisionDirection = "right"
		pCCD2.CollisionDirection = "left"
	}

	// Ent1 comes from right and hits right side of ent2. Facing does not matter.
	if imgRectOne.X < imgRectTwo.X+imgRectTwo.W && (pTCD1.Hspeed < 0 || pTCD2.Hspeed > 0) {
		pCCD1.CollisionDirection = "left"
		pCCD2.CollisionDirection = "right"
	}

	// Ent1's top edge hits ent2's bottom edge. (e.g. Head hits bottom)
	if imgRectOne.Y < imgRectTwo.Y+imgRectTwo.H && (pTCD1.Vspeed > 0 || pTCD2.Vspeed < 0) {
		pCCD1.CollisionDirection = "top"
		pCCD2.CollisionDirection = "bottom"
	}

	// Ent1's bottom edge hits ent2's top edge. (e.g. Feet hit ground)
	if imgRectOne.Y+imgRectOne.H > imgRectTwo.Y && (pTCD1.Vspeed < 0 || pTCD2.Vspeed > 0) {
		pCCD1.CollisionDirection = "bottom"
		pCCD2.CollisionDirection = "top"
	}

	return collisionDirections
}

func (sys *CollideSystem) UpdateComponent(delta float64, essentialData...interface{}) {
	pTCD1 := essentialData[0].(*TransformComponentData)
	pTCD2 := essentialData[1].(*TransformComponentData)

	entityOneHasDynamicComponent := essentialData[2].(bool)
	entityTwoHasDynamicComponent := essentialData[3].(bool)
	//collisionDirections := essentialData[4].(map[string]bool)
	//ent1 := essentialData[5].(uint64)
	//ent2 := essentialData[6].(uint64)

	pCCD1 := essentialData[7].(*CollisionComponentData)
	pCCD2 := essentialData[8].(*CollisionComponentData)


	//imgRectOne := sys.ECSManager.GetEntityRect(ent1)
	//imgRectTwo := sys.ECSManager.GetEntityRect(ent2)

	if entityOneHasDynamicComponent {
		if pCCD1.CollisionDirection == "right" {
			pTCD1.Posx -= float64(pCCD1.IntersectRect.W)
			pTCD1.IsNotMoving = true
			pTCD1.IsJumping = false
		} else if pCCD1.CollisionDirection == "left" {
			pTCD1.Posx += float64(pCCD1.IntersectRect.W)
			pTCD1.IsNotMoving = true
			pTCD1.IsJumping = false
		}

		if pCCD1.CollisionDirection == "bottom" {
			pTCD1.Posy -= float64(pCCD1.IntersectRect.H)
			pTCD1.Vspeed = 0
			pTCD1.IsNotMoving = true
			pTCD1.IsJumping = false
		} else if pCCD1.CollisionDirection == "top" {
			pTCD1.Posy += float64(pCCD1.IntersectRect.H)
			pTCD1.IsJumping = true
			pTCD1.Vspeed = 0
		}
	}

	if entityTwoHasDynamicComponent {
		if pCCD2.CollisionDirection == "right" {
			pTCD2.Posx += float64(pCCD2.IntersectRect.W)
			pTCD2.IsNotMoving = true
			pTCD2.IsJumping = false
		} else if pCCD2.CollisionDirection == "left"  {
			pTCD2.Posx -= float64(pCCD2.IntersectRect.W)
			pTCD2.IsNotMoving = true
			pTCD2.IsJumping = false
		}

		if pCCD2.CollisionDirection == "bottom" {
			pTCD2.Posy += float64(pCCD2.IntersectRect.H)
			pTCD2.Vspeed = 0
			pTCD2.IsNotMoving = true
			pTCD2.IsJumping = false
		} else if pCCD2.CollisionDirection == "top" {
			pTCD2.Posy -= float64(pCCD2.IntersectRect.H)
			pTCD2.IsJumping = true
			pTCD2.Vspeed = 0
		}
	}
}