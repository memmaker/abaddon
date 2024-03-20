package common

import (
	"cmp"
	"fmt"
	"math"
	"math/rand"
	"rpg/defs"
	"rpg/dungen"
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/ldtk"
	"rpg/util"
	"slices"
	"sort"
	"strconv"
	"strings"
)

type MapGenerator struct {
	randomSource *rand.Rand
	factory      *Factory
	ldtkProject  *ldtk.Project
}

const MAX_LEVELS = 12

func NewMapGenerator(randomSource *rand.Rand) *MapGenerator {
	return &MapGenerator{
		randomSource: randomSource,
	}
}
func (g *MapGenerator) SetFactory(factory *Factory) {
	g.factory = factory
}

func (g *MapGenerator) SetLdtkProject(project *ldtk.Project) {
	g.ldtkProject = project
}
func (g *MapGenerator) GenerateMap(targetMap string, mapTemplates *ldtk.Project, dungeonLevel int) *gridmap.GridMap[*Actor, *Item, Object] {
	var dungeonDef defs.LevelDefinition
	if strings.HasPrefix(targetMap, "!spec_") {
		dungeonDef = defs.GetLevelDefinitionByName(g.randomSource, defs.SpecialLevelName(targetMap))
	} else {
		dungeonDef = defs.GetLevelDefinition(g.randomSource, dungeonLevel)
	}

	var generatedLayout *dungen.DungeonMap
	generatedLayout = g.createStoryLayout(mapTemplates, dungeonDef.TemplateRules)

	//generatedLayout.Print()
	genWidth, genHeight := generatedLayout.GetSize()
	emptyMap := gridmap.NewEmptyMap[*Actor, *Item, Object](genWidth, genHeight)

	emptyMap.SetName(targetMap)
	emptyMap.SetMetaString(generatedLayout.GetMetaInfo())

	g.addBasicTiles(emptyMap, generatedLayout, dungeonDef.GetTilemapForBasicTiles(generatedLayout, dungeonLevel))

	transitionLocations, remainingStairLocations := g.addStairs(emptyMap, generatedLayout, dungeonDef)
	emptyMap.AddToMetaInfo(fmt.Sprintf("remaining stair locations: %d", len(remainingStairLocations)))

	poiPositions := g.decorateRooms(emptyMap, generatedLayout, dungeonLevel)

	g.addHazards(emptyMap, generatedLayout, poiPositions, dungeonLevel)

	randomSource := g.randomSource
	enemySpawnPos := make(map[geometry.Point]bool)
	generatedLayout.TraverseTilesRandomly(randomSource, func(pos geometry.Point) {
		if generatedLayout.IsEmptySpace(pos) {
			enemySpawnPos[pos] = true
		}
	})

	keys := getSortedKeys(enemySpawnPos)
	randomIndices := randomSource.Perm(len(keys))

	for _, index := range randomIndices {
		pos := keys[index]
		// 5% chance to spawn an enemy
		distToStairs := minDist(transitionLocations, pos)
		rollForEnemy := randomSource.Intn(100)
		//println(fmt.Sprintf("pos: %v, rollForEnemy: %d, distToEntrance: %d", pos, rollForEnemy, distToEntrance))
		if rollForEnemy < 5 && distToStairs > 7 && emptyMap.CanPlaceActorHere(pos) {
			enemy, companions := g.factory.NewRandomEnemy(dungeonLevel)
			enemy.SetDrawPosition(pos.ToPointF())
			emptyMap.AddActor(enemy, pos)
			g.spawnCompanions(emptyMap, enemy, companions, dungeonLevel)
		} else if randomSource.Intn(100) < 1 && distToStairs > 7 && emptyMap.CanPlaceItemHere(pos) {
			lootItem := g.factory.GenerateLoot(dungeonLevel)
			if lootItem != nil {
				emptyMap.AddItem(lootItem, pos)
			}
		}
	}

	emptyMap.SetDisplayName(dungeonDef.LevelName)
	//emptyMap.Print()

	return emptyMap
}

func minDist(locations []geometry.Point, pos geometry.Point) float64 {
	minimumDist := math.MaxFloat64
	for _, loc := range locations {
		dist := float64(loc.ChebyshevDistanceTo(pos))
		if dist < minimumDist {
			minimumDist = dist
		}
	}
	return minimumDist
}

