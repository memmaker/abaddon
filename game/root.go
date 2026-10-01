package game

import (
	"cmp"
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"os"
	"path"
	"rpg/audio"
	"rpg/common"
	"rpg/defs"
	"rpg/files"
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/ldtk"
	"rpg/stats"
	"rpg/util"
	"slices"
	"strings"
	"time"
)


type GameInstance struct {
	// base infos / config
	title                       string
	deviceIndependentScreenSize geometry.Point
	configuration               common.Configuration
	startInfo                   common.StartInfo

	// dependencies
	uiController common.UserInterface
	audioPlayer  *audio.Player

	// world state
	state        common.State
	actionBroker *common.Broker // registered hooks

	// random
	randomSource *rand.Rand
	worldTicks   uint64

	// rendering

	// input
	mousePosInPixels geometry.Point
	mousePosOnMap    geometry.Point

	// textserver
	serverStats     *TextServer
	serverInventory *TextServer
	serverEquipment *TextServer
	serverLog       *TextServer

	// player state
	isResting   bool

	// game state
	shouldQuitGame   bool
	quitToAscii      bool
	debugModeEnabled bool

	// event handler
	executeAfterTurn []func()

	// data definitions
	dataDefinitions defs.DataContainer


	actions *Actions
	factory *common.Factory

	logHook     func(string)
	logBuffer   []string
	onMapChange func()
}

func (g *GameInstance) SetConsoleMode(consoleMode bool) {
	// only disable audio for now
	g.audioPlayer.SetDisablePlayback(consoleMode)
}
func (g *GameInstance) GetTicks() uint64 {
	return g.worldTicks
}

func (g *GameInstance) GetActionBroker() *common.Broker {
	return g.actionBroker
}
func (g *GameInstance) GetMoveTowardsPlayer(from geometry.Point) geometry.Point {
	return g.getMoveOnPlayerDijkstraMap(from, true)
}

func (g *GameInstance) GetMoveAwayFromPlayer(from geometry.Point) geometry.Point {
	return g.getMoveOnPlayerDijkstraMap(from, false)
}
func (g *GameInstance) getMoveOnPlayerDijkstraMap(from geometry.Point, towardsPlayer bool) geometry.Point {
	dijkstraMap := g.state.GetPlayerDijkstraMap()
	if dijkstraMap == nil {
		return from
	}
	currentDistanceToPlayer, exist := dijkstraMap[from]
	if !exist {
		return from
	}
	var compareFunc func(a, b int) bool
	if towardsPlayer {
		compareFunc = func(a, b int) bool { return a < b }
	} else {
		compareFunc = func(a, b int) bool { return a > b }
	}
	currentMap := g.state.GetMap()
	neighbors := currentMap.GetFilteredNeighborsForMovement(from, func(pos geometry.Point) bool { // choose a possible next step
		if !currentMap.IsCurrentlyPassable(pos) { // only walk on passable tiles
			return false
		}
		if _, existsTransition := currentMap.GetTransitionAt(pos); existsTransition { // don't walk on transitions
			return false
		}
		neighborDist, neighborExists := dijkstraMap[pos]
		if !neighborExists {
			return false
		}
		neighborsWithTransition := currentMap.GetFilteredNeighbors(pos, func(pos geometry.Point) bool { // in fact, don't walk on tiles next to transitions either
			if !currentMap.IsCurrentlyPassable(pos) {
				return false
			}
			if _, existsTransition := currentMap.GetTransitionAt(pos); existsTransition {
				return true
			}
			return false
		})
		if len(neighborsWithTransition) > 0 {
			return false
		}
		return compareFunc(neighborDist, currentDistanceToPlayer) // depends on rolling direction on our dijkstra map
		//return neighborDist < currentDistanceToPlayer // depends on rolling direction on our dijkstra map
	})

	if len(neighbors) == 0 {
		return from
	}
	// we found some locations we can move to, that also bring us closer to the player
	// which of these is the closest to the player?

	nearestDist := math.MaxInt
	nearestPos := from

	for _, neighbor := range neighbors {
		neighborDist, _ := dijkstraMap[neighbor]
		//if neighborDist < nearestDist {  // depends on rolling direction on our dijkstra map
		if compareFunc(neighborDist, nearestDist) {
			nearestDist = neighborDist
			nearestPos = neighbor
		}
	}

	return nearestPos
}

func (g *GameInstance) GetAudio() *audio.Player {
	return g.audioPlayer
}

func NewGameInstance(startInfo common.StartInfo, uiController common.UserInterface) *GameInstance {
	g := &GameInstance{
		startInfo:    startInfo,
		uiController: uiController,
		randomSource: rand.New(rand.NewSource(time.Now().UnixNano())),
		audioPlayer:  audio.NewAudioPlayer(startInfo.FileSystem, audio.SampleRate44100),
	}
	println(fmt.Sprintf("GameInstance created.."))
	println(fmt.Sprintf("Random source loaded.."))
	println(fmt.Sprintf("Audio player loaded.."))
	g.dataDefinitions = defs.LoadGameData(func(err string) {
		println(err)
	})
	println(fmt.Sprintf("Data definitions loaded.."))
	g.ResetGame(g.dataDefinitions.ClassDefinitions[0])
	return g
}

func (g *GameInstance) hasEnergyForAction(being common.TimedBeing) (int, func()) {
	switch typedBeing := being.(type) {
	case *common.Actor:
		return g.getAIAction(typedBeing)
	case Turner:
		return g.getTurnerAction(being.GetBaseSpeed(), being.GetEnergy(), typedBeing)
	}
	return 0, nil
}
func (g *GameInstance) giveAndEquipItemToActor(actor *common.Actor, item *common.Item) {
	actor.GetInventory().Add(item)
	g.actionBroker.Execute(common.NewEquipItemAction(common.AbstractCause("Game Start"), actor, item))
}

