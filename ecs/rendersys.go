package ecs

import (
	"codeberg.org/alluneedistux/GoJumpRunShoot/statemachine"
	"github.com/veandco/go-sdl2/sdl"
)

type RenderComponentData struct {
	Path     string
	Img      *sdl.Surface
	Texture  *sdl.Texture
	Text     *sdl.Surface
	FontSize uint8
}

type RenderSystem struct {
	*CommonSystemData
	Renderer *sdl.Renderer
}

func NewRenderSystem(e *ECSManager, renderer *sdl.Renderer) *RenderSystem {
	return &RenderSystem{
		CommonSystemData: NewCommonSystemData("RENDER_COMPONENT", e),
		Renderer:         renderer,
	}
}

func (sys *RenderSystem) Run(delta float64, statemachine *statemachine.StateMachine) {
	ecsManager := sys.ECSManager
	entityToComponentMapOrdered := ecsManager.EntityToComponentMapOrdered

	sys.Renderer.Clear()

	for el := entityToComponentMapOrdered.Front(); el != nil; el = el.Next() {
		components := el.Value.([]uint16)
		entityID := el.Key.(uint64)

		if !sys.ECSManager.HasComponent(components, sys.SystemID) {
			continue
		}

		pRCD := sys.GetComponentData(entityID).(*RenderComponentData)
		pTCD := sys.ECSManager.GetComponentDataByName(entityID, "TRANSFORM_COMPONENT").(*TransformComponentData)

		sliceWithComponentData := make([]interface{}, 2, 2)
		sliceWithComponentData[0] = pRCD
		sliceWithComponentData[1] = pTCD

		sliceOtherParametersUpdateComponent := make([]interface{}, 1, 1)

		sys.UpdateComponent(delta, sliceWithComponentData, sliceOtherParametersUpdateComponent)
	}
	sys.Renderer.Present()
}

/*func (sys *RenderSystem) Run(delta float64) {
	ecsManager := sys.ECSManager
	entityToComponentMap := ecsManager.EntityToComponentMap

	sys.Renderer.Clear()

	for entityID, components := range *entityToComponentMap {
		log.Printf("%+v\n", entityID)
		if !sys.ECSManager.HasComponent(components, sys.SystemID) {
			continue
		}

		pRCD := sys.GetComponentData(entityID).(*RenderComponentData)
		pTCD := sys.ECSManager.GetComponentDataByName(entityID, "TRANSFORM_COMPONENT").(*TransformComponentData)

		sliceWithComponentData := make([]interface{}, 2, 2)
		sliceWithComponentData[0] = pRCD
		sliceWithComponentData[1] = pTCD

		sliceOtherParametersUpdateComponent := make([]interface{}, 1, 1)

		sys.UpdateComponent(delta, sliceWithComponentData, sliceOtherParametersUpdateComponent)
	}
	sys.Renderer.Present()
}*/

func (sys *RenderSystem) UpdateComponent(delta float64, sliceWithComponentData []interface{}, sliceOtherParametersUpdateComponent []interface{}) {
	pRCD := sliceWithComponentData[0].(*RenderComponentData)
	pTCD := sliceWithComponentData[1].(*TransformComponentData)

	var img *sdl.Surface
	var h int32
	var w int32
	var dstRect *sdl.Rect
	var sdlFlip sdl.RendererFlip
	texture := pRCD.Texture

	if pRCD.Img != nil {
		img = pRCD.Img
		h = img.H
		w = img.W
		dstRect = &sdl.Rect{X: int32(pTCD.Posx), Y: int32(pTCD.Posy), W: w, H: h}
	} else {
		dstRect = &sdl.Rect{X: int32(pTCD.Posx), Y: int32(pTCD.Posy), W: 125, H: 25}
	}

	if pTCD.FlipImg {
		sdlFlip = sdl.FLIP_HORIZONTAL
	} else {
		sdlFlip = sdl.FLIP_NONE
	}

	sys.Renderer.CopyEx(texture, nil, dstRect, 0.0, nil, sdlFlip)
}