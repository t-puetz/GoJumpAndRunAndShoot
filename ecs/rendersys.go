package ecs

import (
	"github.com/t-puetz/GoJumpAndRunAndShoot/statemachine"
	"github.com/veandco/go-sdl2/sdl"
	"sync"
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
	mu *sync.Mutex
}

func NewRenderSystem(e *ECSManager, renderer *sdl.Renderer) *RenderSystem {
	return &RenderSystem{
		CommonSystemData: NewCommonSystemData("RENDER_COMPONENT", e),
		Renderer:         renderer,
		mu: &sync.Mutex{},
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

    sys.mu.Lock()
	sys.Renderer.Present()
	sys.mu.Unlock()
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

	renderImage := pRCD.Image != nil && pRCD.Text == nil
	renderText := pRCD.Image == nil && pRCD.Text != nil

	if renderImage {
		img = pRCD.Image
		h = img.H
		w = img.W
		dstRect = &sdl.Rect{X: int32(pTCD.PosX), Y: int32(pTCD.PosY), W: w, H: h}

		if pTCD.FlipImg {
			sdlFlip = sdl.FLIP_HORIZONTAL
		} else {
			sdlFlip = sdl.FLIP_NONE
		}
	} else if renderText {
		dstRect = &sdl.Rect{X: int32(pTCD.PosX), Y: int32(pTCD.PosY), W: 125, H: 25}
	} else {
		return
	}

	sys.Renderer.CopyEx(texture, nil, dstRect, 0.0, nil, sdlFlip)
}
