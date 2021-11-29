package game

import (
	"encoding/json"
	"errors"
	"github.com/elliotchance/orderedmap"
	"github.com/t-puetz/GoJumpAndRunAndShoot/ecs"
	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	_ "image/png"
	"io/ioutil"
	"log"
	"sort"
	"strconv"
	"strings"
)

type EntityJSONConfig struct {
	Reference   string   `json:"Reference"`
	Components  []uint16 `json:"Components"`
	InitialPosX int      `json:"InitialPosX"`
	InitialPosY int      `json:"InitialPosY"`
	SpreadAlong string   `json:"SpreadAlong"`
}

type LevelPhysics struct {
	Gravity float64 `json:"Gravity"`
}

type LevelJSONConfig struct {
	LevelPhysics         LevelPhysics                  `json:"LevelPhysics"`
	EntitiesDescriptions *map[string]*EntityJSONConfig `json:"Entities"`
	EntitiesDescriptionsOrdered orderedmap.OrderedMap
}

func convertUnorderedToOrderedEntityDescriptionsMap(unorderedMap *map[string]*EntityJSONConfig) *orderedmap.OrderedMap {
    keysStrSliceUnordered := make([]string, 0, len(*unorderedMap))
    keysStrSliceOrdered := make([]string, 0, len(*unorderedMap))
    keysIntSliceOrdered := make([]int, 0, len(*unorderedMap))

	orderedMap := orderedmap.NewOrderedMap()

	var numKeysUpperLimit uint64
	var numKeysLowerLimit uint64

	// First extract all the string keys from the unordered, native Go map
	// and put them in a string slice. If we find a range marked by a hyphen, unravel it
	for unorderedKey,  _ := range *unorderedMap {
		if !strings.Contains(unorderedKey, "-") {
			// Single entity (no hyphen as range indicator)
            keysStrSliceUnordered = append(keysStrSliceUnordered, unorderedKey)
		} else {
			numKeys := strings.Split(unorderedKey, "-")
			numKeysUpperLimit, _ = strconv.ParseUint(numKeys[1], 10, 64)
			numKeysLowerLimit, _ = strconv.ParseUint(numKeys[0], 10, 64)

			var i uint64
			for i = numKeysLowerLimit; i < numKeysUpperLimit + 1; i++ {
                keysStrSliceUnordered = append(keysStrSliceUnordered, strconv.Itoa(int(i)))
			}
		}
	}

	// Convert the string slice to an int slice to order the keys
	for _, keyStrUnordered := range keysStrSliceUnordered {
		keyAsInt, _ := strconv.Atoi(keyStrUnordered)
        keysIntSliceOrdered = append(keysIntSliceOrdered, keyAsInt)
	}

	// Sort the int slice
	sort.Ints(keysIntSliceOrdered)

    // Convert the ordered int slice back to an ordered string slice
	for _, keyIntOrdered := range keysIntSliceOrdered {
		keysStrSliceOrdered = append(keysStrSliceOrdered, strconv.Itoa(keyIntOrdered))
	}

	for _,  keyStrOrdered := range keysStrSliceOrdered {
		for unorderedOriginalKey, entityJSONConfig := range *unorderedMap {
			if !strings.Contains(unorderedOriginalKey, "-") {
				// Single entity (no hyphen as range indicator)
				if keyStrOrdered == unorderedOriginalKey {
					keyIntOrdered, _ := strconv.Atoi(keyStrOrdered)
					orderedMap.Set(uint64(keyIntOrdered), entityJSONConfig)
				}
			} else {
				numKeys := strings.Split(unorderedOriginalKey, "-")
				numKeysUpperLimit, _ = strconv.ParseUint(numKeys[1], 10, 64)
				numKeysLowerLimit, _ = strconv.ParseUint(numKeys[0], 10, 64)

				var i uint64
				for i = numKeysLowerLimit; i < numKeysUpperLimit + 1; i++ {
					if keyStrOrdered == strconv.Itoa(int(i)) {
						keyIntOrdered, _ := strconv.Atoi(keyStrOrdered)
						orderedMap.Set(uint64(keyIntOrdered), entityJSONConfig)
					}
				}
			}
		}
	}
	return orderedMap
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
	Game.LvlDescription.EntitiesDescriptionsOrdered = *(convertUnorderedToOrderedEntityDescriptionsMap(lvl.EntitiesDescriptions))
}

