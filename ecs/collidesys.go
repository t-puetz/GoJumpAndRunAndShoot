package ecs

import (
	"github.com/t-puetz/GoJumpAndRunAndShoot/statemachine"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"math"
)

type CollisionCoreData struct {
	CollisionDirection     map[string]bool
	LastCollisionDirection map[string]bool
	IntersectRect          *sdl.Rect
	EntityCollidingWith    uint64
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

		if !sys.ECSManager.HasNamedComponent(components, "DYNAMIC_COMPONENT") {
			continue
		}

		entityOneHasDynamicComponent := sys.ECSManager.HasNamedComponent(components, "DYNAMIC_COMPONENT")
		lengthOfEntityComponentMap := uint64(entityToComponentMapOrdered.Len())

		for j := entityID + 1; j < lengthOfEntityComponentMap; j++ {
			componentsEntityTwo, _ := entityToComponentMapOrdered.Get(j)

			if !sys.ECSManager.HasNamedComponent(componentsEntityTwo.([]uint16), "COLLIDE_COMPONENT") {
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
			pRCD1 := sys.ECSManager.GetComponentDataByName(ent1, "RENDER_COMPONENT").(*RenderComponentData)
			pRCD2 := sys.ECSManager.GetComponentDataByName(ent2, "RENDER_COMPONENT").(*RenderComponentData)
			pCCD1 := sys.GetComponentData(ent1).(*CollisionComponentData)
			pCCD2 := sys.GetComponentData(ent2).(*CollisionComponentData)

			if math.Abs(float64(pTCD1.PosX-pTCD2.PosX)) > float64(pRCD1.Image.W/2) && math.Abs(float64(pTCD1.PosY-pTCD2.PosY)) > float64(pRCD1.Image.H/2) ||
				math.Abs(float64(pTCD1.PosX-pTCD2.PosX)) > float64(pRCD2.Image.W/2) && math.Abs(float64(pTCD1.PosY-pTCD2.PosY)) > float64(pRCD2.Image.H/2) {

				// Skip entities that are to far away from each other anyways
				continue
			}

			intersectRect, areColliding := sys.detect(sys.ECSManager, ent1, ent2, pTCD1, pTCD2, pCCD1, pCCD2)

			if !areColliding || intersectRect == nil {
				continue
			}

			sys.UpdateComponent(delta, intersectRect, pTCD1, pTCD2, entityOneHasDynamicComponent, entityTwoHasDynamicComponent, pCCD1, pCCD2)
		}
	}
}