// getActionContext requires the following fields to be initialized:
// g.audioPlayer, g.dataDefinitions, g.state
func (g *GameInstance) getActionContext() common.ActionContext {
	return common.ActionContext{
		GameState:     g.GetWorldState(),
		UserInterface: g.GetUserInterface,
		AudioPlayer:   g.audioPlayer,
		Data:          g.dataDefinitions,
		Msg:           g.PlayerMessage,
		Speak:         g.ActorSpeak,
	}
}

func (g *GameInstance) ResetGame(playerStartDefinition defs.LoadoutDefinition) {
	// we have loaded at this point:
	// - random source
	// - audio player
	// - data definitions
	g.actions = nil
	g.actionBroker = nil
	g.factory = nil
	//g.uiController =

	g.configuration = g.startInfo.UserConfiguration
	g.deviceIndependentScreenSize = g.startInfo.ScreenSize

	if g.configuration.ServeLog {
		g.serverLog = NewTextServer("rogue_log", 1)
	}
	if g.configuration.ServeStats {
		g.serverStats = NewTextServer("rogue_stats", 2)
	}
	if g.configuration.ServeInventory {
		g.serverInventory = NewTextServer("rogue_inventory", 3)
	}
	if g.configuration.ServeEquipment {
		g.serverEquipment = NewTextServer("rogue_equipment", 4)
	}

	// WE DO NOT HAVE A UI YET
	if g.configuration.ServeLog {
		g.SetLogHook(g.serverLog.Write)
	}

	ldtkFileName := g.startInfo.LDTKMapFile
	ldtkFile, _ := files.ReadFile(ldtkFileName)
	ldtkProject, _ := ldtk.Open(ldtkFile)

	timeKeeper := common.NewTimeKeeper(g.hasEnergyForAction)
	println(fmt.Sprintf("Timekeeper created.."))

	g.state = common.NewState(ldtkProject)
	g.state.SetPlayerMessage(g.PlayerMessage)
	g.state.SetTimeKeeper(timeKeeper)
	g.state.SelectedClass = playerStartDefinition

	g.actionBroker = common.NewBroker(g.getActionContext())
	g.actionBroker.RegisterHook(common.NewOnHealthChangedHook())

	g.actionBroker.RegisterHook(common.NewMainOnMoveHook(g.configuration.AutoPickup)) // update fov, etc.
	g.actionBroker.RegisterHook(common.NewMainOnMeleeHook())
	g.actionBroker.RegisterHook(common.NewMainOnRangedHook())
	g.actionBroker.RegisterHook(common.NewMainOnKillHook())
	g.actionBroker.RegisterHook(common.NewOnDieHook("lucifer talking", func(itemContext *common.ItemContext, action common.Action, context common.ActionContext) []common.Action {
		dieAction := action.(*common.DieAction)
		dyingActor := dieAction.GetActor()
		if dyingActor.GetInternalName() == "lucifer" {
			g.onLuciferDeath()
		}
		return nil
	}))

	g.actionBroker.RegisterHook(common.NewMainOnDropHook())
	g.actionBroker.RegisterHook(common.NewMainOnEquipHook())

	g.factory = common.NewFactory(common.FactoryContext{
		Definitions:             g.dataDefinitions,
		RandomSource:            g.randomSource,
		State:                   g.GetWorldState(),
		ActionBroker:            g.actionBroker,
		GetTimeKeeper:           g.state.GetTimeKeeper,
		GiveAndEquipItemToActor: g.giveAndEquipItemToActor,
		Msg:                     g.PlayerMessage,
	})
	g.factory.SetUserInterface(g.GetUserInterface)
	g.factory.SetQueueAfterTurn(g.queueAfterTurn)


	g.state.SetPlayerMessage(g.PlayerMessage)
	g.state.SetPlayerTurnHandler(g.onPlayerTurnTaken)
	g.state.SetAnimator(g.actionBroker)
	g.state.SetActionBroker(g.actionBroker)
	g.state.SetFactory(g.factory)

	g.createPlayer(playerStartDefinition)
	g.actions = NewActions(PlayerActionsContext{
		Player:        g.state.GetPlayer(),
		ActionBroker:  g.actionBroker,
		State:         g.GetWorldState(),
		UI:            g.GetUserInterface,
		GetTimeKeeper: g.state.GetTimeKeeper,
		Msg:           g.PlayerMessage,
		Animator:      g.actionBroker,
		Speak:         g.ActorSpeak,
		Configuration: g.configuration,
		QueueAfterTurn: g.queueAfterTurn,
	})
	g.state.SetOnMapChanged(g.onMapChanged)
	g.state.MapTransition(gridmap.Transition{
		TargetMap:      string(defs.LevelNameTownHell), // 7DRL: the ride starts in hell
		TargetLocation: "spawn",
	})

	mapGen := common.NewMapGenerator(g.randomSource)
	mapGen.SetFactory(g.factory)
	mapGen.SetLdtkProject(ldtkProject)

	// set references
	g.state.SetMapGenerator(mapGen)


	townSpawn := g.state.GetMap().GetNamedLocation("spawn")
	g.state.SetTownSpawn(townSpawn)

	//g.spawnActorOnMap(g.state.GetPlayer(), townSpawn, playerStartDefinition)

	// load audio
	if g.configuration.AudioEnabled {
		sfxPath := path.Join("assets", "sfx")
		cues := []string{
			"heavy_door",
			"missile_hit_human",
			"heavy_hit_human",
			"levelup",
			"bat_attack",
			"skeleton_hit",
			"skeleton_death",
		}
		sfxFiles := make([]string, len(cues))
		for i, sfxFile := range cues {
			sfxFiles[i] = path.Join(sfxPath, sfxFile+".ogg")
		}
		g.audioPlayer.RegisterSoundCues(sfxFiles)
		g.audioPlayer.RegisterRandomizedSoundCues([]string{
			path.Join(sfxPath, "bow_fire"),
			//path.Join(sfxPath, "close_door"),
			path.Join(sfxPath, "open_door"),
			path.Join(sfxPath, "melee_attack"),
			path.Join(sfxPath, "human_hit"),
		})

		g.audioPlayer.PreLoadCuesIntoMemory(cues)
	}

	g.state.GetFlags().SetTrue("game_started")

	g.actionBroker.Execute(common.NewUpdateFoVAction(common.AbstractCause("Game Start"), g.state.GetPlayer()))

	g.uiController.UpdateStatsLabel()

	g.state.GetFlags().SetTrue(string(common.FlagStoryModeEnabled))
}