func (l *LevelJSONConfig) GetEntityDescription(entityID uint64) *EntityJSONConfig {
	for el := l.EntitiesDescriptionsOrdered.Front(); el != nil; el = el.Next() {
		currentEntityID := el.Key.(uint64)
		entityJSONConfig := el.Value.(*EntityJSONConfig)

		if entityID == currentEntityID {
			return entityJSONConfig
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
		}
	}
	return -1
}

func CreateEntityComponent(pLvlConfig *LevelJSONConfig) *orderedmap.OrderedMap {
	lvlConfig := *pLvlConfig
	entitiesConfig := lvlConfig.EntitiesDescriptionsOrdered
	entityComponentMap := orderedmap.NewOrderedMap()

	for el := entitiesConfig.Front(); el != nil; el = el.Next() {
		currentEntityID := el.Key.(uint64)
		entityJSONConfig := el.Value.(*EntityJSONConfig)

		entityComponentMap.Set(currentEntityID, (*entityJSONConfig).Components)
	}

	return entityComponentMap
}

func CreateLvlsEntityAndComponents(Game *Game, EntCmpMap *orderedmap.OrderedMap) {
	for el := EntCmpMap.Front(); el != nil; el = el.Next() {
		components := el.Value.([]uint16)
		entityID := el.Key.(uint64)

		{
			Game.ECSManager.InitializeComponentsForEntity(entityID)

			for _, componentID := range components {
				Game.ECSManager.AddComponentToEntity(entityID, componentID)
			}
		}
	}
}

func LoadImagesAndTextures(Game *Game) {
	pLvlConfig := Game.LvlDescription
	pAssetDescriptions := Game.AssetDescriptions

	for el := pLvlConfig.EntitiesDescriptionsOrdered.Front(); el != nil; el = el.Next() {
		currentEntityID := el.Key.(uint64)
		entityJSONConfig := el.Value.(*EntityJSONConfig)

		entityIDStr := strconv.Itoa(int(currentEntityID))

		connectStillImageDataWithRenderAndAnimateComponentData(Game, entityIDStr, entityJSONConfig, pAssetDescriptions)
		connectAnimatedImageDataWithRenderAndAnimateComponentData(Game, entityIDStr, entityJSONConfig, pAssetDescriptions)
		connectTextDataWithRenderAndAnimateComponentData(Game, entityIDStr, entityJSONConfig, pAssetDescriptions)
	}
}


