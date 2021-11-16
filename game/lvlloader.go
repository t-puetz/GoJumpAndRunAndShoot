package game

import (
	"codeberg.org/alluneedistux/GoJumpRunShoot/ecs"
	"encoding/json"
	"errors"
	"github.com/elliotchance/orderedmap"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	_ "image/png"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
)

type EntityJSONConfig struct {
	Reference            string   `json:"Reference"`
	Components           []uint16 `json:"Components"`
	InitialPosX          int      `json:"InitialPosX"`
	InitialPosY          int      `json:"InitialPosY"`
	SpreadAlong          string   `json:"SpreadAlong"`
}

type LevelPhysics struct {
	Gravity float64 `json:"Gravity"`
}

type LevelJSONConfig struct {
	LevelPhysics         LevelPhysics                  `json:"LevelPhysics"`
	EntitiesDescriptions *map[string]*EntityJSONConfig `json:"Entities"`
}

func LoadLvlConfig(Game *Game, path string) {
	lvl := &LevelJSONConfig{}

	data, readInErr := ioutil.ReadFile(path)

	if readInErr != nil {
		panic(readInErr)
	}

	unmarshalErr := json.Unmarshal(data, lvl)

	if unmarshalErr != nil {
		panic(unmarshalErr)
	}

	if Game.LvlDescription != nil {
		Game.LvlDescription = nil
	}

	Game.LvlDescription = lvl
}

func (l *LevelJSONConfig) GetEntityDescription(entityID uint64) *EntityJSONConfig {
	for entityIDStr, entityJSONConfig := range *l.EntitiesDescriptions {
		var numKeysUpperLimit uint64
		var numKeysLowerLimit uint64

		if strings.Contains(entityIDStr, "-") {
			// We have a range of Entities
			numKeys := strings.Split(entityIDStr, "-")
			numKeysUpperLimit, _ = strconv.ParseUint(numKeys[1], 10, 64)
			numKeysLowerLimit, _ = strconv.ParseUint(numKeys[0], 10, 64)

			if entityID >= numKeysLowerLimit && entityID <= numKeysUpperLimit {
				return entityJSONConfig
			}
		} else {
			// Single Entity
			entID, _ := strconv.Atoi(entityIDStr)
			if uint64(entID) == entityID {
				return entityJSONConfig
			}
		}
	}
	return nil
}

func (l *LevelJSONConfig) GetFirstEntityIDFromRange(entityID uint64) int {
	for entityIDStr, _ := range *l.EntitiesDescriptions {
		var numKeysUpperLimit uint64
		var numKeysLowerLimit uint64

		if strings.Contains(entityIDStr, "-") {
			// We have a range of Entities
			numKeys := strings.Split(entityIDStr, "-")
			numKeysUpperLimit, _ = strconv.ParseUint(numKeys[1], 10, 64)
			numKeysLowerLimit, _ = strconv.ParseUint(numKeys[0], 10, 64)

			if entityID >= numKeysLowerLimit && entityID <= numKeysUpperLimit {
				return int(numKeysLowerLimit)
			}
		} else {
			// Single Entity
			entID, _ := strconv.Atoi(entityIDStr)
			if uint64(entID) == entityID {
				return int(entityID)
			}
		}
	}
	return -1
}

func CreateEntityComponentMap(pLvlConfig *LevelJSONConfig) *map[uint64][]uint16 {
	lvlConfig := *pLvlConfig
	pEntitiesConfig := lvlConfig.EntitiesDescriptions
	entitiesConfig := *pEntitiesConfig

	entityComponentMap := make(map[uint64][]uint16)

	for key, _ := range entitiesConfig {
		var numKeysUpperLimit uint64
		var numKeysLowerLimit uint64

		if strings.Contains(key, "-") {
			// We have a range of Entities
			numKeys := strings.Split(key, "-")
			numKeysUpperLimit, _ = strconv.ParseUint(numKeys[1], 10, 64)
			numKeysLowerLimit, _ = strconv.ParseUint(numKeys[0], 10, 64)

			for i := numKeysLowerLimit; i <= numKeysUpperLimit; i++ {
				entityComponentMap[i] = entitiesConfig[key].Components
			}
		} else {
			// Single Entity
			entityID, _ := strconv.Atoi(key)
			entityComponentMap[uint64(entityID)] = entitiesConfig[key].Components
		}
	}
	return &entityComponentMap
}