// SetAndInitUserInterface will be called after the game has finished initialization
// So we have to inform every component about the ui
func (g *GameInstance) SetAndInitUserInterface(ui common.UserInterface) {
	g.uiController = ui
	ui.SetPlayerMessage(g.PlayerMessage)
	ui.OnMapChanged()
	ui.SetDebugActions(g.getDebugActions)
	g.showQuest()
}

// showQuest opens the 7DRL story text (shipped build: on start and on F3).
func (g *GameInstance) showQuest() {
	lines := []util.LabelText{
		util.LabelTextFromStringWithColor("Abaddon - a 7 day roguelike by Felix Ruzzoli", color.White),
		util.LabelTextFromString(""),
	}
	lines = append(lines, common.NewDialogueTextFromFile("abaddon_quest.txt")...)
	lines = append(lines, util.LabelTextFromString(""), util.LabelTextFromString("Press F1 during play for help"))
	g.uiController.OpenTextWindow(lines)
}
func (g *GameInstance) queueAfterTurn(call func()) {
	g.executeAfterTurn = append(g.executeAfterTurn, call)
}
func (g *GameInstance) ActorSpeak(actor *common.Actor, trigger defs.ActorSpeechTrigger) {
	textLines := actor.GetSpeech(g.randomSource, trigger)
	if len(textLines) == 0 {
		return
	}

	if oldAnim, exists := g.actionBroker.GetAnimation(func(anim common.Animation) bool {
		if speechAnim, isSpeechAnim := anim.(*common.SpeechAnimation); isSpeechAnim {
			return speechAnim.GetActor() == actor
		}
		return false
	}); exists {
		oldAnim.Cancel()
	}

	speechAnim := common.NewSpeechAnimation(actor, textLines)
	g.actionBroker.AddAnimation(speechAnim)

	g.uiController.Say(actor, textLines)
	g.PlayerMessage(fmt.Sprintf("%s says: %s", actor.Name(), strings.Join(textLines, "\n")))
}

func (g *GameInstance) GetUserInterface() common.UserInterface {
	return g.uiController
}

func (g *GameInstance) spawnDebugItems() {
	masterSword := NewMasterSword(g.GetWorldState())
	g.spawnItem(geometry.Point{X: 15, Y: 10}, masterSword)
}

func (g *GameInstance) handleLightSource() {
	player := g.state.GetPlayer()
	if player.GetEquipment().HasLightSourceEquipped() && !g.state.PlayerStandsInLitArea() {
		lightSource := player.GetEquipment().GetLightSource()
		asGameItem := lightSource
		asGameItem.UseCharge()
	}
}
func (g *GameInstance) onPlayerTurnTaken(autUsed int) {
	player := g.state.GetPlayer()

	g.state.GetTimeKeeper().PlayerTakesTurn(autUsed)

	// execute after turn events, =?
	for _, event := range g.executeAfterTurn {
		event()
	}
	g.executeAfterTurn = make([]func(), 0)

	g.checkForLevelUp()

	player.BeforeTurn()
	g.handleLightSource()


	// automatic pickup & player movement..
	if g.configuration.AutoPickup && g.state.GetMap().IsItemAt(player.MapPosition()) {
		itemAtLocation := g.state.GetMap().ItemAt(player.MapPosition())
		if itemAtLocation != nil && player.CanPickup() && !itemAtLocation.IsDroppedByPlayer() {
			pickUpAction := common.NewPickUpAction(common.AbstractCause("AutoPickup"), player)
			endTurnAction := common.NewEndPlayerTurnAction(pickUpAction.Cause(), pickUpAction.GetDelayInAut())
			g.actionBroker.Execute(pickUpAction)
			g.actionBroker.Execute(endTurnAction)
		}
	} else {
		g.actions.nextStepOnPlayerPath()
	}

	g.uiController.OnPlayerTurnTaken()
}

func (g *GameInstance) procFireAtLocation(location geometry.Point, startAnimating func() bool) {
	if g.state.GetMap().IsActorAt(location) {
		actor := g.state.GetMap().ActorAt(location)
		action := common.NewChangeResourceAction(common.AbstractCause("Fire"), actor, stats.Health, -3)
		action.SetAnimationStartPredicate(startAnimating)
		g.actionBroker.Execute(action)
	}
}
func (g *GameInstance) updateStatsDisplay() {
	g.uiController.UpdateStatsLabel()
	g.sendStatsToWindowClient()
}