func connectAnimatedImageDataWithRenderAndAnimateComponentData(game *Game, entityIDStr string, entityDescription *EntityJSONConfig, assetDescriptions *map[string]*AssetJSONConfig) {
	entityManager := game.ECSManager
	entityID, _ := strconv.Atoi(entityIDStr)

	reference := entityDescription.Reference

	pRCD := entityManager.GetComponentDataByName(uint64(entityID), "RENDER_COMPONENT").(*ecs.RenderComponentData)
	pACD := entityManager.GetComponentDataByName(uint64(entityID), "ANIMATE_COMPONENT").(*ecs.AnimateComponentData)

	if pRCD.Text != nil {
		return
	}

	var fullImagePath string
	var imageName string
	var pTexture *sdl.Texture
	var pImage *sdl.Surface

	mainEntity := (*assetDescriptions)[reference]

	if mainEntity.AnimatedByDefault {
		basePath := mainEntity.ImagesBasePath

		for animationType, _ := range *mainEntity.Animations {
			// TODO: Take care of ALL animation types

			imageName = (*mainEntity.Animations)[animationType].Image

			if imageName == "" {
				continue
			}

			fullImagePath = basePath + imageName
			pACDCore := &ecs.AnimationComponentDataCore{}
			pACDCore.Paths = make([]string, 0, 0)
			pACDCore.Images = make([]*sdl.Surface, 0, 0)
			pACDCore.Textures = make([]*sdl.Texture, 0, 0)

			imagePathsWhenRange, moreThanOneImage := connectAnimatedImageDataWithRenderAndAnimateComponentUnwrapImageRange(fullImagePath)
			var imagesWhenRange []*sdl.Surface
			var texturesWhenRange []*sdl.Texture
			err := errors.New("")

			if moreThanOneImage {
				for i := 0; i < len(imagePathsWhenRange); i++ {
					if i == 0 {
						imagesWhenRange = make([]*sdl.Surface, len(imagePathsWhenRange), len(imagePathsWhenRange))
					}

					pImage, err = img.Load(imagePathsWhenRange[i])

					if err != nil {
						log.Fatalf("Not able to create image for RenderComponent of Entity number %s\n", entityIDStr)
					}

					imagesWhenRange[i] = pImage

					if i == 0 {
						texturesWhenRange = make([]*sdl.Texture, len(imagePathsWhenRange), len(imagePathsWhenRange))
					}

					pTexture, err = game.Renderer.CreateTextureFromSurface(imagesWhenRange[i])

					if err != nil {
						log.Fatalf("Not able to create texture from surface for RenderComponent of Entity number %s\n", entityIDStr)
					}

					texturesWhenRange[i] = pTexture
				}

				pRCD.Path = imagePathsWhenRange[0]
				pRCD.Image = imagesWhenRange[0]
				pRCD.Texture = texturesWhenRange[0]

				pACDCore.Paths = append(pACDCore.Paths, imagePathsWhenRange...)
				pACDCore.Images = append(pACDCore.Images, imagesWhenRange...)
				pACDCore.Textures = append(pACDCore.Textures, texturesWhenRange...)
			} else {
				pImage, err = img.Load(fullImagePath)

				if err != nil {
					log.Fatalf("Not able to create image for RenderComponent of Entity number %s\n", entityIDStr)
				}

				pTexture, err = game.Renderer.CreateTextureFromSurface(pImage)

				if err != nil {
					log.Fatalf("Not able to create texture from surface for RenderComponent of Entity number %s\n", entityIDStr)
				}

				pRCD.Path = fullImagePath
				pRCD.Image = pImage
				pRCD.Texture = pTexture

				pACDCore.Paths = append(pACDCore.Paths, fullImagePath)
				pACDCore.Images = append(pACDCore.Images, pImage)
				pACDCore.Textures = append(pACDCore.Textures, pTexture)
			}

			pACDCore.NumberAnimations = (*mainEntity.Animations)[animationType].NumberAnimations
			pACDCore.DefaultAnimationDuration = mainEntity.DefaultAnimationDuration

			(*pACD.AnimationData)[animationType] = pACDCore
		}
	}
}

func connectAnimatedImageDataWithRenderAndAnimateComponentUnwrapImageRange(fullImagePath string) ([]string, bool) {
	var imagePathsWhenRange []string

	if strings.Contains(fullImagePath, "till") && strings.Count(fullImagePath, "|") == 2 {
		posFirstPipe := strings.Index(fullImagePath, "|")
		posLastPipe := strings.LastIndex(fullImagePath, "|")
		posTill := strings.Index(fullImagePath, "till")
		firstImageNumber := fullImagePath[posFirstPipe+1:posTill]
		lastImageNumber := fullImagePath[posTill+4:posLastPipe]
		lowerBound, _ := strconv.Atoi(firstImageNumber)
		upperBound, _ := strconv.Atoi(lastImageNumber)

		assumeZeroPadding := false

		if strings.HasPrefix(firstImageNumber, "0") {
			assumeZeroPadding = true
		}
		posLastFwdSlash := strings.LastIndex(fullImagePath, "/")
		fullBasePath := fullImagePath[:posLastFwdSlash+1]
		posFormatDot := strings.LastIndex(fullImagePath, ".")
		format := fullImagePath[posFormatDot:len(fullImagePath)]
		posFirstImageNumber := strings.Index(fullImagePath, firstImageNumber)
		baseImageName := fullImagePath[posLastFwdSlash+1:posFirstImageNumber]
		basePathWithBaseImageName := fullBasePath + baseImageName
		imagePathsWhenRange = make([]string, 0, 0)

		for i := lowerBound; i <= upperBound; i++ {
			if assumeZeroPadding {
				if i < 10 {
					imagePathsWhenRange = append(imagePathsWhenRange, basePathWithBaseImageName + "0" + strconv.Itoa(i) + format)
				} else {
					imagePathsWhenRange = append(imagePathsWhenRange, basePathWithBaseImageName + strconv.Itoa(i) + format)
				}
			} else {
				imagePathsWhenRange = append(imagePathsWhenRange, basePathWithBaseImageName + strconv.Itoa(i) + format)
			}
		}
	}
	return imagePathsWhenRange, len(imagePathsWhenRange) > 1
}

