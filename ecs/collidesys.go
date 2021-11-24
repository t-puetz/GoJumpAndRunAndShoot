package ecs

import (
	"codeberg.org/alluneedistux/GoJumpRunShoot/statemachine"
	"github.com/veandco/go-sdl2/sdl"
)

type CollisionCoreData struct {
	CollisionDirection  map[string]bool
	IntersectRect       *sdl.Rect
	EntityCollidingWith uint64
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

			areColliding := sys.CollideSystemCoreDetectAABB(sys.ECSManager, ent1, ent2, pTCD1, pTCD2, pCCD1, pCCD2)

			if !areColliding {
				continue
			}

			sys.UpdateComponent(delta, pTCD1, pTCD2, entityOneHasDynamicComponent, entityTwoHasDynamicComponent, pCCD1, pCCD2)
		}
	}
}

func (sys *CollideSystem) CollideSystemCoreDetectAABB(ecsManager *ECSManager, ent1, ent2 uint64, pTCD1, pTCD2 *TransformComponentData, pCCD1, pCCD2 * CollisionComponentData) bool {
	imgRectOne := ecsManager.GetEntityRect(ent1)
	imgRectTwo := ecsManager.GetEntityRect(ent2)

	collideRect, areColliding := imgRectOne.Intersect(imgRectTwo)

	if !areColliding {
		return false
	}

	pCCD1.EntityCollidingWith = ent2
	pCCD2.EntityCollidingWith = ent1
	pCCD1.IntersectRect = &collideRect
	pCCD2.IntersectRect = &collideRect

	// Ent1 comes from left and hits left side of ent2. Facing does not matter.
	pCCD1.CollisionDirection["right"] = imgRectOne.X+imgRectOne.W > imgRectTwo.X && (pTCD1.Hspeed > 0 || pTCD2.Hspeed < 0)

	// Ent1 comes from right and hits right side of ent2. Facing does not matter.
	pCCD1.CollisionDirection["left"] = imgRectOne.X < imgRectTwo.X+imgRectTwo.W && (pTCD1.Hspeed < 0 || pTCD2.Hspeed > 0)

	// Ent1's top edge hits ent2's bottom edge. (e.g. Head hits bottom)
	pCCD1.CollisionDirection["top"] = imgRectOne.Y < imgRectTwo.Y+imgRectTwo.H && (pTCD1.Vspeed > 0 || pTCD2.Vspeed < 0)

	// Ent1's bottom edge hits ent2's top edge. (e.g. Feet hit ground)
	pCCD1.CollisionDirection["bottom"] = imgRectOne.Y+imgRectOne.H > imgRectTwo.Y && (pTCD1.Vspeed < 0 || pTCD2.Vspeed > 0)

	return true
}

func (sys *CollideSystem) UpdateComponent(delta float64, essentialData...interface{}) {
	pTCD1 := essentialData[0].(*TransformComponentData)
	pTCD2 := essentialData[1].(*TransformComponentData)

	entityOneHasDynamicComponent := essentialData[2].(bool)
	entityTwoHasDynamicComponent := essentialData[3].(bool)

	pCCD1 := essentialData[4].(*CollisionComponentData)
	pCCD2 := essentialData[5].(*CollisionComponentData)

	if entityOneHasDynamicComponent {
		if pCCD1.CollisionDirection["right"] {
			pTCD1.Posx -= float64(pCCD1.IntersectRect.W)
			pTCD1.IsNotMoving = true
			pTCD1.IsJumping = false
		} else if pCCD1.CollisionDirection["left"] {
			pTCD1.Posx += float64(pCCD1.IntersectRect.W)
			pTCD1.IsNotMoving = true
			pTCD1.IsJumping = false
		}

		if pCCD1.CollisionDirection["bottom"] {
			pTCD1.Posy -= float64(pCCD1.IntersectRect.H)
			pTCD1.Vspeed = 0
			pTCD1.IsJumping = false
		} else if pCCD1.CollisionDirection["top"] {
			pTCD1.Posy += float64(pCCD1.IntersectRect.H)
			pTCD1.IsJumping = true
			pTCD1.Vspeed = 0
		}
	}

	if entityTwoHasDynamicComponent {
		if pCCD1.CollisionDirection["right"] {
			pTCD2.Posx += float64(pCCD2.IntersectRect.W)
			pTCD2.IsNotMoving = true
			pTCD2.IsJumping = false
		} else if pCCD1.CollisionDirection["left"]  {
			pTCD2.Posx -= float64(pCCD2.IntersectRect.W)
			pTCD2.IsNotMoving = true
			pTCD2.IsJumping = false
		}

		if pCCD1.CollisionDirection["bottom"] {
			pTCD2.Posy += float64(pCCD2.IntersectRect.H)
			pTCD2.Vspeed = 0
			pTCD2.IsJumping = false
		} else if pCCD1.CollisionDirection["top"] {
			pTCD2.Posy -= float64(pCCD2.IntersectRect.H)
			pTCD2.IsJumping = true
			pTCD2.Vspeed = 0
		}
	}
}