func (g *GameInstance) createPlayer(playerStartDefinition defs.LoadoutDefinition) *common.Actor {
	player := common.NewActor(playerStartDefinition.Icon)
	player.SetAsciiIcon('@')
	player.SetCP437Icon(int32(util.UnicodeToCP437Byte('☺')))
	player.SetColor(defs.White)
	player.SetName("Player")
	g.state.SetPlayer(player)

	// Innate Skills
	g.initPlayerSkills(player, playerStartDefinition.TargetedSkills, playerStartDefinition.UntargetedSkills)

	// Starting gear
	// 7DRL: the Avenger always rides out with the sword "Vengeance" (reconstructed, the jam code is lost)
	if vengeance := g.factory.NewUniqueItemFromName("sword_of_vengeance"); vengeance != nil {
		g.giveAndEquipItemToActor(player, vengeance)
	}
	for _, itemName := range playerStartDefinition.Items {
		item := g.factory.NewItemFromDefinitionName(itemName, 1)
		if item != nil {
			g.giveAndEquipItemToActor(player, item)
		}
	}

	playerStats := player.GetStats()
	for attr, value := range playerStartDefinition.Attributes {
		playerStats.SetAttributeAsFloat(stats.StatFromString(attr), value)
	}
	for resourceName, value := range playerStartDefinition.Resources {
		maxAttributeName := "max_" + resourceName
		playerStats.SetAttributeAsFloat(stats.StatFromString(maxAttributeName), value)
		playerStats.GetResource(stats.StatFromString(resourceName)).SetToMax()
	}
	// Gold
	playerStats.SetAttribute(stats.Gold, playerStartDefinition.Gold)
	player.GetStats().SetOnChangeHandler(g.updateStatsDisplay)
	player.GetEquipment().SetOnChangeHandler(g.sendEquipmentToWindowClient)
	player.GetInventory().SetOnChangeHandler(g.sendInventoryToWindowClient)
	return player
}

func (g *GameInstance) spawnItem(mapPos geometry.Point, item *common.Item) {
	g.state.GetMap().AddItem(item, mapPos)
}

func (g *GameInstance) spawnItemInTown(namedLocation string, item *common.Item) {
	townMap := g.state.GetTownMap(string(defs.LevelNameTown))
	location := townMap.GetNamedLocation(namedLocation)
	townMap.AddItem(item, location)
}

func (g *GameInstance) GetWorldState() *common.State {
	return &g.state
}

func (g *GameInstance) showDebugMenu() {
	g.uiController.OpenMenu([]util.MenuItem{
		{
			MainText: "Toggle Debug Mode",
			Action: func() {
				g.debugModeEnabled = !g.debugModeEnabled
			},
		},
		{
			MainText: "Open class selection",
			Action: func() {
				g.uiController.OpenClassSelectionMenu()
			},
		},
		{
			MainText: "Drop to ASCII Mode",
			Action: func() {
				g.quitToAscii = true
			},
		},
		{
			MainText: "Map Info",
			Action: func() {
				g.uiController.OpenTextWindow(util.ToLabelText(g.state.GetMap().GetMetaString()))
			},
		},
		{
			MainText: "Put End-Game Gear into Stash",
			Action: func() {
				stash := g.state.GetPlayerStashInTown()
				stash.Add(g.factory.NewUniqueItemFromName("veil_of_darkness"))
				stash.Add(g.factory.NewUniqueItemFromName("plate_mail_of_the_dark_lord"))
				stash.Add(g.factory.NewUniqueItemFromName("eternal_darkness"))
				stash.Add(g.factory.NewUniqueItemFromName("never_ending_night"))
				stash.Add(g.factory.NewUniqueItemFromName("ancient_nightmares_hood"))
				stash.Add(g.factory.NewUniqueItemFromName("boots_of_the_shadow_dancer"))
				stash.Add(g.factory.NewUniqueItemFromName("black_hands"))
				stash.Add(g.factory.NewUniqueItemFromName("amulet_of_shadows"))
				stash.Add(g.factory.NewUniqueItemFromName("shadows_embrace"))
				stash.Add(g.factory.NewUniqueItemFromName("wall_of_thorns"))
				stash.Add(g.factory.NewUniqueItemFromName("soul_reaver"))
				stash.Add(g.factory.NewUniqueItemFromName("malleus_maleficarum"))
				stash.Add(g.factory.NewUniqueItemFromName("light_morning_star"))
			},
		},
		{
			MainText: "Put Level 2 Gear into Stash",
			Action: func() {
				stash := g.state.GetPlayerStashInTown()
				stash.Add(g.factory.NewItemFromDefinitionName("hard_leather_cap", 2))
				stash.Add(g.factory.NewItemFromDefinitionName("partial_leather_armour", 2))
				stash.Add(g.factory.NewItemFromDefinitionName("leather_gloves", 2))
				stash.Add(g.factory.NewItemFromDefinitionName("leather_boots", 2))
				stash.Add(g.factory.NewItemFromDefinitionName("cloak", 2))
				stash.Add(g.factory.NewItemFromDefinitionName("small_leather_shield", 2))
				stash.Add(g.factory.NewItemFromDefinitionName("long_bow", 2))
			},
		},
		{
			MainText: "Put Random Wands into Stash",
			Action: func() {
				stash := g.state.GetPlayerStashInTown()
				for i := 0; i < 10; i++ {
					stash.Add(g.factory.NewRandomWand(1))
				}
			},
		},
		{
			MainText: "Enchant Stash to +10",
			Action: func() {
				stash := g.state.GetPlayerStashInTown()
				for _, item := range stash.Items() {
					gameItem := item
					if item.IsWeapon() {
						gameItem.GetWeaponStats().Plus = 10
					}
					if item.IsArmor() {
						gameItem.GetArmorStats().Plus = 10
					}
				}
			},
		},
		{
			MainText: "Toggle UI Overlays", // should be an ui specific option
			Action: func() {
				if g.uiController.HasOverlays() {
					//g.uiController.ClearAllOverlays()
					//g.mapRenderer.OnScreenSizeChanged(g.GetDeviceIndependentScreenSize())
				} else {
					//g.setupUI()
				}
			},
		},
		{
			MainText: "Debug Map Info",
			Action: func() {
				visibleMap := g.uiController.GetVisibleMap()
				g.PlayerMessage(fmt.Sprintf("Visible map: (%d, %d) - (%d, %d)", visibleMap.Min.X, visibleMap.Min.Y, visibleMap.Max.X, visibleMap.Max.Y))
			},
		},
		{
			MainText: "Spawn Debug Items",
			Action: func() {
				g.spawnDebugItems()
			},
		},
		{
			MainText: "Trigger nightfall in town",
			Action: func() {
				g.state.GetFlags().SetTrue("town_changed_to_night")
			},
		},
		{
			MainText: "Trigger death of Lucifer",
			Action: func() {
				g.onLuciferDeath()
			},
		},
		{
			MainText: "Trigger completion of poisoned water supply",
			Action: func() {
				g.state.GetFlags().SetTrue(string("all_clear:" + defs.SpecialLevelPoisonedWaterSupply))
			},
		},
		{
			MainText: "Toggle Story Mode",
			Action: func() {
				newState := !g.state.IsStoryModeEnabled()
				g.GetWorldState().GetFlags().Set(string(common.FlagStoryModeEnabled), newState)
				if newState {
					g.PlayerMessage("Story mode enabled")
				} else {
					g.PlayerMessage("Story mode disabled")
				}
			},
		},
		{
			MainText: "Open Stash",
			Action:   g.uiController.OpenPlayerStashInTown,
		},
		{
			MainText: "give stuff",
			Action: func() {
				playerStats := g.state.GetPlayer().GetStats()
				playerStats.SetAttribute(stats.MaxHealth, 10000)
				playerStats.GetResource(stats.Health).Increase(10000)
				playerStats.SetAttribute(stats.MaxMana, 10000)
				playerStats.GetResource(stats.Mana).Increase(10000)
				playerStats.SetAttribute(stats.Gold, 200000)

				inv := g.state.GetPlayer().GetInventory()
				for i := 0; i < 10; i++ {
					inv.Add(g.factory.NewItemFromDefinitionName("healing_potion", 1))
					inv.Add(g.factory.NewItemFromDefinitionName("mana_potion", 1))
					inv.Add(g.factory.NewItemFromDefinitionName("speed_potion", 1))
				}
				g.uiController.UpdateStatsLabel()
			},
		},
		{
			MainText: "give 10k XP",
			Action: func() {
				playerStats := g.state.GetPlayer().GetStats()
				playerStats.SetAttribute(stats.Experience, 10000)
				g.uiController.UpdateStatsLabel()
			},
		},
	})
}