func (g *MapGenerator) addShrine(emptyMap *gridmap.GridMap[*Actor, *Item, Object], poiPositions []geometry.Point, dungeonLevel int) {
	if len(poiPositions) == 0 {
		return
	}
	randomIndex := g.randomSource.Intn(len(poiPositions))
	poiPos := poiPositions[randomIndex]
	g.addRandomShrine(g.randomSource, emptyMap, dungeonLevel, poiPos)
}
func (g *MapGenerator) addRandomShrine(source *rand.Rand, emptyMap *gridmap.GridMap[*Actor, *Item, Object], dungeonLevel int, pos geometry.Point) {
	actionForShrine := g.factory.GetRandomUntargetedAction(dungeonLevel, defs.ForShrine)

	newShrine := NewBumpObject("a shrine", defs.IconShrine, func(actor *Actor, startAnimating func() bool) {
		actionForShrine.Invoke(actor, "shrine")
	})
	emptyMap.AddObject(newShrine, pos)
	shrineName := fmt.Sprintf("Shrine of %s", actionForShrine.Name())
	emptyMap.AddToMetaInfo(shrineName)
}
func (g *MapGenerator) decorateRooms(emptyMap *gridmap.GridMap[*Actor, *Item, Object], generatedLayout *dungen.DungeonMap, dungeonLevel int) []geometry.Point {
	maxLitPercentage := 0.8
	litRoomPercentage := util.Clamp(0, maxLitPercentage, 1.0-(float64(dungeonLevel)/float64(MAX_LEVELS)))

	allRooms := generatedLayout.AllRooms()
	roomCount := len(allRooms)
	randomIndices := g.randomSource.Perm(roomCount)

	var poiPositions []geometry.Point
	roomsWeLit := 0
	roomsWeWantLit := int(math.Ceil(float64(roomCount) * litRoomPercentage))
	for _, randomIndex := range randomIndices {
		room := allRooms[randomIndex]
		roomPoIs := room.GetPOIPositions()
		if len(roomPoIs) > 0 {
			poiPositions = append(poiPositions, roomPoIs...)
		}
		if room.HasDecoratedTemplate() {
			decoLevel := room.GetDecorationLevelName()
			deco := g.getDecorationLevel(decoLevel)
			g.decorateRoomFromTemplate(emptyMap, deco, room)
			emptyMap.AddToMetaInfo(fmt.Sprintf("DecoRoom: %s", deco.Identifier))
		}
		if roomsWeLit < roomsWeWantLit { // && room.CanBeLit() { TODO: maybe we'll add this in again later
			roomsWeLit++
			for _, pos := range room.GetAbsoluteRoomTiles() {
				emptyMap.SetLit(pos, true)
				for _, neighbour := range emptyMap.GetFilteredNeighbors(pos, func(pos geometry.Point) bool {
					return emptyMap.Contains(pos)
				}) {
					emptyMap.SetLit(neighbour, true)
				}
			}
		}
	}
	return poiPositions
}
func (g *MapGenerator) decorateRoomFromTemplate(emptyMap *gridmap.GridMap[*Actor, *Item, Object], template *ldtk.Level, room *dungen.DungeonRoom) {
	tileLayer := template.LayerByIdentifier("Tiles")
	tileSet := g.ldtkProject.TilesetByIdentifier("World")
	for _, tile := range tileLayer.Tiles {
		if tile.ID < 0 || tile.ID > 223 {
			continue
		}
		relativeMapPos, mapTile := MapTileFromLDTK(tileSet, tileLayer, tile)
		position := room.ToAbsolutePosition(relativeMapPos)
		if emptyMap.IsTileWalkable(position) && room.Contains(position) {
			emptyMap.SetTile(position, mapTile)
		}
	}

	entityLayer := template.LayerByIdentifier("Entities")
	for _, entity := range entityLayer.Entities {
		posX, posY := entityLayer.ToGridPosition(entity.Position[0], entity.Position[1])
		relativeMapPos := geometry.Point{X: posX, Y: posY}
		mapPos := room.ToAbsolutePosition(relativeMapPos)

		if entity.Identifier != "NPC" { // only unique NPCs for now
			continue
		}
		if emptyMap.IsCurrentlyPassable(mapPos) {
			if !entity.PropertyByIdentifier("UniqueName").IsNull() {
				uniqueName := entity.PropertyByIdentifier("UniqueName").AsString()
				uniqueActor := g.factory.NewUniqueEnemyFromName(uniqueName)
				emptyMap.AddActor(uniqueActor, mapPos)
				uniqueActor.SetDrawPosition(mapPos.ToPointF())
			}
		}

	}
}

