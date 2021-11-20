package ecs

import (
	"codeberg.org/alluneedistux/GoJumpRunShoot/statemachine"
	"github.com/elliotchance/orderedmap"
	"github.com/veandco/go-sdl2/sdl"
	"strconv"
	"strings"
)

/*
 My take on the Entity Component System

 Entities are just an ID, in our case of type uint64 integer.
 So entities are represented by a slice of uint64 integers.
 Thus, index and element values are always identical.

 Since the entity is ultimately just the index to identify the
 component slice, we don't even need a standalone entity slice.

 Components are also just an ID, in our case of type uint16 integer.
 To connect entities and components we just use the entity ID
 as a key in a map. The value is then a slice of uint16 integers
 representing the components.

 => ECSManager field entCmpMap map[uint64][]uint16

 Entities describe Objects in our world that have an ID
 Alone they are useless. Components also just are numbers,
 HOWEVER we just arbitrarily which Component number describe what
 behaviour. E.g. we just say number 7 will be a RenderComponent
 that gets connected to RenderComponentData that is holding
 paths to an image file, the loaded image, and the loaded texture.
 Systems than will iterate over the Entities, that are indices
 pointing to slices of Components that are connected to data
 and on that data systems will act!

 Components then need data that their systems can act upon.
 So, how do we connect components and component data?
 Once again, we use a map!

 The map's key will be a string in the form of "<entityID>-<componentID>"
 which is by design unique. The value then is the pointer to the ComponentData.

 => var entityComponentStringToComponentDataMap map[string]*ComponentData

 COMPONENTS:
 0 - Dummy component with no function
 1 - Real      (Real things in the world, as opposed to meta stuff like health bars etc.)
 2 - Alive     (Players and NPCs, as opposed to dead objects like boxes etc)
 3 - Collide   (Can collide, as opposed to things that do not care about collision)
 4 - Transform (Position + Rotate + Scale, as opposed to Static objects)
 5 - Gravity   (Is affected by laws of gravity, can jump, can fall)
 6 - Dynamic   (Can be transformed and changed in any way)
 7 - Render    (Can be rendered, so effectively it is visible)
 8 - Animate
 ...

 65534 is the maximum of allowed components.
 65535 is used as the value for NO COMPONENT CONNECTED YET.
 Using 65534 components is super unlikely.
 That is why I chose 65535 as a phony "nil" value for components.

 // The last piece of the puzzle: The systems

 Systems loop through the entity component map, and check whether the entities
 possess the component the system is specialized on:

 RenderSystem() looks for a RENDER_COMPONENT, TransformSystem for a TRANSFORM_COMPONENT and so on.
 Otherwise, the system will skip the current iteration or break completely.

 If the component was found, the systems may then access the respective
 ComponentData by using some key, function or method that is able to retrieve
 the correct ComponentData by somehow accessing the entityComponentStringToComponentDataMap.

 Since ComponentData.data is an empty interface,
 type assertion to the correct ComponentData (e.g. RenderComponentData)
 is always required.

 Generics in upcoming go 1.18 might solve that and might also simplify the handling of systems etc.
*/

type ComponentIDStorage struct {
	// we can get CURRENT_MAX_COMPONENTS from len(componentNameToIDMap)
	ComponentNameToIDMap *map[string]uint16
}

func NewComponentIDStorage() *ComponentIDStorage {
	componentNameToIDMap := make(map[string]uint16)
	componentNameToIDMap["DUMMY_COMPONENT"] = 0 // Should never be used
	componentNameToIDMap["REAL_COMPONENT"] = 1
	componentNameToIDMap["ACTIVE_CONTROL_COMPONENT"] = 2
	componentNameToIDMap["PASSIVE_CONTROL_COMPONENT"] = 3
	componentNameToIDMap["COLLIDE_COMPONENT"] = 4
	componentNameToIDMap["TRANSFORM_COMPONENT"] = 5
	componentNameToIDMap["GRAVITY_COMPONENT"] = 6
	componentNameToIDMap["DYNAMIC_COMPONENT"] = 7
	componentNameToIDMap["RENDER_COMPONENT"] = 8
	componentNameToIDMap["ANIMATE_COMPONENT"] = 9

	cmpIDVault := ComponentIDStorage{ComponentNameToIDMap: &componentNameToIDMap}

	return &cmpIDVault
}