func (g *GameInstance) killLevel() {
	actorsOnMap := g.state.GetMap().Actors()
	for i := len(actorsOnMap) - 1; i >= 0; i-- {
		actor := actorsOnMap[i]
		if actor == g.state.GetPlayer() || !actor.IsAlive() || actor.IsUnique() {
			continue
		}
		currentHP := actor.GetStats().GetResource(stats.Health).Value() * 2
		g.actionBroker.Execute(common.NewChangeResourceAction(common.NewActorCause("Kill All", g.state.GetPlayer()), actor, stats.Health, float64(-currentHP)))
	}
}

// getAIAction returns the action to execute and the cost in AUT associated with it
func (g *GameInstance) getAIAction(actor *common.Actor) (int, func()) {
	autCostMove := 100 / actor.GetMoveSpeed()
	autCostMeleeAttack := 100 / actor.GetAttackSpeed(common.MeleeAttack)
	autCostRangedAttack := 100 / actor.GetAttackSpeed(common.RangedAttack)
	minCostAttack := min(autCostMeleeAttack, autCostRangedAttack)

	minCost := min(autCostMove, minCostAttack)
	if actor.GetEnergy() < minCost { // don't act at all, conserve energy
		return 0, nil
	}
	if actor.IsVendor() { // vendors just wait by default
		return 10, func() {}
	}

	canAct := !actor.HasStatusEffect(common.StatusStunned)
	if !canAct {
		//println(fmt.Sprintf("%s is still unable to act. (Stunned)", actor.Name()))
		return 10, func() {}
	}

	canMove := actor.GetEnergy() >= autCostMove && !actor.HasStatusEffect(common.StatusImmobile)
	canAttack := actor.GetEnergy() >= minCostAttack && !actor.HasStatusEffect(common.StatusPacified)

	if !canMove && !canAttack {
		//println(fmt.Sprintf("%s can neither move nor attack. (Immobile & Pacified)", actor.Name()))
		return 10, func() {}
	}

	isScared := actor.HasStatusEffect(common.StatusScared)
	if isScared && canMove {
		return autCostMove, func() {
			targetPos := g.GetMoveAwayFromPlayer(actor.MapPosition())
			if g.state.IsClosedDoorAt(targetPos) {
				g.actionBroker.Execute(common.NewOpenDoorAction(common.AbstractCause("OpenDoor"), actor, targetPos))
			}
			g.actionBroker.Execute(common.NewMoveAction(common.AbstractCause("Movement"), actor, targetPos, false))
		}
	}

	isConfused := actor.HasStatusEffect(common.StatusConfused)

	if isConfused && canMove {
		return autCostMove, func() {
			possibleNeighbors := g.state.GetMap().GetFilteredNeighbors(actor.MapPosition(), func(point geometry.Point) bool {
				return g.state.GetMap().IsWalkable(point)
			})
			if len(possibleNeighbors) == 0 {
				return
			}
			randomNeighbor := possibleNeighbors[g.randomSource.Intn(len(possibleNeighbors))]
			if g.state.IsClosedDoorAt(randomNeighbor) {
				g.actionBroker.Execute(common.NewOpenDoorAction(common.AbstractCause("OpenDoor"), actor, randomNeighbor))
			}
			if g.state.GetMap().IsActorAt(randomNeighbor) {
				actorAt := g.state.GetMap().ActorAt(randomNeighbor)
				// attack
				g.actionBroker.Execute(common.NewMeleeAttackAction(common.AbstractCause("MeleeAttack"), actor, actorAt))
			} else {
				g.actionBroker.Execute(common.NewMoveAction(common.AbstractCause("Movement"), actor, randomNeighbor, false))
			}
		}
	}

	// random movement for idle actors
	if actor.GetState() == common.ActorStateIdle && canMove {
		return autCostMove, func() {
			freeNeighbor := g.state.GetMap().GetRandomFreeAndSafeNeighbor(g.randomSource, actor.MapPosition())
			if freeNeighbor != actor.MapPosition() {
				if g.state.IsClosedDoorAt(freeNeighbor) {
					g.actionBroker.Execute(common.NewOpenDoorAction(common.AbstractCause("OpenDoor"), actor, freeNeighbor))
				}
				moveAction := common.NewMoveAction(common.AbstractCause("Movement"), actor, freeNeighbor, false)
				if g.actions.IsPlayerAutoMoving() {
					moveAction.DisableAnimation()
				}
				g.actionBroker.Execute(moveAction)
			}
		}
	}

	player := g.state.GetPlayer()
	if actor.GetState() == common.ActorStateAttacking && player.IsAlive() {
		// can use a skill?
		playerPosition := player.MapPosition()

		untargetedSkills := actor.GetUntargetedSkills()
		if len(untargetedSkills) > 0 {
			for _, skill := range untargetedSkills {
				if !skill.CanActivate(actor) {
					continue
				}
				skillUseChance := g.randomSource.Intn(10) == 0
				if !skillUseChance {
					continue
				}
				return skill.GetTimeNeededInAut(), func() {
					skill.Invoke(actor, "ai_skill_usage")
				}
			}
		}

		targetedSkills := actor.GetTargetedSkills()
		if len(targetedSkills) > 0 {
			for _, skill := range targetedSkills {
				skillUseChance := g.randomSource.Intn(10) == 0
				if !skillUseChance {
					continue
				}
				targetingRules := skill.GetTargetingRules(actor, g.GetWorldState().GetMap())
				targetInfo := targetingRules.OnTargetChanged(playerPosition)
				if !targetInfo.IsValid {
					continue
				}
				return skill.GetTimeNeededInAut(), func() {
					skill.Execute(actor, actor.MapPosition(), playerPosition)
				}
			}
		}

		canRangedAttack := actor.GetEnergy() >= autCostRangedAttack && actor.GetEquipment().HasRangedWeaponEquipped() && !actor.HasStatusEffect(common.StatusPacified)
		canSeePlayer := actor.CanSee(playerPosition)
		if canRangedAttack && canSeePlayer {
			return autCostRangedAttack, func() {
				g.actionBroker.Execute(common.NewRangedAttackAction(common.AbstractCause("RangedAttack"), actor, playerPosition))
			}
		}
		var dist int
		dist = g.state.GetMap().MoveDistance(actor.MapPosition(), playerPosition)
		if dist <= 1 && canAttack {
			return autCostMeleeAttack, func() {
				g.actionBroker.Execute(common.NewMeleeAttackAction(common.AbstractCause("MeleeAttack"), actor, player))
			}
		} else if canMove {
			return autCostMove, func() {
				targetPos := g.GetMoveTowardsPlayer(actor.MapPosition())
				if g.state.IsClosedDoorAt(targetPos) {
					g.actionBroker.Execute(common.NewOpenDoorAction(common.AbstractCause("OpenDoor"), actor, targetPos))
				}
				g.actionBroker.Execute(common.NewMoveAction(common.AbstractCause("Movement"), actor, targetPos, false))
			}
		}
	}

	return 10, func() {}
}

