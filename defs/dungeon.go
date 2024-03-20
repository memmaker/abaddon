package defs

import (
	"fmt"
	"math/rand"
	"rpg/dungen"
	"rpg/geometry"
	"rpg/gridmap"
	"strings"
)

type LevelName string

const (
	LevelNameTown     LevelName = "Town"
	LevelNameTownHell LevelName = "TownHell"
)

func IsTownLevel(levelName string) bool {
	return levelName == string(LevelNameTown) || levelName == string(LevelNameTownHell)
}

type SpecialLevelName string

const (
	SpecialLevelPoisonedWaterSupply SpecialLevelName = "!spec_poisoned_water_supply"
)

type BasicTileMapping struct {
	FloorTile       int32
	ChooseWallTile  func(mapPos geometry.Point) int32
	ChooseFloorTile func(mapPos geometry.Point) int32
	ClosedDoor      int32
	OpenDoor        int32
}

/*
Level 1 - Cathedral - None
Level 2 - Cathedral - Meat Jack
Level 3 - Catacombs - Poisoned Water supply
Level 4 - Catacombs - King Zedek's Tomb: Skeleton King Boss & Skeletons
Level 5 - Caves - The Weak demon
Level 6 - Caves - Minor Puzzles -> Unique Armor & Quest Item - MacGuffin
Level 7 - Hell - Chamber of Bone -> Room of magical Items & Guardian Spell
Level 8 - Hell - Minor Puzzles to open hidden passage -> Unique Amulet & Lucifer
Level 9 - Dungeon - Puzzle Room
Level 10 - Dungeon - Unique Boss Butchering Adventurers -> Unique Sword
Level 11 - Merchant District -
Level 12 - Merchant District - Unique Boss Shekel -> Unique Helmet
Level 13 - Town - Final Confrontation with King Kirioth -> Unique Item Drop & 4 magical weapons & 2 magical armors
*/
func getHellTilemap(randomSource *rand.Rand, layout *dungen.DungeonMap) BasicTileMapping {
	chooseFloorTile := func(mapPos geometry.Point) int32 {
		//roll := randomSource.Intn(20)
		return IconDungeonRedEarthFloor
	}
	chooseWallTile := func(mapPos geometry.Point) int32 {
		return IconDungeonRockyWall
	}

	return BasicTileMapping{
		FloorTile:       IconDungeonRedEarthFloor,
		ChooseWallTile:  chooseWallTile,
		ChooseFloorTile: chooseFloorTile,
		OpenDoor:        IconOpenDoor,
		ClosedDoor:      IconClosedDoor,
	}
}

func getCavesTilemap(randomSource *rand.Rand, layout *dungen.DungeonMap) BasicTileMapping {
	chooseFloorTile := func(mapPos geometry.Point) int32 {
		//roll := randomSource.Intn(20)
		return IconDirt
	}
	chooseWallTile := func(mapPos geometry.Point) int32 {
		return IconMountain
	}

	return BasicTileMapping{
		FloorTile:       IconDirt,
		ChooseWallTile:  chooseWallTile,
		ChooseFloorTile: chooseFloorTile,
		OpenDoor:        IconOpenDoor,
		ClosedDoor:      IconClosedDoor,
	}
}
func getCathedralTilemap(randomSource *rand.Rand, layout *dungen.DungeonMap) BasicTileMapping {
	chooseFloorTile := func(mapPos geometry.Point) int32 {
		roll := randomSource.Intn(20)
		if roll == 0 {
			brokenFloorTiles := []int32{65, 66, 67, 68}
			randomTile := brokenFloorTiles[randomSource.Intn(len(brokenFloorTiles))]
			return randomTile
		} else {
			return IconCathedralFloor
		}
	}
	chooseWallTile := func(mapPos geometry.Point) int32 {
		// if the tile below is a wall, we use the normal wall tile
		if !layout.Contains(mapPos.Add(geometry.Point{X: 0, Y: 1})) {
			return IconCathedralWallWithShading
		}
		tileBelow := layout.GetTile(mapPos.X, mapPos.Y+1)
		if tileBelow == dungen.Wall || tileBelow == dungen.Door {
			return IconCathedralWall
		}
		return IconCathedralWallWithShading
	}

	return BasicTileMapping{
		FloorTile:       IconCathedralFloor,
		ChooseWallTile:  chooseWallTile,
		ChooseFloorTile: chooseFloorTile,
		OpenDoor:        IconOpenDoor,
		ClosedDoor:      IconClosedDoor,
	}
}