func CreateEntityComponentOrdered(pLvlConfig *LevelJSONConfig) *orderedmap.OrderedMap {
	lvlConfig := *pLvlConfig
	pEntitiesConfig := lvlConfig.EntitiesDescriptions
	entitiesConfig := *pEntitiesConfig

	entityComponentMap := orderedmap.NewOrderedMap()

	for key, _ := range entitiesConfig {
		var numKeysUpperLimit uint64
		var numKeysLowerLimit uint64

		if strings.Contains(key, "-") {
			// We have a range of Entities
			numKeys := strings.Split(key, "-")
			numKeysUpperLimit, _ = strconv.ParseUint(numKeys[1], 10, 64)
			numKeysLowerLimit, _ = strconv.ParseUint(numKeys[0], 10, 64)

			for i := numKeysLowerLimit; i <= numKeysUpperLimit; i++ {
				entityComponentMap.Set(i, entitiesConfig[key].Components)
			}
		} else {
			// Single Entity
			entityID, _ := strconv.Atoi(key)
			entityComponentMap.Set(uint64(entityID), entitiesConfig[key].Components)
		}
	}
	return entityComponentMap
}

func CreateLvlsEntityAndComponents(Game *Game, EntCmpMap *map[uint64][]uint16) {
	for entityID, components := range *(EntCmpMap) {
		Game.ECSManager.InitializeComponentsForEntity(entityID)

		for _, componentID := range components {
			Game.ECSManager.AddComponentToEntity(entityID, componentID)
		}
	}
}

func CreateLvlsEntityAndComponentsOrdered(Game *Game, EntCmpMap *orderedmap.OrderedMap) {
	for el := EntCmpMap.Front(); el != nil; el = el.Next() {
		{
			Game.ECSManager.InitializeComponentsForEntityOrdered(el.Key.(uint64))

			for _, componentID := range el.Value.([]uint16) {
				Game.ECSManager.AddComponentToEntityOrdered(el.Key.(uint64), componentID)
			}
		}
	}
}

func LoadImagesAndTextures(Game *Game) {
	pLvlConfig := Game.LvlDescription
	pAssetDescriptions := Game.AssetDescriptions

	for key, EntityDescription := range *pLvlConfig.EntitiesDescriptions {
		entityIDStr := key

		var numKeysUpperLimit uint64
		var numKeysLowerLimit uint64

		if strings.Contains(entityIDStr, "-") {
			numKeys := strings.Split(entityIDStr, "-")
			numKeysUpperLimit, _ = strconv.ParseUint(numKeys[1], 10, 64)
			numKeysLowerLimit, _ = strconv.ParseUint(numKeys[0], 10, 64)

			for i := numKeysLowerLimit; i <= numKeysUpperLimit; i++ {
				entityIDStr = strconv.Itoa(int(i))
				assertImageDataRenderAndAnimationComponentData(Game, entityIDStr, EntityDescription, pAssetDescriptions)
			}
		} else {
			assertImageDataRenderAndAnimationComponentData(Game, entityIDStr, EntityDescription, pAssetDescriptions)
		}
	}
}