func (g *MapGenerator) addStairs(emptyMap *gridmap.GridMap[*Actor, *Item, Object], generatedLayout *dungen.DungeonMap, levelDef defs.LevelDefinition) ([]geometry.Point, []geometry.Point) {
	randomSource := g.randomSource
	ladderUpTile := gridmap.Tile{
		DefinedIcon:        defs.IconStairsUp,
		DefinedDescription: "a ladder leading up",
		IsWalkable:         true,
		IsTransparent:      true,
		Special:            gridmap.SpecialTileStairsUp,
	}
	ladderDownTile := gridmap.Tile{
		DefinedIcon:        defs.IconStairsDown,
		DefinedDescription: "a ladder leading down",
		IsWalkable:         true,
		IsTransparent:      true,
		Special:            gridmap.SpecialTileStairsDown,
	}

	stairPositions := generatedLayout.GetPotentialStairPositions()
	var transitionLocations []geometry.Point
	if levelDef.AddDefaultDungeonTransitions {
		if len(stairPositions) < 2 {
			panic("not enough stair positions") // TODO: do something more sensible
		}

		entranceIndex := randomSource.Intn(len(stairPositions))
		entrance := stairPositions[entranceIndex]
		// remove the entrance from the list of possible exits
		stairPositions = append(stairPositions[:entranceIndex], stairPositions[entranceIndex+1:]...)

		slices.SortStableFunc(stairPositions, func(i, j geometry.Point) int {
			return cmp.Compare(i.ManhattanDistanceTo(entrance), j.ManhattanDistanceTo(entrance))
		})

		var exitIndex int
		// 50/50 chance for random entrance and exit
		if randomSource.Intn(2) == 0 { // largest distance
			exitIndex = len(stairPositions) - 1
		} else { // random pairs, but not the same
			exitIndex = randomSource.Intn(len(stairPositions))
		}
		exit := stairPositions[exitIndex]
		// remove the exit from the list of possible exits
		stairPositions = append(stairPositions[:exitIndex], stairPositions[exitIndex+1:]...)


		dungeonLevel := levelDef.DungeonLevel

		if dungeonLevel > 0 && dungeonLevel < 5 { // add ladder up
			mapEntryPosition := entrance
			emptyMap.SetTile(mapEntryPosition, ladderUpTile)
			emptyMap.AddNamedLocation("ladder_up", mapEntryPosition)
			emptyMap.AddTransitionAt(mapEntryPosition, gridmap.Transition{
				TargetMap:      defs.LevelNameFromDepth(dungeonLevel + 1),
				TargetLocation: "ladder_down",
			})
			transitionLocations = append(transitionLocations, mapEntryPosition)
		}
		// 6 = Town
		// 5 = Cathedral (down only)
		// 4 = Catacombs
		// 3 = Caves
		// 2 = Dungeon
		// 1 = Hell (up only)
		// 0 = Hell Town
		if dungeonLevel > 1 && dungeonLevel < 6 { // add ladder down
			mapExitPosition := exit
			emptyMap.SetTile(mapExitPosition, ladderDownTile)
			emptyMap.AddNamedLocation("ladder_down", mapExitPosition)

			emptyMap.AddTransitionAt(mapExitPosition, gridmap.Transition{
				TargetMap:      defs.LevelNameFromDepth(dungeonLevel - 1),
				TargetLocation: "ladder_up",
			})
			transitionLocations = append(transitionLocations, mapExitPosition)
		}


		slices.SortStableFunc(stairPositions, func(i, j geometry.Point) int {
			iDist := i.ManhattanDistanceTo(entrance) + i.ManhattanDistanceTo(exit)
			jDist := j.ManhattanDistanceTo(entrance) + j.ManhattanDistanceTo(exit)
			return cmp.Compare(iDist, jDist)
		})
	}

	if len(levelDef.Transitions) > 0 {
		if len(levelDef.Transitions) > len(stairPositions) {
			println(fmt.Sprintf("ERR: not enough stair positions for all transitions on map %s", emptyMap.GetName()))
		} else {
			for _, additionalTransition := range levelDef.Transitions {
				//pop pos
				furthestIndex := len(stairPositions) - 1
				stairPos := stairPositions[furthestIndex]
				stairPositions = stairPositions[:furthestIndex]
				locationName := additionalTransition.LocationName
				transitionTo := additionalTransition.Target
				var tile gridmap.Tile
				if additionalTransition.IsUp {
					tile = ladderUpTile
				} else {
					tile = ladderDownTile
				}
				emptyMap.SetTile(stairPos, tile)
				emptyMap.AddNamedLocation(locationName, stairPos)
				emptyMap.AddTransitionAt(stairPos, transitionTo)
				transitionLocations = append(transitionLocations, stairPos)
			}
		}

	}

	return transitionLocations, stairPositions
}