func getCatacombsTilemap(randomSource *rand.Rand, layout *dungen.DungeonMap) BasicTileMapping {
	chooseFloorTile := func(mapPos geometry.Point) int32 {
		return IconCatacombsFloor
		roll := randomSource.Intn(20)
		if roll == 0 {
			brokenFloorTiles := []int32{65, 66, 67, 68}
			randomTile := brokenFloorTiles[randomSource.Intn(len(brokenFloorTiles))]
			return randomTile
		} else {
			return IconCathedralFloor
		}
	}
	chooseWallTile := func(mapPos geometry.Point) int32 {
		// if the tile below is a wall, we use the normal wall tile
		if !layout.Contains(mapPos.Add(geometry.Point{X: 0, Y: 1})) {
			return IconCatacombsWallWithShading
		}
		tileBelow := layout.GetTile(mapPos.X, mapPos.Y+1)
		if tileBelow == dungen.Wall || tileBelow == dungen.Door {
			return IconCatacombsWall
		}
		return IconCatacombsWallWithShading
	}

	return BasicTileMapping{
		FloorTile:       IconCatacombsFloor,
		ChooseWallTile:  chooseWallTile,
		ChooseFloorTile: chooseFloorTile,
		OpenDoor:        IconOpenDoor,
		ClosedDoor:      IconClosedDoor,
	}
}

func getModernDungeonOfDoomTilemap(randomSource *rand.Rand, layout *dungen.DungeonMap) BasicTileMapping {
	chooseFloorTile := func(mapPos geometry.Point) int32 {
		//roll := randomSource.Intn(20)
		return IconDungeonOfDoomFloorModern
	}
	chooseWallTile := func(mapPos geometry.Point) int32 {
		return IconDungeonOfDoomWallModern
	}

	return BasicTileMapping{
		FloorTile:       IconDungeonOfDoomFloorModern,
		ChooseWallTile:  chooseWallTile,
		ChooseFloorTile: chooseFloorTile,
		OpenDoor:        IconDungeonOfDoomDoorModern,
		ClosedDoor:      IconDungeonOfDoomDoorModern,
	}
}
func getDungeonOfDoomTilemap(randomSource *rand.Rand, layout *dungen.DungeonMap) BasicTileMapping {
	isCorridor := func(mapPos geometry.Point) bool {
		roomAt := layout.GetRoomAt(mapPos)
		if roomAt == nil {
			return false
		}
		if strings.Contains(roomAt.GetTemplateName(), "corridor") {
			return true
		}
		return false
	}

	isRoom := func(mapPos geometry.Point) bool {
		roomAt := layout.GetRoomAt(mapPos)
		if roomAt == nil {
			return false
		}
		if strings.Contains(roomAt.GetTemplateName(), "corridor") {
			return false
		}
		return true
	}

	chooseFloorTile := func(mapPos geometry.Point) int32 {
		//roll := randomSource.Intn(20)
		if isCorridor(mapPos) {
			return IconDungeonOfDoomCorridorFloor
		}
		return IconDungeonOfDoomFloor
	}
	chooseWallTile := func(mapPos geometry.Point) int32 {
		neighbors := layout.GetCardinalNeighbours(mapPos)
		for _, neighbor := range neighbors {
			if isRoom(neighbor) {
				direction := neighbor.Sub(mapPos)
				if direction.X == 0 {
					return IconDungeonOfDoomWallHoriz
				}
				return IconDungeonOfDoomWallVert
			}
		}
		return IconDungeonOfDoomCorridorWall
	}

	return BasicTileMapping{
		FloorTile:       IconDungeonOfDoomFloor,
		ChooseWallTile:  chooseWallTile,
		ChooseFloorTile: chooseFloorTile,
		OpenDoor:        IconDungeonOfDoomDoor,
		ClosedDoor:      IconDungeonOfDoomDoor,
	}
}

type TransitionDef struct {
	LocationName string
	Target       gridmap.Transition
	IsUp         bool
}

type LevelDefinition struct {
	TemplateRules                dungen.TemplateRules
	GetTilemapForBasicTiles      func(layout *dungen.DungeonMap, level int) BasicTileMapping
	DungeonLevel                 int
	Transitions                  []TransitionDef
	AddDefaultDungeonTransitions bool
	LevelName                    string
}

func GetLevelDefinitionByName(randomSource *rand.Rand, specialMapName SpecialLevelName) LevelDefinition {
	switch specialMapName {
	case SpecialLevelPoisonedWaterSupply:
		return getPoisonedWaterSupplyLevelDefinition(randomSource)
	}
	// fallback
	return GetLevelDefinition(randomSource, 1)
}