func assertImageDataRenderAndAnimationComponentData(game *Game, entityIDStr string, entityDescription *EntityJSONConfig, assetDescriptions *map[string]*AssetJSONConfig) {
	entityManager := game.ECSManager
	entityID, _ := strconv.Atoi(entityIDStr)

	reference := entityDescription.Reference

	pRCD := entityManager.GetComponentDataByName(uint64(entityID),"RENDER_COMPONENT").(*ecs.RenderComponentData)
	pACD := entityManager.GetComponentDataByName(uint64(entityID),"ANIMATE_COMPONENT").(*ecs.AnimateComponentData)

	var fullImagePath string
	var imageName string
	var pTexture *sdl.Texture
	var pImage *sdl.Surface

	mainEntity := (*assetDescriptions)[reference]

	if mainEntity.AnimatedByDefault {
		basePath := mainEntity.ImagesBasePath

		if pRCD.Text != nil {
			return
		}

		imageCounter := 0

		for animationType, _ := range *mainEntity.Animations {
			// TODO: Take care of ALL animation types
			if animationType != "Front" {
				continue
			}

			imageName = (*mainEntity.Animations)[animationType].Image
			fullImagePath = basePath + imageName

			err := errors.New("")

			pImage, err = img.Load(fullImagePath)

			if err != nil {
				log.Fatalf("Not able to create image for RenderComponent of Entity number %s\n", entityIDStr)
			}

			pTexture, err = game.Renderer.CreateTextureFromSurface(pImage)

			if err != nil {
				log.Fatalf("Not able to create texture from surface for RenderComponent of Entity number %s\n", entityIDStr)
			}

			// If we have multiple images for animations as in this case here
			// the RenderComponentData as initial data gets the data of the first image
            // TODO: Take care of ALL animation types!
			if imageCounter == 0 && animationType == "Front" {
				pRCD.Path = fullImagePath
				pRCD.Img = pImage
				pRCD.Texture = pTexture
			}

			// The AnimationComponentData gets everything
			pACDCore := &ecs.AnimationComponentDataCore{}

			pACDCore.Paths = make([]string, 0, 4)
			pACDCore.Images = make([]*sdl.Surface, 0, 4)
			pACDCore.Textures = make([]*sdl.Texture, 0, 4)

			pACDCore.Paths = append(pACDCore.Paths, fullImagePath)
			pACDCore.Images = append(pACDCore.Images, pImage)
			pACDCore.Textures = append(pACDCore.Textures, pTexture)

			pACDDataMap := make(map[string]*ecs.AnimationComponentDataCore)
			pACD.AnimationData = &pACDDataMap
			(*pACD.AnimationData)[animationType] = pACDCore

			imageCounter += 1
		}
	} else {
		// Non-animated assets

		if mainEntity.Image != "" && mainEntity.ImagesBasePath != "" {
			// Image assets

			fullImagePath = mainEntity.ImagesBasePath + mainEntity.Image

			err := errors.New("")

			pImage, err = img.Load(fullImagePath)

			if err != nil {
				log.Fatalf("Not able to create image for RenderComponent of Entity number %s from path %s\n", entityIDStr, fullImagePath)
			}

			pTexture, err = game.Renderer.CreateTextureFromSurface(pImage)

			if err != nil {
				log.Fatalf("Not able to create texture from surface for RenderComponent of Entity number %s from path %s\n", entityIDStr, fullImagePath)
			}

			pRCD.Path = fullImagePath
			pRCD.Texture = pTexture
			pRCD.Img = pImage
		}

		if mainEntity.Text != "" && mainEntity.FontSize > 0 {
			// Text assets

			var textToRender string
			var font *ttf.Font
			var text *sdl.Surface
			var err error

			textToRender = mainEntity.Text

			font, err = ttf.OpenFont("./assets/SourceCodePro-Bold.ttf", int(mainEntity.FontSize))

			if err != nil {
				panic("Error creating SDL font.")
			}


			text, err = font.RenderUTF8Blended(textToRender, sdl.Color{R: 255, G: 0, B: 0, A: 255})

			if err != nil {
				panic("Error creating SDL text.")
			}

			textTexture, err := game.Renderer.CreateTextureFromSurface(text)

			if err != nil {
				panic("Error creating text texture.")
			}

			pRCD.Text = text
			pRCD.Texture = textTexture
		}
	}
}