func (g *MapGenerator) addBasicTiles(emptyMap *gridmap.GridMap[*Actor, *Item, Object], layout *dungen.DungeonMap, tileMap defs.BasicTileMapping) []geometry.Point {
	wallTileFromIcon := func(icon int32) gridmap.Tile {
		return gridmap.Tile{
			DefinedIcon:        icon,
			DefinedDescription: "a wall",
			IsWalkable:         false,
			IsTransparent:      false,
		}
	}
	floorTileFromIcon := func(icon int32) gridmap.Tile {
		return gridmap.Tile{
			DefinedIcon:        icon,
			DefinedDescription: "a floor",
			IsWalkable:         true,
			IsTransparent:      true,
		}
	}

	sizeX, sizeY := layout.GetSize()
	var allDoorPositions []geometry.Point
	for x := 0; x < sizeX; x++ {
		for y := 0; y < sizeY; y++ {
			mapPos := geometry.Point{X: x, Y: y}
			switch layout.GetTile(x, y) {
			case dungen.Wall:
				emptyMap.SetTile(mapPos, wallTileFromIcon(tileMap.ChooseWallTile(mapPos)))
			case dungen.Door:
				emptyMap.SetTile(mapPos, floorTileFromIcon(tileMap.FloorTile))
				emptyMap.AddObject(NewDoor(tileMap.OpenDoor, tileMap.ClosedDoor), mapPos)
				allDoorPositions = append(allDoorPositions, mapPos)
			case dungen.Room:
				fallthrough
			case dungen.Corridor:
				fallthrough
			default:
				emptyMap.SetTile(mapPos, floorTileFromIcon(tileMap.ChooseFloorTile(mapPos)))
			}
		}
	}
	return allDoorPositions
}

func (g *MapGenerator) createStoryLayout(mapTemplates *ldtk.Project, levelRules dungen.TemplateRules) *dungen.DungeonMap {
	randomSource := g.randomSource

	scratchSpaceWidth := 40
	scratchSpaceHeight := 40
	dunGen := dungen.NewAccretionGenerator(randomSource)
	dunGen.LoadTemplatesFrom(mapTemplates, levelRules.TemplatePrefix)
	dunGen.SetLevelRules(levelRules)

	//dunGen.SetDebugSteps(true)
	dunGen.SetRoomCount(16)
	generatedLayout := dunGen.Generate(scratchSpaceWidth, scratchSpaceHeight)
	for len(generatedLayout.AllRooms()) < 10 {
		generatedLayout = dunGen.Generate(scratchSpaceWidth, scratchSpaceHeight)
	}
	return generatedLayout
}

func (g *MapGenerator) createMegaDungeonLayout() *dungen.DungeonMap {
	randomSource := g.randomSource

	scratchSpaceWidth := 60
	scratchSpaceHeight := 60
	dunGen := dungen.NewMegaDungeonGenerator(randomSource)
	generatedLayout := dunGen.Generate(scratchSpaceWidth, scratchSpaceHeight)
	for len(generatedLayout.AllRooms()) < 10 {
		generatedLayout = dunGen.Generate(scratchSpaceWidth, scratchSpaceHeight)
	}
	return generatedLayout
}

func (g *MapGenerator) getDecorationLevel(decoLevel string) *ldtk.Level {
	var validLevels []*ldtk.Level
	for _, level := range g.ldtkProject.Levels {
		if strings.HasPrefix(level.Identifier, decoLevel) {
			validLevels = append(validLevels, level)
		}
	}
	if len(validLevels) == 0 {
		panic("no valid decoration levels found")
	}
	randomDeco := validLevels[g.randomSource.Intn(len(validLevels))]
	println("deco level: " + randomDeco.Identifier)
	return randomDeco
}