func connectStillImageDataWithRenderAndAnimateComponentData(game *Game, entityIDStr string, entityDescription *EntityJSONConfig, assetDescriptions *map[string]*AssetJSONConfig) {
	entityManager := game.ECSManager
	entityID, _ := strconv.Atoi(entityIDStr)

	reference := entityDescription.Reference

	pRCD := entityManager.GetComponentDataByName(uint64(entityID), "RENDER_COMPONENT").(*ecs.RenderComponentData)

	if pRCD.Text != nil {
		return
	}

	var fullImagePath string
	var imageName string
	var pTexture *sdl.Texture
	var pImage *sdl.Surface

	mainEntity := (*assetDescriptions)[reference]
	if !mainEntity.AnimatedByDefault {
		// Non-animated assets

		if mainEntity.Image != "" && mainEntity.ImagesBasePath != "" {
			// Image assets

			imageName = mainEntity.Image
			fullImagePath = mainEntity.ImagesBasePath + imageName

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
			pRCD.Image = pImage
		}
	}
}

func connectTextDataWithRenderAndAnimateComponentData(game *Game, entityIDStr string, entityDescription *EntityJSONConfig, assetDescriptions *map[string]*AssetJSONConfig) {
	entityManager := game.ECSManager
	entityID, _ := strconv.Atoi(entityIDStr)

	reference := entityDescription.Reference

	pRCD := entityManager.GetComponentDataByName(uint64(entityID), "RENDER_COMPONENT").(*ecs.RenderComponentData)

	if pRCD.Image != nil {
		return
	}

	mainEntity := (*assetDescriptions)[reference]

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

func TransformSystemSetInitialVals(g *Game) {
	// Get the entity config map keys that represent entity ranges
	lvlConfig := g.LvlDescription

	for el := g.ECSManager.EntityToComponentMap.Front(); el != nil; el = el.Next() {
		hasTransformComponent := g.ECSManager.HasNamedComponent(el.Value.([]uint16), "TRANSFORM_COMPONENT")

		if !hasTransformComponent {
			continue
		}

		pTCD := g.ECSManager.GetComponentDataByName(el.Key.(uint64), "TRANSFORM_COMPONENT").(*ecs.TransformComponentData)
		entityJSONConfig := lvlConfig.GetEntityDescription(el.Key.(uint64))

		if entityJSONConfig.SpreadAlong == "X" {
			pRCD := g.ECSManager.GetComponentDataByName(el.Key.(uint64), "RENDER_COMPONENT").(*ecs.RenderComponentData)
			firstEntity := lvlConfig.GetFirstEntityIDFromRange(el.Key.(uint64))
			pTCD.Posy = float64(entityJSONConfig.InitialPosY)
			pTCD.Posx = float64(entityJSONConfig.InitialPosX) + float64(pRCD.Image.W)*(float64(el.Key.(uint64))-float64(firstEntity))
		} else {
			pTCD.Posx = float64(entityJSONConfig.InitialPosX)
			pTCD.Posy = float64(entityJSONConfig.InitialPosY)
		}

		pTCD.FlipImg = false
		pTCD.IsJumping = false
		pTCD.Hspeed = 0

	}
}

func InitializeLevel(g *Game) {
	entityComponentMap := CreateEntityComponent(g.LvlDescription)
	g.ECSManager.EntityToComponentMap = nil
	g.ECSManager.EntityToComponentMap = entityComponentMap
	CreateLvlsEntityAndComponents(g, entityComponentMap)
	g.ECSManager.LinkComponentsWithProperDataStruct()
	LoadImagesAndTextures(g)
	TransformSystemSetInitialVals(g)
}