func InitializeLevel(g *Game) {
	entityComponentMap := CreateEntityComponentMap(g.LvlDescription)
	CreateLvlsEntityAndComponents(g, entityComponentMap)
	g.ECSManager.LinkComponentsWithProperDataStruct()
	LoadImagesAndTextures(g)
	TransformSystemSetInitialVals(g)
}

func InitializeLevelOrdered(g *Game) {
	entityComponentMap := CreateEntityComponentOrdered(g.LvlDescription)
	CreateLvlsEntityAndComponentsOrdered(g, entityComponentMap)
	g.ECSManager.LinkComponentsWithProperDataStructOrdered()
	LoadImagesAndTextures(g)
	TransformSystemSetInitialValsOrdered(g)
}

func TransformSystemSetInitialVals(g *Game) {
	// Get the entity config map keys that represent entity ranges
	lvlConfig := g.LvlDescription

	for entityID, components := range *g.ECSManager.EntityToComponentMap {
        hasTransformComponent := g.ECSManager.HasNamedComponent(components, "TRANSFORM_COMPONENT")

		if ! hasTransformComponent {
			continue
		}

		pTCD := g.ECSManager.GetComponentDataByName(entityID, "TRANSFORM_COMPONENT").(*ecs.TransformComponentData)
		entityJSONConfig := lvlConfig.GetEntityDescription(entityID)

		if entityJSONConfig.SpreadAlong == "X" {
			pRCD := g.ECSManager.GetComponentDataByName(entityID, "RENDER_COMPONENT").(*ecs.RenderComponentData)
			firstEntity := lvlConfig.GetFirstEntityIDFromRange(entityID)
			pTCD.Posy = float64(entityJSONConfig.InitialPosY)
			pTCD.Posx = float64(entityJSONConfig.InitialPosX) + float64(pRCD.Img.W) * (float64(entityID) - float64(firstEntity))
		} else {
			pTCD.Posx = float64(entityJSONConfig.InitialPosX)
			pTCD.Posy = float64(entityJSONConfig.InitialPosY)
		}

		pTCD.FlipImg = false
		pTCD.IsJumping = false
		pTCD.Hspeed = 0
	}
}

func TransformSystemSetInitialValsOrdered(g *Game) {
	// Get the entity config map keys that represent entity ranges
	lvlConfig := g.LvlDescription

	for el := g.ECSManager.EntityToComponentMapOrdered.Front(); el != nil; el = el.Next() {
		hasTransformComponent := g.ECSManager.HasNamedComponent(el.Value.([]uint16), "TRANSFORM_COMPONENT")

		if ! hasTransformComponent {
			continue
		}

		pTCD := g.ECSManager.GetComponentDataByName(el.Key.(uint64), "TRANSFORM_COMPONENT").(*ecs.TransformComponentData)
		entityJSONConfig := lvlConfig.GetEntityDescription(el.Key.(uint64))

		if entityJSONConfig.SpreadAlong == "X" {
			pRCD := g.ECSManager.GetComponentDataByName(el.Key.(uint64), "RENDER_COMPONENT").(*ecs.RenderComponentData)
			firstEntity := lvlConfig.GetFirstEntityIDFromRange(el.Key.(uint64))
			pTCD.Posy = float64(entityJSONConfig.InitialPosY)
			pTCD.Posx = float64(entityJSONConfig.InitialPosX) + float64(pRCD.Img.W) * (float64(el.Key.(uint64)) - float64(firstEntity))
		} else {
			pTCD.Posx = float64(entityJSONConfig.InitialPosX)
			pTCD.Posy = float64(entityJSONConfig.InitialPosY)
		}

		pTCD.FlipImg = false
		pTCD.IsJumping = false
		pTCD.Hspeed = 0
	}
}