func getSortedKeys(pos map[geometry.Point]bool) []geometry.Point {
	keys := make([]geometry.Point, len(pos))
	i := 0
	for k := range pos {
		keys[i] = k
		i++
	}
	sort.SliceStable(keys, func(i, j int) bool {
		return keys[i].X < keys[j].X || (keys[i].X == keys[j].X && keys[i].Y < keys[j].Y)
	})
	return keys
}

func LevelFromName(targetMap string) int {
	nameParts := strings.Split(targetMap, "_")
	level, _ := strconv.Atoi(nameParts[len(nameParts)-1])
	return level
}

func (g *MapGenerator) spawnCompanions(emptyMap *gridmap.GridMap[*Actor, *Item, Object], leader *Actor, companions []defs.EnemyDefinition, dungeonLevel int) {
	origin := leader.MapPosition()
	companyCount := len(companions)

	freePositions := emptyMap.GetFreeCellsForDistribution(origin, companyCount, func(pos geometry.Point) bool {
		return emptyMap.CanPlaceActorHere(pos)
	})
	spawnCount := min(len(freePositions), companyCount)
	for i := 0; i < spawnCount; i++ {
		spawnPos := freePositions[i]
		companion := g.factory.NewEnemyFromDefinition(companions[i], dungeonLevel)
		emptyMap.AddActor(companion, spawnPos)
	}
}

func (g *MapGenerator) addHazards(emptyMap *gridmap.GridMap[*Actor, *Item, Object], layout *dungen.DungeonMap, poiPositions []geometry.Point, level int) {
	// water, lava, chasms, spikes, barrels, oil
	addHazards := []func(gridMap *gridmap.GridMap[*Actor, *Item, Object], room *dungen.DungeonRoom) {
		g.addWater,
		g.addLava,
		g.addSpikes,
		func(gridMap *gridmap.GridMap[*Actor, *Item, Object], room *dungen.DungeonRoom) {g.addHazardObjects(gridMap, room, g.factory.NewBarrel)},
		//g.addOil,
	}
	if level > 1 { // only add chasms after level 1
		addHazards = append(addHazards, g.addChasms)
	}
	for _, room := range layout.AllRooms() {
		// how many hazards?
		// which hazards?
		if g.randomSource.Intn(3) != 0 { // 33% chance to contain hazards
			continue
		}
		hazardCount := g.randomSource.Intn(len(addHazards))
		randomIndices := g.randomSource.Perm(len(addHazards))
		for i := 0; i < hazardCount; i++ {
			addHazards[randomIndices[i]](emptyMap, room)
		}
	}
}

func (g *MapGenerator) addHazardObjects(gridMap *gridmap.GridMap[*Actor, *Item, Object], room *dungen.DungeonRoom, createObject func() Object) {
	barrelCount := g.randomSource.Intn(5) + 1
	barrelsPlaced := 0
	roomTiles := room.GetAbsoluteRoomTiles()
	randomOrder := g.randomSource.Perm(len(roomTiles))
	for _, index := range randomOrder {
		roomPos := roomTiles[index]
		if room.IsConnectorPosition(roomPos) || gridMap.IsNextToTileWithSpecial(roomPos, gridmap.SpecialTileStairsDown) && gridMap.IsNextToTileWithSpecial(roomPos, gridmap.SpecialTileStairsUp) {
			continue
		}
		if gridMap.IsEmptyNonSpecialFloor(roomPos) {
			hazardObject := createObject()
			barrelsPlaced++
			gridMap.AddObject(hazardObject, roomPos)
			if barrelsPlaced >= barrelCount {
				break
			}
		}
	}
}

func (g *MapGenerator) addSpikes(gridMap *gridmap.GridMap[*Actor, *Item, Object], room *dungen.DungeonRoom) {
	startPos, found := getRandomEmptyRoomPos(gridMap, room)
	if !found {
		return
	}
	size := g.randomSource.Intn(5) + 1
	lavaLocations := gridMap.GetDijkstraMap(startPos, size, func(pos geometry.Point) bool {
		return gridMap.IsEmptyNonSpecialFloor(pos)
	})

	for pos, _ := range lavaLocations {
		gridMap.SetTile(pos, defs.SpikeTile)
	}
}