func (g *GameInstance) TransferRandomEnchantment(source *common.Item, destination *common.Item) {
	enchantments := source.GetAllEnchantments()

	if len(enchantments) == 0 {
		return
	}

	index := rand.Intn(len(enchantments))
	randomEnchantment := enchantments[index]

	source.RemoveEnchantment(randomEnchantment)
	destination.AddEnchantment(randomEnchantment)
}

func (g *GameInstance) onMapChanged(oldMapName string, newMapName string) {
	if defs.IsTownLevel(newMapName) {
		currentMap := g.state.GetMap()
		g.state.SetTownMap(currentMap)
		for _, actor := range currentMap.Actors() {
			if !actor.IsVendor() {
				continue
			}
			g.ensureVendorInventory(actor)
		}
		if newMapName == string(defs.LevelNameTown) {
			g.setupTownTransitions(currentMap)
		} else if newMapName == string(defs.LevelNameTownHell) {
			g.setupHellTownTransitions(currentMap)
		}
	} else { // dungeon map
		g.state.SetDungeonMap(newMapName, g.state.GetMap())
	}

	if defs.IsTownLevel(oldMapName) {
		g.GetAudio().PlayCue("heavy_door")
	}

	g.actions.updateFoV()
	g.uiController.OnMapChanged()

	if g.onMapChange != nil {
		g.onMapChange()
	}
}

