package game

import (
	"encoding/json"
	"io/ioutil"
)

type animation struct {
	SpritesheetAvailable bool   `json:"SpritesheetAvailable"`
	Spritesheet          string `json:"Spritesheet"`
	NumberAnimations     int8   `json:"NumberAnimations"`
	Image                string `json:"Image"`
	ImageBasePath        string `json:"ImageBasePath"`
}

type AssetJSONConfig struct {
	AnimatedByDefault        bool                   `json:"AnimatedByDefault"`
	ImagesBasePath           string                 `json:"ImagesBasePath"`
	Image                    string                 `json:"Image"`
	DefaultAnimationDuration int8                   `json:"DefaultAnimationDuration"`
	Animations               *map[string]*animation `json:"Animations"`
	FontSize                 uint8                  `json:"FontSize"`
	Text                     string                 `json:"Text"`
}

func LoadAssetDescriptions(Game *Game) {
	assetDescriptions := make(map[string]*AssetJSONConfig)

	data, readInErr := ioutil.ReadFile("./game/assets.json")

	if readInErr != nil {
		panic(readInErr)
	}

	unmarshalErr := json.Unmarshal(data, &assetDescriptions)

	if unmarshalErr != nil {
		panic(unmarshalErr)
	}

	if Game.AssetDescriptions != nil {
		Game.AssetDescriptions = nil
	}

	Game.AssetDescriptions = &assetDescriptions
}