func (g *MapGenerator) addChasms(gridMap *gridmap.GridMap[*Actor, *Item, Object], room *dungen.DungeonRoom) {
	startPos, found := getRandomEmptyRoomPos(gridMap, room)
	if !found {
		return
	}
	size := g.randomSource.Intn(5) + 1
	chasmLocs := gridMap.GetDijkstraMap(startPos, size, func(pos geometry.Point) bool {
		return gridMap.IsEmptyNonSpecialFloor(pos) && !gridMap.IsNextToTileWithSpecial(pos, gridmap.SpecialTileStairsDown) && !gridMap.IsNextToTileWithSpecial(pos, gridmap.SpecialTileStairsUp)
	})

	for pos, _ := range chasmLocs {
		gridMap.SetTile(pos, defs.ChasmTile)
	}
}
func (g *MapGenerator) addLava(gridMap *gridmap.GridMap[*Actor, *Item, Object], room *dungen.DungeonRoom) {
	startPos, found := getRandomEmptyRoomPos(gridMap, room)
	if !found {
		return
	}
	size := g.randomSource.Intn(5) + 1
	lavaLocations := gridMap.GetDijkstraMap(startPos, size, func(pos geometry.Point) bool {
		return gridMap.IsEmptyNonSpecialFloor(pos)
	})

	for pos, _ := range lavaLocations {
		gridMap.SetTile(pos, defs.LavaTile)
	}
}

func (g *MapGenerator) addWater(gridMap *gridmap.GridMap[*Actor, *Item, Object], room *dungen.DungeonRoom) {
	startPos, found := getRandomEmptyRoomPos(gridMap, room)
	if !found {
		return
	}
	size := g.randomSource.Intn(5) + 1
	lavaLocations := gridMap.GetDijkstraMap(startPos, size, func(pos geometry.Point) bool {
		return gridMap.IsEmptyNonSpecialFloor(pos)
	})

	for pos, _ := range lavaLocations {
		gridMap.SetTile(pos, defs.WaterTile)
	}
}

func getRandomEmptyRoomPos(gridMap *gridmap.GridMap[*Actor, *Item, Object], room *dungen.DungeonRoom) (emptyPos geometry.Point, found bool) {
	roomTiles := room.GetAbsoluteRoomTiles()
	randomOrder := rand.Perm(len(roomTiles))
	for _, index := range randomOrder {
		roomPos := roomTiles[index]
		if room.IsConnectorPosition(roomPos) {
			continue
		}
		if !gridMap.IsEmptyNonSpecialFloor(roomPos) {
			continue
		}
		if gridMap.IsNextToTileWithSpecial(roomPos, gridmap.SpecialTileStairsUp) || gridMap.IsNextToTileWithSpecial(roomPos, gridmap.SpecialTileStairsDown) {
			continue
		}
		return roomPos, true
	}
	return geometry.Point{}, false
}

func MapTileFromLDTK(tileSet *ldtk.Tileset, layer *ldtk.Layer, tile *ldtk.Tile) (geometry.Point, gridmap.Tile) {
	posX, posY := layer.ToGridPosition(tile.Position[0], tile.Position[1])
	pos := geometry.Point{X: posX, Y: posY}
	enums := tileSet.EnumsForTile(tile.ID)

	walkable := !enums.Contains("IsBlockingMovement")
	transparent := !enums.Contains("IsBlockingView")
	isSpikes := enums.Contains("IsSpikes")
	isWater := enums.Contains("IsWater")
	isChasm := enums.Contains("IsChasm")
	isLava := enums.Contains("IsLava")
	specialTile := gridmap.SpecialTileNone

	if isWater {
		specialTile = gridmap.SpecialTileWater
	} else if isSpikes {
		specialTile = gridmap.SpecialTileSpikes
	} else if isChasm {
		specialTile = gridmap.SpecialTileChasm
	} else if isLava {
		specialTile = gridmap.SpecialTileLava
	}

	tileType := gridmap.Tile{
		DefinedIcon:   int32(tile.ID),
		IsWalkable:    walkable,
		IsTransparent: transparent,
		Special:       specialTile,
	}
	return pos, tileType
}