type ComponentData struct {
	Data interface{}
}

type ECSManager struct {
	EntityToComponentMap                    *orderedmap.OrderedMap
	EntityComponentStringToComponentDataMap *map[string]*ComponentData
	ComponentIDStorage                      *ComponentIDStorage
	ComponentData                           *ComponentData
	Systems									[]System
}

func NewECSManager() *ECSManager {
	entityComponentStringToComponentDataMap := make(map[string]*ComponentData)

	ecsManager := ECSManager {
		EntityToComponentMap: orderedmap.NewOrderedMap(),
		EntityComponentStringToComponentDataMap: &entityComponentStringToComponentDataMap,
		ComponentIDStorage: NewComponentIDStorage(),
		ComponentData: nil,
		Systems: make([]System, 6, 8),
	}

	return &ecsManager
}

func (e *ECSManager) HasComponent(components []uint16, componentID uint16) bool {
	return components[componentID] != 65535
}

func (e *ECSManager) GetComponentDataByID(entityID uint64, componentID uint16) interface{} {
	entIDStr := strconv.Itoa(int(entityID))
	cmpIDStr := strconv.Itoa(int(componentID))
	entityComponentStringToComponentDataMap := e.EntityComponentStringToComponentDataMap

	componentKey := entIDStr + "-" + cmpIDStr
	return (*entityComponentStringToComponentDataMap)[componentKey].Data
}

func (e *ECSManager) GetComponentDataByName(entityID uint64, componentName string) interface{} {
	return e.GetComponentDataByID(entityID, e.GetComponentID(componentName))
}

func (e *ECSManager) GetComponentID(componentName string) uint16 {
	componentNameToIDMap := e.ComponentIDStorage.ComponentNameToIDMap
	return (*componentNameToIDMap)[componentName]
}

func (e *ECSManager) GetComponentIDAsStr(componentName string) string {
	return strconv.Itoa(int(e.GetComponentID(componentName)))
}

func (e *ECSManager) GetEntityComponentKey(entityID uint64, componentName string) string {
	return strconv.Itoa(int(entityID)) + "-" + e.GetComponentIDAsStr(componentName)
}


func (e *ECSManager) HasNamedComponent(components []uint16, componentName string) bool {
	return e.HasComponent(components, e.GetComponentID(componentName))
}

func (e *ECSManager) InitializeComponentsForEntity(entityID uint64) {
	componentNameToIDMap := e.ComponentIDStorage.ComponentNameToIDMap
	entityToComponentMapOrdered := *e.EntityToComponentMap
	CURRENT_MAX_COMPONENTS := len(*componentNameToIDMap)
	componentMap, _ := entityToComponentMapOrdered.Get(entityID)

	if componentMap == nil {
		// This is the first time we add a Component to the Entity since length
		// of the value component slice is 0
		entityToComponentMapOrdered.Set(entityID, make([]uint16, CURRENT_MAX_COMPONENTS, CURRENT_MAX_COMPONENTS))
		// Only initialize components to "NULL" values once
		for i := uint16(0); i < uint16(CURRENT_MAX_COMPONENTS); i++ {
			// 65535 means NO COMPONENT, our artificial NULL VALUE for components
			thisComponentMap, _ := entityToComponentMapOrdered.Get(entityID)
			thisComponentMap.([]uint16)[i] = 65535
		}
	}
}

func (e *ECSManager) AddComponentToEntity(entityID uint64, componentID uint16) {
	entityComponentStringToComponentDataMap := e.EntityToComponentMap
	componentMap, _ := entityComponentStringToComponentDataMap.Get(entityID)
	componentMap.([]uint16)[componentID] = componentID
}

