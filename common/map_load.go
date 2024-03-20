package common

import (
	"path"
	"rpg/defs"
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/ldtk"
	"rpg/renderer"
	"rpg/util"
	"strings"
)

func (s *State) NewGridMapFromLDTK(mapName string) (*gridmap.GridMap[*Actor, *Item, Object], renderer.TextureAtlas) {
	ldtkMapProject := s.ldtkProject
	worldTileset := ldtkMapProject.TilesetByIdentifier("World")

	currentMap := ldtkMapProject.LevelByIdentifier(mapName)
	locations := loadLocations(currentMap)

	mapDisplayName := currentMap.PropertyByIdentifier("DisplayName").AsString()

	environmentLayer := currentMap.LayerByIdentifier("Tiles")
	//metaLayer := currentMap.LayerByIdentifier("Meta")
	npcLayer := currentMap.LayerByIdentifier("Entities")

	loadedMap := gridmap.NewEmptyMap[*Actor, *Item, Object](environmentLayer.CellWidth, environmentLayer.CellHeight)
	loadedMap.SetName(mapName)
	loadedMap.SetDisplayName(mapDisplayName)
	loadedMap.SetNamedLocations(locations)

	for _, tile := range environmentLayer.Tiles {
		pos, tileType := MapTileFromLDTK(worldTileset, environmentLayer, tile)
		loadedMap.SetTile(pos, tileType)
	}

	for _, entity := range npcLayer.Entities {
		typeOfEntity := entity.Identifier
		posX, posY := npcLayer.ToGridPosition(entity.Position[0], entity.Position[1])
		pos := geometry.Point{X: posX, Y: posY}
		if typeOfEntity == "NPC" {
			s.addActorFromEntity(ldtkMapProject, loadedMap, entity, pos)
		} else if typeOfEntity == "Sign" {
			signText := entity.PropertyByIdentifier("Text").AsString()
			textLines := strings.Split(signText, "\n")
			s.addSignFromEntity(loadedMap, textLines, pos)
		} else if typeOfEntity == "Door" {
			loadedMap.AddObject(NewDoor(defs.IconOpenDoor, defs.IconClosedDoor), pos)
		} else if typeOfEntity == "Barrel" {
			loadedMap.AddObject(s.factory.NewBarrel(), pos)
		}

	}

	textureAtlas := renderer.NewTextureAtlas(path.Join("assets", worldTileset.Path), worldTileset.GridSize, worldTileset.GridSize)
	return loadedMap, textureAtlas
}

func (s *State) addSignFromEntity(loadedMap *gridmap.GridMap[*Actor, *Item, Object], inscription []string, pos geometry.Point) {
	sign := s.factory.NewSign(inscription)
	loadedMap.AddObject(sign, pos)
}

func  (s *State) addActorFromEntity(ldtkMapProject *ldtk.Project, loadedMap *gridmap.GridMap[*Actor, *Item, Object], entity *ldtk.Entity, pos geometry.Point) {
	var newActor *Actor
	if !entity.PropertyByIdentifier("UniqueName").IsNull() {
		uniqueName := entity.PropertyByIdentifier("UniqueName").AsString()
		newActor = s.factory.NewUniqueEnemyFromName(uniqueName)
	} else {
		newActor = actorFromEntity(ldtkMapProject, entity, pos)
	}
	loadedMap.AddActor(newActor, pos)
}

func actorFromEntity(ldtkMapProject *ldtk.Project, entity *ldtk.Entity, pos geometry.Point) *Actor {
	icon, _ := resolveLDTKIcon(ldtkMapProject, entity.PropertyByIdentifier("Tile"))

	flipLook := entity.PropertyByIdentifier("FlipLookDirection").AsBool()

	actorName := entity.PropertyByIdentifier("Name").AsString()
	actor := NewActor(icon)
	actor.SetName(actorName)
	asciiIcon := entity.PropertyByIdentifier("ASCII").AsString()
	ascii := int32(asciiIcon[0])
	actor.SetAsciiIcon(ascii)
	actor.SetCP437Icon(int32(util.UnicodeToCP437Byte(ascii)))
	actor.SetUnicodeIcon(icon)

	actor.SetColor(defs.White)

	actor.SetDrawPosition(pos.ToPointF())

	if !entity.PropertyByIdentifier("InternalName").IsNull() {
		actor.SetInternalName(entity.PropertyByIdentifier("InternalName").AsString())
	}

	npcTypeProp := entity.PropertyByIdentifier("Type").AsString()
	if strings.HasPrefix(npcTypeProp, "Store") {
		actor.SetVendorType(VendorType(strings.ToLower(npcTypeProp)))
		actor.SetFaction(defs.FactionNeutral)
		actor.SetNoCorpse()
		actor.SetDeathStringFormats([]string{
			"%s teleported away at the last second.",
		})
	} else if strings.HasPrefix(npcTypeProp, "Neutral") {
		actor.SetNeutralFaction()
	} else if strings.HasPrefix(npcTypeProp, "Aggressive") {
		actor.SetMonsterFaction()
	}
	actor.SetFlipLook(flipLook)
	return actor
}

func resolveLDTKIcon(ldtkMapProject *ldtk.Project, prop *ldtk.Property) (int32, ldtk.EnumSet) {
	tileData := prop.Value.(map[string]interface{})
	tileset := ldtkMapProject.TilesetByUID(int(tileData["tilesetUid"].(float64)))
	tilesetWidth := tileset.Width / tileset.GridSize
	atlasX := int(tileData["x"].(float64) / float64(tileset.GridSize))
	atlasY := int(tileData["y"].(float64) / float64(tileset.GridSize))
	textureIndex := atlasY*tilesetWidth + atlasX
	enumsForTile := tileset.EnumsForTile(textureIndex)
	return int32(textureIndex), enumsForTile
}
func loadLocations(level *ldtk.Level) []gridmap.NamedLocation {
	locations := make([]gridmap.NamedLocation, 0)
	metaLayer := level.LayerByIdentifier("Meta")
	for _, entity := range metaLayer.Entities {
		if entity.Identifier != "Location" {
			continue
		}
		nameOfLocation := entity.PropertyByIdentifier("Name").AsString()
		posX, posY := metaLayer.ToGridPosition(entity.Position[0], entity.Position[1])
		pos := geometry.Point{X: posX, Y: posY}
		if nameOfLocation != "" {
			locations = append(locations, gridmap.NamedLocation{
				LocationName: nameOfLocation,
				Pos:          pos,
			})
		}
	}
	return locations
}
