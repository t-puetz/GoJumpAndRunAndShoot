package ecs

import (
	"codeberg.org/alluneedistux/GoJumpRunShoot/statemachine"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"strings"
)

type RenderComponentData struct {
	Path     string
	Image    *sdl.Surface
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
	entityToComponentMapOrdered := ecsManager.EntityToComponentMap

	sys.Renderer.Clear()

	for el := entityToComponentMapOrdered.Front(); el != nil; el = el.Next() {
		components := el.Value.([]uint16)
		entityID := el.Key.(uint64)

		if !sys.ECSManager.HasComponent(components, sys.SystemID) {
			continue
		}

		pRCD := sys.GetComponentData(entityID).(*RenderComponentData)
		pTCD := sys.ECSManager.GetComponentDataByName(entityID, "TRANSFORM_COMPONENT").(*TransformComponentData)

		sys.UpdateComponent(delta, pRCD, pTCD)
	}
	sys.Renderer.Present()
}

func (sys *RenderSystem) UpdateComponent(delta float64, essentialData ...interface{}) {
	pRCD := essentialData[0].(*RenderComponentData)
	pTCD := essentialData[1].(*TransformComponentData)

	var img *sdl.Surface
	var h int32
	var w int32
	var dstRect *sdl.Rect
	var sdlFlip sdl.RendererFlip
	texture := pRCD.Texture

	if pRCD.Image != nil {
		// We render images
		img = pRCD.Image
		h = img.H
		w = img.W
		dstRect = &sdl.Rect{X: int32(pTCD.Posx), Y: int32(pTCD.Posy), W: w, H: h}

		if pTCD.FlipImg {
			sdlFlip = sdl.FLIP_HORIZONTAL
		} else {
			sdlFlip = sdl.FLIP_NONE
		}

		// Just for debugging
		if strings.Contains(pRCD.Path, "walk") {
			// TODO: Why do I only ever get image 11 "p1_walk11.png" from the AnimateSystem although the AnimateSystem correctly finds ALL the images and passes them to the render data???
			log.Printf("Image Render Sytem: %v\n", pRCD.Path)
		}

		//if strings.Contains(pRCD.Path, "p1") {
		//	log.Println(pTCD.IsNotMoving)
		//}
		// End of debugging
	} else {
		// We render text
		dstRect = &sdl.Rect{X: int32(pTCD.Posx), Y: int32(pTCD.Posy), W: 125, H: 25}
	}

	sys.Renderer.CopyEx(texture, nil, dstRect, 0.0, nil, sdlFlip)
}