func (g *GameInstance) setupTownTransitions(currentMap *gridmap.GridMap[*common.Actor, *common.Item, common.Object]) {
	entrancePos := currentMap.GetNamedLocation("to_cathedral")

	entranceTile := currentMap.GetCell(entrancePos).TileType
	currentMap.SetTile(entrancePos, entranceTile.WithSpecial(gridmap.SpecialTileStairsDown))

	currentMap.AddTransitionAt(entrancePos, gridmap.Transition{
		TargetMap:      defs.LevelNameFromDepth(1),
		TargetLocation: "ladder_up",
	})

	entrancePos = currentMap.GetNamedLocation("to_catacombs")

	entranceTile = currentMap.GetCell(entrancePos).TileType
	currentMap.SetTile(entrancePos, entranceTile.WithSpecial(gridmap.SpecialTileStairsDown))

	currentMap.AddTransitionAt(entrancePos, gridmap.Transition{
		TargetMap:      defs.LevelNameFromDepth(4),
		TargetLocation: "to_town",
	})

	entrancePos = currentMap.GetNamedLocation("to_caves")

	entranceTile = currentMap.GetCell(entrancePos).TileType
	currentMap.SetTile(entrancePos, entranceTile.WithSpecial(gridmap.SpecialTileStairsDown))

	currentMap.AddTransitionAt(entrancePos, gridmap.Transition{
		TargetMap:      defs.LevelNameFromDepth(6),
		TargetLocation: "to_town",
	})
}
func (g *GameInstance) setupHellTownTransitions(currentMap *gridmap.GridMap[*common.Actor, *common.Item, common.Object]) {
	caveEntrace := currentMap.GetNamedLocation("to_caves")

	entranceTile := currentMap.GetCell(caveEntrace).TileType
	currentMap.SetTile(caveEntrace, entranceTile.WithSpecial(gridmap.SpecialTileStairsDown))

	currentMap.AddTransitionAt(caveEntrace, gridmap.Transition{
		TargetMap:      defs.LevelNameFromDepth(8),
		TargetLocation: "to_hell_town",
	})

	dungeonEntrance := currentMap.GetNamedLocation("to_dungeon")
	entranceTile = currentMap.GetCell(dungeonEntrance).TileType
	currentMap.SetTile(dungeonEntrance, entranceTile.WithSpecial(gridmap.SpecialTileStairsDown))

	currentMap.AddTransitionAt(dungeonEntrance, gridmap.Transition{
		TargetMap:      defs.LevelNameFromDepth(9),
		TargetLocation: "to_hell_town",
	})

}

func (g *GameInstance) ensureVendorInventory(actor *common.Actor) {
	inventory := actor.GetInventory()
	if !inventory.IsEmpty() {
		return
	}
	var inventoryItems []*common.Item
	definition, exists := g.dataDefinitions.VendorDefinitions[string(actor.GetVendorType())]
	if !exists {
		g.ErrLog(fmt.Sprintf("No store definition for %s", actor.GetVendorType()))
		return
	}

	inventoryItems = g.getVendorInventoryFromDefinition(definition)

	for _, item := range inventoryItems {
		inventory.Add(item)
	}
}

func (g *GameInstance) initPlayerSkills(player *common.Actor, targetedSkills []string, untargetedSkills []string) {
	var targeted []*common.TargetedAction
	for _, skillName := range targetedSkills {
		targeted = append(targeted, g.factory.NewTargetedActionFromName(skillName))
	}
	player.SetTargetedSkills(targeted)
	var untargeted []*common.UntargetedAction
	for _, skillName := range untargetedSkills {
		untargeted = append(untargeted, g.factory.NewUntargetedActionFromName(skillName))
	}
	player.SetUntargetedSkills(untargeted)
}

type Turner interface {
	Turn()
}

func (g *GameInstance) getTurnerAction(baseSpeed int, energy int, being Turner) (int, func()) {
	cost := 100 / baseSpeed
	if energy < cost {
		return 0, nil
	}
	return baseSpeed, being.Turn
}

func (g *GameInstance) getDisplacePositionForActor(victimToDisplace *common.Actor, direction geometry.Point) geometry.Point {
	currentMap := g.state.GetMap()
	victimPos := victimToDisplace.MapPosition()
	var displacePos geometry.Point
	stepsInDegrees := 45
	for turnBy := stepsInDegrees; turnBy < 180; turnBy += stepsInDegrees {
		curDir := geometry.RotateVector(direction, float64(turnBy))
		displacePos = victimPos.Add(curDir)
		if currentMap.IsCurrentlyPassable(displacePos) {
			return displacePos
		}

		curDir = geometry.RotateVector(direction, float64(-turnBy))
		displacePos = victimPos.Add(curDir)
		if currentMap.IsCurrentlyPassable(displacePos) {
			return displacePos
		}
	}
	displacePos = victimPos.Add(direction)
	if currentMap.IsCurrentlyPassable(displacePos) {
		return displacePos
	}
	// return source
	behindUs := victimPos.Add(direction.Mul(-1))
	return behindUs
}

func (g *GameInstance) checkForLevelUp() {
	if !g.state.GetPlayer().CanLevelUp() {
		return
	}
	g.GetAudio().PlayCue("levelup")
}



func (g *GameInstance) gotoHellTown() {
	dest := gridmap.Transition{TargetMap: "TownHell", TargetLocation: "spawn"}
	g.state.TransitionTo(dest)
	g.actionBroker.Execute(common.NewEndPlayerTurnAction(common.AbstractCause("Travel"), 10))
}

func (g *GameInstance) onLuciferDeath() {
	lines := []util.LabelText{
		util.LabelTextFromStringWithColor("Lucifer", color.White),
		util.LabelTextFromString(""),
	}

	luciferLines := append(lines, common.NewDialogueTextFromFile("lucifer_death.txt")...)
	g.uiController.OpenTextWindow(luciferLines)

	g.spawnItemInTown("true_sight_potion", g.NewPotionOfTrueSight())
}

func (g *GameInstance) sendStatsToWindowClient() {
	if !g.configuration.ServeStats {
		return
	}
	lines := g.GetWorldState().GetPlayer().GetFullStats()
	var asStrings []string
	for _, line := range lines {
		asStrings = append(asStrings, line.Text)
	}
	g.serverStats.Write(strings.Join(asStrings, "\n"))
}