func (e *ECSManager) LinkComponentsWithProperDataStruct() {
	entityComponentStringToComponentDataMap := *e.EntityComponentStringToComponentDataMap
	entityToComponentMap := *e.EntityToComponentMap

	curNumEntities := uint64(entityToComponentMap.Len())

	for i := uint64(0); i < curNumEntities; i++ {
		entityIDStr := strconv.Itoa(int(i))

		var componentIDStr string
		var keyForEntityComponentDataMap string
		componentMap, _ := entityToComponentMap.Get(i)

		for j, _ := range componentMap.([]uint16) {
			if j != 65535 {
				componentIDStr = strconv.Itoa(j)
			}

			keyForEntityComponentDataMap = entityIDStr + "-" + componentIDStr
			cd := ComponentData{Data: nil}

			// Only link data if entityComponentMap map's val for that key is nil
			// Also only link data if a real component exists.
			if keyForEntityComponentDataMap != "-" && j != 65535 {
				switch uint16(j) {
				case (*e.ComponentIDStorage.ComponentNameToIDMap)["TRANSFORM_COMPONENT"]:
					cd.Data = &TransformComponentData{}
				case (*e.ComponentIDStorage.ComponentNameToIDMap)["RENDER_COMPONENT"]:
					cd.Data = &RenderComponentData{}
				case (*e.ComponentIDStorage.ComponentNameToIDMap)["ANIMATE_COMPONENT"]:
					acd := make(map[string]*AnimationComponentDataCore)
					cd.Data = &AnimateComponentData{AnimationData: &acd}
				case (*e.ComponentIDStorage.ComponentNameToIDMap)["GRAVITY_COMPONENT"]:
					cd.Data = &GravityComponentData{}
				}
			}
			entityComponentStringToComponentDataMap[keyForEntityComponentDataMap] = &cd
		}
	}
}

func (e *ECSManager) GetEntityIDBoundariesFromEntityRange(entityIDStr string) *[2]uint64 {
	// Does not necessarily need to be a receiver function/method
	// See with time which choice is better
	if strings.Contains(entityIDStr, "-") {
		numKeys := strings.Split(entityIDStr, "-")
		numKeysUpperLimitStr := numKeys[1]
		numKeysLowerLimitStr := numKeys[0]
		numKeysUpperLimit, _ := strconv.ParseUint(numKeysUpperLimitStr, 10, 64)
		numKeysLowerLimit, _ := strconv.ParseUint(numKeysLowerLimitStr, 10, 64)

		return &[2]uint64{numKeysLowerLimit, numKeysUpperLimit}
	}
	return nil
}

func (e *ECSManager) GetEntityRect(entityID uint64) *sdl.Rect {
	pRCD := e.GetComponentDataByName(entityID, "RENDER_COMPONENT").(*RenderComponentData)
	pTCD := e.GetComponentDataByName(entityID, "TRANSFORM_COMPONENT").(*TransformComponentData)

	h := pRCD.Image.H
	w := pRCD.Image.W

	return &sdl.Rect{X: int32(pTCD.Posx), Y: int32(pTCD.Posy), W: w, H: h}
}

// Systems

type CommonSystemData struct {
	SystemID   uint16
	ECSManager *ECSManager
}

func NewCommonSystemData(componentName string, ecsManager *ECSManager) *CommonSystemData {
	return &CommonSystemData{
		SystemID: ecsManager.GetComponentID(componentName),
		ECSManager:     ecsManager,
	}
}

type System interface {
	Run(delta float64, statemachine *statemachine.StateMachine)
	UpdateComponent(float64, ...interface{})
}

func (bs *CommonSystemData) GetComponentData(entityID uint64) interface{} {
	componentID := bs.SystemID
	return bs.ECSManager.GetComponentDataByID(entityID, componentID)
}