func (sys *CollideSystem) detect(ecsManager *ECSManager, ent1, ent2 uint64, pTCD1, pTCD2 *TransformComponentData, pCCD1, pCCD2 *CollisionComponentData) (*sdl.Rect, bool) {
	imgRectOne := sys.ECSManager.GetEntityRect(ent1)
	imgRectTwo := sys.ECSManager.GetEntityRect(ent2)

	type Line struct {
		pointA *sdl.Point
		pointB *sdl.Point
	}

	intersectRect, areColliding := imgRectOne.Intersect(imgRectTwo)

	if areColliding {
		log.Println("Entities ARE colliding. Checking from what directions.")
		topLineRectOne := Line{pointA: &sdl.Point{X: imgRectOne.X, Y: imgRectOne.Y}, pointB: &sdl.Point{X: imgRectOne.X + imgRectOne.W, Y: imgRectOne.Y}}
		leftLineRectOne := Line{pointA: topLineRectOne.pointA, pointB: &sdl.Point{X: imgRectOne.X, Y: imgRectOne.Y + imgRectOne.H}}
		bottomLineRectOne := Line{pointA: leftLineRectOne.pointB, pointB: &sdl.Point{X: imgRectOne.X + imgRectOne.W, Y: leftLineRectOne.pointB.Y}}
		rightLineRectOne := Line{pointA: topLineRectOne.pointB, pointB: bottomLineRectOne.pointB}

		topLineRectTwo := Line{pointA: &sdl.Point{X: imgRectTwo.X, Y: imgRectTwo.Y}, pointB: &sdl.Point{X: imgRectTwo.X + imgRectTwo.W, Y: imgRectTwo.Y}}
		leftLineRectTwo := Line{pointA: topLineRectTwo.pointA, pointB: &sdl.Point{X: imgRectTwo.X, Y: imgRectTwo.Y + imgRectTwo.H}}
		bottomLineRectTwo := Line{pointA: leftLineRectTwo.pointB, pointB: &sdl.Point{X: imgRectTwo.X + imgRectTwo.W, Y: leftLineRectTwo.pointB.Y}}
		rightLineRectTwo := Line{pointA: topLineRectTwo.pointB, pointB: bottomLineRectTwo.pointB}

		if rightLineRectOne.pointA.X >= leftLineRectTwo.pointA.X && leftLineRectOne.pointA.X < rightLineRectTwo.pointA.X && pTCD1.PosX > pTCD1.LastPosX {
			pCCD1.LastCollisionDirection = pCCD1.CollisionDirection
			pCCD2.LastCollisionDirection = pCCD2.CollisionDirection
			pCCD1.CollisionDirection["right"] = true
			pCCD1.CollisionDirection["left"] = false
			pCCD2.CollisionDirection["right"] = false
			pCCD2.CollisionDirection["left"] = true
		}

		if leftLineRectOne.pointA.X >= rightLineRectTwo.pointA.X && rightLineRectOne.pointA.X < leftLineRectTwo.pointA.X && pTCD1.PosX < pTCD1.LastPosX {
			pCCD1.LastCollisionDirection = pCCD1.CollisionDirection
			pCCD2.LastCollisionDirection = pCCD2.CollisionDirection
			pCCD1.CollisionDirection["left"] = true
			pCCD1.CollisionDirection["right"] = false
			pCCD2.CollisionDirection["left"] = false
			pCCD2.CollisionDirection["right"] = true
		}

		if bottomLineRectOne.pointA.Y >= topLineRectTwo.pointA.Y && topLineRectOne.pointA.Y < bottomLineRectTwo.pointA.Y && pTCD1.PosY > pTCD1.LastPosY {
			pCCD1.LastCollisionDirection = pCCD1.CollisionDirection
			pCCD1.CollisionDirection["bottom"] = true
			pCCD1.CollisionDirection["top"] = false
		}

		if topLineRectOne.pointA.Y <= bottomLineRectTwo.pointA.Y && bottomLineRectOne.pointA.Y > topLineRectTwo.pointA.Y && pTCD1.PosY < pTCD1.LastPosY {
			pCCD1.LastCollisionDirection = pCCD1.CollisionDirection
			pCCD1.CollisionDirection["top"] = true
			pCCD1.CollisionDirection["bottom"] = false
		}

		log.Println(intersectRect, areColliding)
		return &intersectRect, areColliding
	}

	return nil, false
}

func (sys *CollideSystem) resolve(intersectRect *sdl.Rect, pTCD1, pTCD2 *TransformComponentData, entityOneHasDynamicComponent, entityTwoHasDynamicComponent bool, pCCD1, pCCD2 *CollisionComponentData) {
	log.Println(pCCD1.CollisionDirection)
	log.Println(pCCD2.CollisionDirection)
	if entityOneHasDynamicComponent && pCCD1.CollisionDirection["right"] && !pCCD1.LastCollisionDirection["right"] {
		pTCD1.PosX -= intersectRect.W
		pTCD1.Hspeed = 0
	}

	if entityOneHasDynamicComponent && pCCD1.CollisionDirection["left"] {
		pTCD1.PosX += intersectRect.W
		pTCD1.Hspeed *= -1
	}

	if entityOneHasDynamicComponent && pCCD1.CollisionDirection["bottom"] && !pCCD2.LastCollisionDirection["top"] {
		pTCD1.PosY -= intersectRect.H
		pTCD1.Vspeed = 0
		pTCD1.IsJumping = false
	}

	if entityOneHasDynamicComponent && pCCD1.CollisionDirection["top"] {
		pTCD1.IsJumping = false
		pTCD1.Vspeed = 0
		pTCD1.PosY += intersectRect.H
	}

}

func (sys *CollideSystem) UpdateComponent(delta float64, essentialData ...interface{}) {

	intersectRect := essentialData[0].(*sdl.Rect)
	pTCD1 := essentialData[1].(*TransformComponentData)
	pTCD2 := essentialData[2].(*TransformComponentData)
	entityOneHasDynamicComponent := essentialData[3].(bool)
	entityTwoHasDynamicComponent := essentialData[4].(bool)
	pCCD1 := essentialData[5].(*CollisionComponentData)
	pCCD2 := essentialData[6].(*CollisionComponentData)

	sys.resolve(intersectRect, pTCD1, pTCD2, entityOneHasDynamicComponent, entityTwoHasDynamicComponent, pCCD1, pCCD2)
}