func getPoisonedWaterSupplyLevelDefinition(source *rand.Rand) LevelDefinition {
	var levelDef LevelDefinition
	dungeonLevel := 3
	levelDef.DungeonLevel = dungeonLevel
	rules := dungen.NewTemplateRules(map[int][]string{
		3: {
			"Cave_template_for_poisoned_water",
		},
	}, dungeonLevel, "Cave").WithNoDoors()
	levelDef.TemplateRules = rules
	levelDef.GetTilemapForBasicTiles = func(layout *dungen.DungeonMap, level int) BasicTileMapping {
		return getCavesTilemap(source, layout)
	}
	levelDef.Transitions = []TransitionDef{
		{
			LocationName: "from_catacombs",
			Target: gridmap.Transition{
				TargetMap:      LevelNameFromDepth(dungeonLevel),
				TargetLocation: "to_poisoned_water_supply",
			},
			IsUp: false,
		},
	}
	levelDef.AddDefaultDungeonTransitions = false
	levelDef.LevelName = "Poisoned Water Supply"
	return levelDef
}

func GetLevelDefinition(randomSource *rand.Rand, dungeonLevel int) LevelDefinition {
	//Hell Town -> Dungeons of Doom -> Caves -> Crypt -> Cathedral Town
	getTransitionsForLevel := map[int][]TransitionDef{
		1: {
			{
				LocationName: "ladder_down",
				Target: gridmap.Transition{
					TargetMap:      string(LevelNameTownHell),
					TargetLocation: "to_dungeon",
				},
				IsUp: false,
			},
		},
		5: {
			{
				LocationName: "ladder_up",
				Target: gridmap.Transition{
					TargetMap:      string(LevelNameTown),
					TargetLocation: "to_dungeon",
				},
				IsUp: true,
			},
		},
	}
	getRequiredRoomsForLevel := map[int][]string{
		1: {
			"Hell_template_for_lucifer",
		},
		2: {
			"Dung_template_for_rogue",
		},
		4: {
			"Cata_template_for_king_zedek",
		},
		5: {
			"Cath_template_for_meat_jack",
		},
	}
	getTemplatePrefixForLevel := func(level int) string {
		if level == 1  {
			return "Hell"
		} else if level == 2 {
			return "Dung" //return "Dung"
		} else if level == 3 {
			return "Cave"
		} else if level == 4 {
			return "Cata"
		} else if level == 5 {
			return "Cath"
		}
		return "Cath"
	}
	getLevelName := func(level int) string {
		if level == 1 {
			return fmt.Sprintf("Level %d\nHell", level)
		} else if level == 2 {
			return fmt.Sprintf("Level %d\nDungeon of Doom", level)
		} else if level == 3 {
			return fmt.Sprintf("Level %d\nCaves", level)
		} else if level == 4 {
			return fmt.Sprintf("Level %d\nCatacombs", level)
		} else if level == 5 {
			return fmt.Sprintf("Level %d\nCathedral", level)
		} else if level == 6 {
			return fmt.Sprintf("Level %d\nTown", level)
		}
		return fmt.Sprintf("Void Level %d", level)
	}
	getTilemapForLevel := func(layout *dungen.DungeonMap, level int) BasicTileMapping {
		if level == 1 {
			return getHellTilemap(randomSource, layout)
		} else if level == 2 {
			return getDungeonOfDoomTilemap(randomSource, layout)
		} else if level == 3 {
			return getCavesTilemap(randomSource, layout)
		} else if level == 4 {
			return getCatacombsTilemap(randomSource, layout)
		} else if level == 5 {
			return getCathedralTilemap(randomSource, layout)
		}
		return getCathedralTilemap(randomSource, layout)
	}
	rules := dungen.NewTemplateRules(getRequiredRoomsForLevel, dungeonLevel, getTemplatePrefixForLevel(dungeonLevel)).
		WithForbidden("Cave_template_for_poisoned_water")

	if dungeonLevel == 1 || dungeonLevel == 3 {
		rules = rules.WithNoDoors()
	}
	levelDef := LevelDefinition{
		DungeonLevel:                 dungeonLevel,
		TemplateRules:                rules,
		GetTilemapForBasicTiles:      getTilemapForLevel,
		Transitions:                  getTransitionsForLevel[dungeonLevel],
		AddDefaultDungeonTransitions: true,
		LevelName:                    getLevelName(dungeonLevel),
	}
	return levelDef
}

func LevelNameFromDepth(level int) string {
	//!gen_dungeon_level_1
	return fmt.Sprintf("!gen_dungeon_level_%d", level)
}