func (g *GameInstance) sendInventoryToWindowClient() {
	if !g.configuration.ServeInventory {
		return
	}
	asStrings := g.state.GetPlayer().GetInventory().AsColorCodedStrings()
	g.serverInventory.Write(strings.Join(asStrings, "\n"))
}

func (g *GameInstance) sendEquipmentToWindowClient() {
	if !g.configuration.ServeEquipment {
		return
	}
	equipment := g.state.GetPlayer().GetEquipment()
	inventory := g.state.GetPlayer().GetInventory()
	items := sortEquipForUI(equipment)
	var asStrings []string

	maxItemLength := 0
	maxSlotLength := 0
	for _, item := range items {
		// index 0 => a, 1 => b, 2 => c, ...
		//itemColor := common.ColorCodeFromColor(item.TintColor())
		slotName := equipment.GetSlot(item)
		slotDesc := fmt.Sprintf("<-- %s", slotName.ToString())

		//itemLine := fmt.Sprintf("[:white]%c) %s%s", 'a'+index, itemColor, item.Name())
		shortcut := inventory.GetLetterForItem(item)
		itemLineWithoutColor := fmt.Sprintf("%c) %s", shortcut, item.BandStyleName())
		//minLineLength := len(itemLineWithoutColor) + len(slotDesc) + 1

		//asStrings = append(asStrings, itemLine)
		if len(itemLineWithoutColor) > maxItemLength {
			maxItemLength = len(itemLineWithoutColor)
		}
		if len(slotDesc) > maxSlotLength {
			maxSlotLength = len(slotDesc)
		}
	}

	for _, item := range items {
		// index 0 => a, 1 => b, 2 => c, ...
		itemColor := defs.ColorCodeFromColor(item.TintColor())

		slotName := equipment.GetSlot(item)
		slotDesc := fmt.Sprintf("[:white]<-- %s", slotName.ToString())
		shortcut := inventory.GetLetterForItem(item)
		itemLine := fmt.Sprintf("[:white]%c) %s%s", shortcut, itemColor, item.BandStyleName())

		withoutColor := common.RemoveColorCodes(slotDesc)
		padForSlot := (maxSlotLength + 1) - len(withoutColor)
		padForItem := maxItemLength - len(common.RemoveColorCodes(itemLine))
		slotDesc = util.LeftPadCount(slotDesc, padForSlot)
		itemLine = util.RightPadCount(itemLine, padForItem)

		asStrings = append(asStrings, itemLine+slotDesc)
	}
	g.serverEquipment.Write(strings.Join(asStrings, "\n"))
}
func sortEquipForUI(equip *common.Equipment) []*common.Item {
	itemList := equip.AllItems()

	slices.SortStableFunc(itemList, func(i, j *common.Item) int {
		slotNameI := equip.GetSlot(i)
		slotNameJ := equip.GetSlot(j)
		return cmp.Compare(slotNameI, slotNameJ)
	})
	return itemList
}
func (g *GameInstance) SetOnMapChange(onMapChange func()) {
	g.onMapChange = onMapChange
}

func (g *GameInstance) ErrLog(errorMessage string) {
	// write to stderr
	fmt.Fprintf(os.Stderr, errorMessage+"\n")
}

func (g *GameInstance) GetGameData() defs.DataContainer {
	return g.dataDefinitions
}

func (g *GameInstance) GenerateLoot(dungeonLevel int, count int) []*common.Item {
	var result []*common.Item
	for i := 0; i < count; i++ {
		result = append(result, g.factory.GenerateLoot(dungeonLevel))
	}
	return result
}

func (g *GameInstance) getVendorInventoryFromDefinition(definition defs.VendorDefinition) []*common.Item {
	itemsNeeded := max(len(definition.Always), definition.ItemCount.LinearRandom())
	sometimesCount := 0
	alwaysCount := 0
	if len(definition.Always) > 0 {
		alwaysCount = min(itemsNeeded, len(definition.Always))
	}
	if len(definition.Sometimes) > 0 {
		sometimesCount = max(1, itemsNeeded - alwaysCount)
	}


	var result []*common.Item

	for i := 0; i < len(definition.Always); i++ {
		// add all the always items at least once
		dungeonLevel := max(1, g.randomSource.Intn(g.state.GetLastDungeonLevel()))
		itemName := definition.Always[i]
		item := g.factory.NewItemFromDefinitionName(itemName, dungeonLevel)
		result = append(result, item)
	}

	for i := 0; i < sometimesCount; i++ {
		// choose a random item
		dungeonLevel := max(1, g.randomSource.Intn(g.state.GetLastDungeonLevel()))
		randomIndex := g.randomSource.Intn(len(definition.Sometimes))
		itemName := definition.Sometimes[randomIndex]
		item := g.factory.NewItemFromDefinitionName(itemName, dungeonLevel)
		result = append(result, item)
	}

	return result
}

func (g *GameInstance) GetActions() common.PlayerActions {
	return g.actions
}

func (g *GameInstance) PlayerMessage(message string) {
	//println(text)
	rawText := common.RemoveColorCodes(message)
	if g.logHook != nil {
		g.logHook(rawText)
	}
	if len(g.logBuffer) > 200 {
		g.logBuffer = g.logBuffer[1:]
	}
	g.logBuffer = append(g.logBuffer, message)

	g.GetUserInterface().ShowLogMessage(message)
}

func (g *GameInstance) SetLogHook(hook func(string)) {
	g.logHook = hook
}

func (g *GameInstance) GetLog() []string {
	return g.logBuffer
}

func (g *GameInstance) getDebugActions() common.DebugActions {
	return common.DebugActions{
		KillLevel:      g.killLevel,
		GotoHellTown:   g.gotoHellTown,
		OpenDebugMenu:  g.showDebugMenu,
		GoOneLevelDown: g.actions.goOneLevelDown,
		GoOneLevelUp:   g.actions.goOneLevelUp,
	}
}
