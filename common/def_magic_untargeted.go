package common

import (
	"cmp"
	"fmt"
	"image/color"
	"math"
	"rpg/defs"
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/stats"
	"rpg/util"
	"slices"
)

const (
	ActionNameHealHealthAndMana UntargetedActionName = "heal_health_mana_full"
	ActionNameFirestarter       UntargetedActionName = "firestarter"
	ActionNameBerserk           UntargetedActionName = "apply_berserk_self"
	ActionNameApplyStatusSelf   UntargetedActionName = "apply_status_self"
	ActionNameHealMana          UntargetedActionName = "heal_mana"
	ActionNameTeleport          UntargetedActionName = "random_near_teleport"
	ActionNameBuffDamage        UntargetedActionName = "buff_damage"
	ActionNameBuffSpeed         UntargetedActionName = "buff_speed"
	ActionNameRecall            UntargetedActionName = "recall"
	ActionNameBless             UntargetedActionName = "buff_armor"
	ActionNameStoneSkin         UntargetedActionName = "stoneskin"
	ActionNameEnchantWeapon     UntargetedActionName = "enchant_weapon"
	ActionNameEnchantArmor      UntargetedActionName = "enchant_armor"
	ActionNameAcquirement       UntargetedActionName = "acquirement"
	ActionNameMultiShot         UntargetedActionName = "multi_shot"
	ActionNameTelefrag          UntargetedActionName = "telefrag"
	ActionNameReadMessage       UntargetedActionName = "read_message"
	ActionNameReplicate         UntargetedActionName = "replicate_with"
	ActionNameEarthquake        UntargetedActionName = "earthquake"
)

func (g *Factory) NewUntargetedActionFromName(name string) *UntargetedAction {
	var args util.Arguments

	if util.LooksLikeAFunction(name) {
		name, args = util.GetNameAndArgs(name)
	}

	if actionDef, exists := g.definitions.UntargetedDefinitions[name]; exists {
		if args.IsEmpty() && !actionDef.DefaultArgs.IsEmpty() {
			args = actionDef.DefaultArgs
		}
		return g.NewUntargetedActionFromDefinition(actionDef, args)
	}

	g.logError(fmt.Sprintf("no untargeted action found for name %s", name))
	return nil
}

type UntargetedEffect struct {
	Effect      func(user *Actor, source EffectSource)
	Description []util.LabelText
}

func (u UntargetedEffect) Invoke(user *Actor, source EffectSource) {
	u.Effect(user, source)
}

func (u UntargetedEffect) GetDescription() []util.LabelText {
	return u.Description
}

func (g *Factory) getReplicate(args util.Arguments) UntargetedEffect {
	effect := func(user *Actor, source EffectSource) {
		currentMap := g.state.GetMap()
		origin := user.MapPosition()
		// find position for spawning
		freePositions := currentMap.GetFreeCellsForDistribution(origin, 1, func(p geometry.Point) bool {
			return currentMap.IsCurrentlyPassable(p)
		})

		if len(freePositions) == 0 { // no space on map
			return
		}

		clonePosition := freePositions[0]

		// spawn clone
		clonedActor := user.Clone()
		clonedActor.SetDrawPosition(clonePosition.ToPointF())
		currentMap.AddActor(clonedActor, clonePosition)

		keeper := g.getTimeKeeper()

		keeper.AddTimed(clonedActor)
		if user == g.state.GetPlayer() {
			clonedActor.SetFollowLeader(user)
		}
		g.logError(fmt.Sprintf("A new %s is born!", clonedActor.Name()))
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString("Replicate"),
		},
	}
}

func (g *Factory) getReadMessage(textForMessage []util.LabelText, onRead func()) UntargetedEffect {
	effect := func(user *Actor, source EffectSource) {
		g.gui().OpenTextWindow(textForMessage)
		if onRead != nil {
			onRead()
		}
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString("Read message"),
		},
	}
}
func (g *Factory) getSpeedBuff(args util.Arguments) UntargetedEffect {
	amount := util.ParseDice(args.Get(0)).Roll()
	turnCount := util.ParseDice(args.Get(1)).Roll()
	effect := func(user *Actor, source EffectSource) {
		modifier := &stats.Modifier{
			Attribute:      stats.BaseSpeed,
			Description:    "of haste",
			Value:          float64(amount),
			Mode:           stats.ModifyAdditive,
			IsPercentBased: false,
			StackID:        string(source),
		}
		timekeeper := g.getTimeKeeper()

		RemoveTimedEffectsBySource(timekeeper, source)

		timedModifier := NewTimedModifier(string(ActionNameBuffSpeed), user, modifier, turnCount)
		timedModifier.Apply()
		timekeeper.AddTimed(timedModifier)

		g.logError("You feel the world around you slow down!")
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Haste for %d turns", turnCount)),
		},
	}
}

func RemoveTimedEffectsBySource(timekeeper *TimeKeeper, source EffectSource) {
	if source == "" {
		return
	}
	timeModsWithSameSource := timekeeper.ClearTrackedBeings(func(being TimedBeing) bool {
		if mod, ok := being.(*TimedModifier); ok {
			return mod.modifier.StackID == string(source)
		}
		if untargeted, ok := being.(*UntargetedAction); ok {
			return untargeted.effectSource == source
		}
		return false
	})
	for _, timedMod := range timeModsWithSameSource {
		if mod, ok := timedMod.(*TimedModifier); ok {
			mod.Remove()
		}
	}
}
func (g *Factory) getEnchantWeapon(args util.Arguments) UntargetedEffect {
	maxEnchantLevel := 3
	effect := func(user *Actor, source EffectSource) {
		isWeapon := func(element *Item) bool { return element.IsWeapon() }
		selected := func(itemKey rune) {
			g.gui().PopAll()
			inventory := user.GetInventory()
			selectedItem := inventory.GetItemByLetter(itemKey)
			weaponStats := selectedItem.GetWeaponStats()
			if weaponStats.Plus >= maxEnchantLevel {
				g.logError(fmt.Sprintf("Your [:darkText]%s[:white] cannot be enchanted any further.", selectedItem.Name()))
				return
			}
			weaponStats.Plus += 1
			g.logError(fmt.Sprintf("Your [:darkText]%s[:white] glows [:blue]blue[:white] for a moment.", selectedItem.Name()))
			g.audioPlayer.PlayCue("enchant")
		}

		if user == g.state.GetPlayer() {
			g.gui().OpenListInventoryForSelection(isWeapon, selected)
		} else if user.GetEquipment().HasWeaponEquipped() {
			// select wielded weapon
			wieldedWeapon := user.GetEquipment().GetBySlot(defs.SlotNameMainHand)
			weaponStats := wieldedWeapon.GetWeaponStats()
			if weaponStats.Plus >= maxEnchantLevel {
				return
			}
			weaponStats.Plus += 1
		}
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString("Enchant weapon damage"),
		},
	}
}
func (g *Factory) getEnchantArmor(args util.Arguments) UntargetedEffect {
	maxEnchantLevel := 3
	effect := func(user *Actor, source EffectSource) {
		isArmor := func(element *Item) bool { return element.IsArmor() }
		selected := func(itemKey rune) {
			g.gui().PopAll()
			inventory := user.GetInventory()
			selectedItem := inventory.GetItemByLetter(itemKey)
			weaponStats := selectedItem.GetArmorStats()
			if weaponStats.Plus >= maxEnchantLevel {
				g.logError(fmt.Sprintf("Your [:darkText]%s[:white] cannot be enchanted any further.", selectedItem.Name()))
				return
			}
			weaponStats.Plus += 1
			g.logError(fmt.Sprintf("Your [:darkText]%s[:white] glows [:blue]blue[:white] for a moment.", selectedItem.Name()))
			g.audioPlayer.PlayCue("enchant")
		}

		if user == g.state.GetPlayer() {
			g.gui().OpenListInventoryForSelection(isArmor, selected)
		} else if user.GetEquipment().HasArmorEquipped() {
			// select worn armor
			wornArmor := user.GetEquipment().GetBySlot(defs.SlotNameArmorTorso)
			armorStats := wornArmor.GetWeaponStats()
			if armorStats.Plus >= maxEnchantLevel {
				return
			}
			armorStats.Plus += 1
		}
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString("Enchant armor protection"),
		},
	}
}
func (g *Factory) getAcquirement(args util.Arguments) UntargetedEffect {
	level := 12
	if !args.IsEmpty() {
		level = util.ParseDice(args.Get(0)).Roll()
	}
	effect := func(user *Actor, source EffectSource) {
		state := g.state
		magicItem := g.GenerateMagicItem(level)
		state.AddItemToMapWithDistribution(magicItem, user.MapPosition())
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString("Acquire one item from another realm of existence"),
		},
	}
}
func (g *Factory) getHealHealthAndMana(args util.Arguments) UntargetedEffect {
	effect := func(user *Actor, source EffectSource) {
		healHealthAction := NewChangeResourceAction(NewActorCause("Full Heal", user), user, stats.Health, float64(user.GetStats().GetAttribute(stats.MaxHealth)))
		g.actionBroker.Execute(healHealthAction)
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString("Completely heal health and mana"),
		},
	}
}
func (g *Factory) getHealHealth(args util.Arguments) UntargetedEffect {
	effect := func(user *Actor, source EffectSource) {
		healHealthAction := NewChangeResourceAction(NewActorCause("Full Heal", user), user, stats.Health, float64(user.GetStats().GetAttribute(stats.MaxHealth)))
		g.actionBroker.Execute(healHealthAction)
		g.gui().UpdateStatsLabel()
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString("Completely heal health"),
		},
	}
}
func (g *Factory) logIfPlayer(user *Actor, message string) {
	if user == g.state.GetPlayer() {
		g.logError(message)
	}
}
func (g *Factory) getFirestarter(args util.Arguments) UntargetedEffect {
	effect := func(user *Actor, source EffectSource) {
		currentMap := g.state.GetMap()
		for _, pos := range user.GetFOV().Visibles {
			if !currentMap.IsObjectAt(pos) {
				continue
			}
			object := currentMap.ObjectAt(pos)
			if exploder, isExploder := object.(*ExplodingObject); isExploder {
				exploder.DamageReceived(5)
			}
		}
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString("Explode all visible barrels"),
		},
	}
}
func (g *Factory) getMultiShot(args util.Arguments) UntargetedEffect {
	shotCount := args.GetInt(0)
	effect := func(user *Actor, source EffectSource) {
		if !user.GetEquipment().HasRangedWeaponEquipped() {
			g.logError("No ranged weapon equipped")
			return
		}
		nearbyEnemies := g.getVisibleEnemiesByDistance(user)
		if len(nearbyEnemies) == 0 {
			g.logError("No enemies in sight")
			return
		}
		hitCount := min(shotCount, len(nearbyEnemies))
		for i := 0; i < hitCount; i++ {
			target := nearbyEnemies[i]
			rangedAttackAction := NewRangedAttackAction(NewActorCause("Multi Shot", user), user, target.MapPosition())
			g.actionBroker.Execute(rangedAttackAction)
		}
		g.tryEndTurn(user, 100/user.GetAttackSpeed(RangedAttack))
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Shoot at %d enemies at the same time", shotCount)),
		},
	}
}

func (g *Factory) getEarthquake(args util.Arguments) UntargetedEffect {
	earthQuakeRadius := args.GetIntOrDefault(0, 4)
	earthQuakeDamageRange := args.GetIntOrDefault(1, 10)
	effect := func(user *Actor, source EffectSource) {
		currentMap := g.state.GetMap()
		generator := g.randomSource
		allPOIs := g.state.GetAllPOIs()

		defWallIcon := defs.IconDungeonRockyWall
		defFloorIcon := defs.IconDungeonFloorAlternative1

		floorNumber := g.state.GetCurrentDungeonLevel()

		switch floorNumber {
		case 1, 2:
			defWallIcon = defs.IconCathedralWall
			defFloorIcon = defs.IconCathedralFloor
		case 3, 4:
			defWallIcon = defs.IconCatacombsWall
			defFloorIcon = defs.IconCatacombsFloor
		case 5, 6:
			defWallIcon = defs.IconDungeonRockyWall
			defFloorIcon = defs.IconDungeonRedEarthFloor
		case 7, 8:
			defWallIcon = defs.IconDungeonOfDoomWallModern
			defFloorIcon = defs.IconDungeonOfDoomFloorModern
		}
		wallTile := gridmap.Tile{
			DefinedIcon:        defWallIcon,
			DefinedDescription: "a wall",
			IsWalkable:         false,
			IsTransparent:      false,
		}

		floorTile := gridmap.Tile{
			DefinedIcon:        defFloorIcon,
			DefinedDescription: "floor tile",
			IsWalkable:         true,
			IsTransparent:      false,
		}

		areaOfEffect := currentMap.GetDijkstraMap(user.MapPosition(), earthQuakeRadius, func(point geometry.Point) bool {
			return !currentMap.IsStairsAt(point) && !currentMap.IsBorderWall(point)
		})


		delete(areaOfEffect, user.MapPosition())

		shuffleAreaOfEffect := func() {
			for pos, _ := range areaOfEffect {
				if generator.Intn(2) == 0 {
					continue
				}
				// flip terrain
				if currentMap.IsWalkable(pos) {
					currentMap.SetTile(pos, wallTile)
				} else {
					currentMap.SetTile(pos, floorTile)
				}
			}
		}

		// this is the real algorithm we are using
		// shuffle stuff around
		shuffleAreaOfEffect()
		// update the reachability map
		g.state.updatePlayerDijkstra()

		// do this until the level is connected
		for !g.state.IsReachableByPlayer(allPOIs) {
			shuffleAreaOfEffect()
			g.state.updatePlayerDijkstra()
		}

		getNearestFreePositionInAoE := func(origin geometry.Point) geometry.Point {
			minDistance := math.MaxInt64
			nearestFree := origin
			for pos, _ := range areaOfEffect {
				if !currentMap.CanPlaceActorHere(pos) {
					continue
				}
				distance := geometry.DistanceManhattan(origin, pos)
				if distance < minDistance {
					minDistance = distance
					nearestFree = pos
				}
			}
			return nearestFree
		}
		isSomethingStuckHere := func(pos geometry.Point) bool {
			if currentMap.IsTileWalkable(pos) {
				return false
			}
			if currentMap.IsObjectAt(pos) || currentMap.IsItemAt(pos) || currentMap.IsActorAt(pos) || currentMap.IsDownedActorAt(pos) {
				return true
			}
			return false
		}
		var affectedActorActions []Action
		for pos, _ := range areaOfEffect {
			if !isSomethingStuckHere(pos) {
				continue
			}
			if currentMap.IsItemAt(pos) { // destroy items on affected tiles
				currentMap.RemoveItemAt(pos)
			}

			if currentMap.IsObjectAt(pos) { // destroy objects on affected tiles
				currentMap.RemoveObjectAt(pos)
			}

			if currentMap.IsDownedActorAt(pos) { // destroy corpses on affected tiles
				downedActor := currentMap.DownedActorAt(pos)
				currentMap.RemoveDownedActor(downedActor)
			}

			if currentMap.IsActorAt(pos) {
				// move around the actors to the nearest free position and damage them in the process
				hitActor := currentMap.ActorAt(pos)
				newPos := getNearestFreePositionInAoE(hitActor.MapPosition())
				moveAction := NewMoveAction(NewActorCause("Earthquake", user), hitActor, newPos, false)
				affectedActorActions = append(affectedActorActions, moveAction)

				damage := generator.Intn(earthQuakeDamageRange) + 1
				damageAction := NewChangeResourceAction(NewActorCause("Earthquake", user), hitActor, stats.Health, float64(-damage))
				affectedActorActions = append(affectedActorActions, damageAction)

				updateFoV := NewUpdateFoVAction(NewActorCause("Earthquake", hitActor), hitActor)
				affectedActorActions = append(affectedActorActions, updateFoV)
			}
		}
		g.actionBroker.QueueForExecution(NewUpdateFoVAction(NewActorCause("Earthquake", user), user))
		g.actionBroker.QueueForExecutionAtTheSameTime(affectedActorActions)

		g.audioPlayer.PlayCue("earthquake")
		actionCost := 10 // default cost for any action is 10 aut
		g.tryEndTurn(user, actionCost)
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString("changes the terrain around the actor"),
		},
	}
}

func (g *Factory) tryEndTurn(user *Actor, autUsed int) {
	if user == g.state.GetPlayer() {
		g.actionBroker.QueueForExecution(NewEndPlayerTurnAction(NewActorCause("Action Used", user), autUsed))
	}
}

func (g *Factory) getBlessing(args util.Arguments) UntargetedEffect {
	//interval := stats.NewIntervalFromString(args.Get(0))
	turnCount := args.GetInt(0)
	effect := func(user *Actor, source EffectSource) {

		currentAC := user.GetStats().GetAttribute(stats.ArmorClass)
		acBonus := 0
		if currentAC < 5 {
			acBonus = 5
		} else {
			acBonus = currentAC
		}

		timeKeeper := g.getTimeKeeper()
		RemoveTimedEffectsBySource(timeKeeper, source)

		modifier := defs.AddToBaseAttribute(g.randomSource, stats.ArmorClass, float64(acBonus))
		modifier.StackID = string(source)

		timedModifier := NewTimedModifier(string(ActionNameBless), user, modifier, turnCount)
		timedModifier.Apply()
		timeKeeper.AddTimed(timedModifier)

		g.logIfPlayer(user, "You feel protected")
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Increase your Armor Class for %d turns", turnCount)),
		},
	}
}
func (g *Factory) getStoneSkinBuff(args util.Arguments) UntargetedEffect {
	//amount := args.GetInt(0)
	turnCount := util.ParseDice(args.Get(0)).Roll()
	effect := func(user *Actor, source EffectSource) {
		speedMod := &stats.Modifier{
			Attribute:      stats.BaseSpeed,
			Description:    "Slowed",
			Value:          float64(-5),
			Mode:           stats.ModifyAdditive,
			IsPercentBased: false,
			StackID:        string(source),
		}
		timeKeeper := g.getTimeKeeper()
		RemoveTimedEffectsBySource(timeKeeper, source)

		timedSpeedMod := NewTimedModifier(string(ActionNameStoneSkin), user, speedMod, turnCount)
		timedSpeedMod.Apply()
		timeKeeper.AddTimed(timedSpeedMod)

		currentAC := user.GetStats().GetAttribute(stats.ArmorClass)
		missingToThirty := max(0, 30-currentAC)
		acMod := defs.AddToBaseAttribute(g.randomSource, stats.ArmorClass, float64(missingToThirty))
		acMod.StackID = string(source)
		timedACMod := NewTimedModifier(string(ActionNameStoneSkin), user, acMod, turnCount)
		timedACMod.Apply()
		timeKeeper.AddTimed(timedACMod)

		g.logIfPlayer(user, "Your skin hardens but movement becomes difficult.")
		g.tryEndTurn(user, 100/user.GetBaseSpeed())
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Apply Stoneskin for %d turns", turnCount)),
		},
	}
}

func (g *Factory) getBerserkSelfAction(args util.Arguments) UntargetedEffect {
	if args.IsEmpty() {
		g.logError(fmt.Sprintf("Berserk effect requires arguments"))
	}
	intervalRange := util.ParseDice(args.Get(0))
	turnCount := args.GetInt(1)
	effect := func(user *Actor, source EffectSource) {
		applyStatusAction := NewApplyStatusAction(NewActorCause("ApplyStatusEffect", user), user, StatusBerserk, turnCount)
		g.actionBroker.Execute(applyStatusAction)

		timeKeeper := g.getTimeKeeper()

		RemoveTimedEffectsBySource(timeKeeper, source)
		modifier := defs.AddToBaseAttributeFromRoll(g.randomSource, stats.Blood, intervalRange)
		modifier.StackID = string(source)
		timedModifier := NewTimedModifier(string(ActionNameBuffDamage), user, modifier, turnCount)
		timedModifier.Apply()

		timeKeeper.AddTimed(timedModifier)

		g.logError("Blood for the blood god!")
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Berserker rage for %d turns", turnCount)),
		},
	}
}

func (g *Factory) getApplyStatusAction(args util.Arguments) UntargetedEffect {
	statusEffect := args.Get(0)
	turnCount := args.GetInt(1)
	message := args.Get(2)
	effect := func(user *Actor, source EffectSource) {
		applyStatusAction := NewApplyStatusAction(NewActorCause("ApplyStatusEffect", user), user, StatusEffect(statusEffect), turnCount)
		g.actionBroker.Execute(applyStatusAction)
		if message != "" {
			g.logError(message)
		}
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Apply %s for %d turns", statusEffect, turnCount)),
		},
	}
}

func (g *Factory) getRandomTelefragAction(args util.Arguments) UntargetedEffect {
	effect := func(user *Actor, source EffectSource) {
		gui := g.gui
		currentMap := g.state.GetMap()
		var visEnemies []*Actor
		for _, visPos := range user.GetFOV().Visibles {
			if currentMap.IsActorAt(visPos) {
				actorAt := currentMap.ActorAt(visPos)
				if (actorAt.IsHostileTo(user) || gui().IsPlayerForcingAggression()) && actorAt.IsAlive() && actorAt != user {
					visEnemies = append(visEnemies, actorAt)
				}
			}
		}
		if len(visEnemies) == 0 {
			g.logError("No enemies in sight")
			return
		}
		randomEnemy := visEnemies[g.randomSource.Intn(len(visEnemies))]
		location := randomEnemy.MapPosition()
		isPlayer := user == g.state.GetPlayer()
		moveAction := NewMoveAction(NewActorCause("Telefrag", user), user, location, isPlayer)
		moveAction.DisableAnimation()

		changeHealthAmount := -randomEnemy.GetStats().GetResource(stats.Health).Value()
		killAction := NewChangeResourceAction(NewActorCause("Telefrag", user), randomEnemy, stats.Health, float64(changeHealthAmount))
		g.actionBroker.Execute(killAction)
		g.actionBroker.Execute(moveAction)
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Magically displace a random visible enemy")),
			util.LabelTextFromString(fmt.Sprintf("with your own body, killing it instantly")),
		},
	}
}

func (g *Factory) getRandomTeleportAction(args util.Arguments) UntargetedEffect {
	if args.IsEmpty() {
		g.logError(fmt.Sprintf("Teleport effect requires arguments"))
	}
	distanceInterval := stats.NewIntervalFromString(args.Get(0))

	effect := func(user *Actor, source EffectSource) {
		currentMap := g.state.GetMap()
		origin := user.MapPosition()
		location, found := currentMap.GetRandomLocation(func(location geometry.Point) bool {
			dist := location.ChebyshevDistanceTo(origin)
			return currentMap.IsCurrentlyPassable(location) && dist >= distanceInterval.Min() && dist <= distanceInterval.Max()
		})
		if !found {
			return
		}
		isPlayer := user == g.state.GetPlayer()
		moveAction := NewMoveAction(NewActorCause("Teleport", user), user, location, isPlayer)
		moveAction.DisableAnimation()

		g.actionBroker.QueueForExecution(moveAction)
		g.audioPlayer.PlayCue("teleport")
	}
	return UntargetedEffect{
		Effect: effect,
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Will teleport you %s steps in a random direction", distanceInterval.ToString())),
		},
	}
}

func (g *Factory) getPlayerRecallAction(args util.Arguments) UntargetedEffect {
	return UntargetedEffect{
		Effect: func(user *Actor, source EffectSource) {
			recallAction := func() {
				currentMapName := g.state.GetMap().GetName()
				isInTown := defs.IsTownLevel(currentMapName)

				var dest gridmap.Transition

				if isInTown {
					dest = gridmap.Transition{
						TargetMap:      defs.LevelNameFromDepth(g.state.GetLastDungeonLevel()),
						TargetLocation: "ladder_down",
					}
				} else { // in dungeon
					currentLevel := LevelFromName(currentMapName)
					if g.state.GetLastDungeonLevel() < currentLevel {
						g.state.SetLastDungeonLevel(currentLevel)
					}
					dest = gridmap.Transition{TargetMap: string(defs.LevelNameTownHell), TargetLocation: "spawn"}
				}

				g.state.TransitionTo(dest)
			}
			g.queueAfterTurn(recallAction)
		},
		Description: []util.LabelText{
			util.LabelTextFromString("Recall to town"),
		},
	}
}

/*
Name() string

	GetTooltipLines() []util.LabelText
	InventoryIcon() int32
	TintColor() color.Color
*/
type TimedModifier struct {
	user      *Actor
	modifier  *stats.Modifier
	turnsLeft int
	energy    int
	icon      int32
	name      string
}

func (t *TimedModifier) TintColor() color.RGBA {
	return defs.White
}
func (t *TimedModifier) Name() string {
	return defs.NiceNameForStatusEffect(t.name)
}
func (t *TimedModifier) String() string {
	return defs.NiceNameForStatusEffect(t.name)
}
func (t *TimedModifier) InventoryIcon() int32 {
	return t.icon
}
func (t *TimedModifier) GetTooltipLines(textParser func(string) []util.ColoredTextPart) []util.LabelText {
	return []util.LabelText{
		{
			Text:               defs.DescriptionForStatusEffect(t.name, t.turnsLeft),
			TextColor:          defs.DarkTextColor,
			UseColorCodes:      false,
			TextWithColorCodes: textParser(t.Name()),
		},
	}
}
func (t *TimedModifier) AddEnergy(energy int) {
	t.energy += energy
}

func (t *TimedModifier) RemoveEnergy(used int) {
	t.energy -= used
}

func (t *TimedModifier) IsAlive() bool {
	return t.turnsLeft > 0
}

func (t *TimedModifier) GetBaseSpeed() int {
	return 10
}

func (t *TimedModifier) GetEnergy() int {
	return t.energy
}

func (t *TimedModifier) Turn() {
	if t.turnsLeft == 0 {
		return
	}
	t.turnsLeft--
	if t.turnsLeft == 0 {
		t.Remove()
	}
}

func (t *TimedModifier) Apply() {
	t.user.GetStats().AddModifier(t.modifier)
}

func (t *TimedModifier) Remove() {
	t.user.GetStats().RemoveModifier(t.modifier)
}
func (t *TimedModifier) Counter() int {
	return t.turnsLeft
}
func NewTimedModifier(name string, user *Actor, modifier *stats.Modifier, turns int) *TimedModifier {
	return &TimedModifier{
		name:      name,
		user:      user,
		modifier:  modifier,
		turnsLeft: turns,
	}
}

func (g *Factory) getVisibleEnemiesByDistance(user *Actor) []*Actor {
	currentMap := g.state.GetMap()
	var nearbyEnemies []*Actor
	for _, visPos := range user.GetFOV().Visibles {
		if !currentMap.IsActorAt(visPos) {
			continue
		}
		actor := currentMap.ActorAt(visPos)
		if (actor.IsHostileTo(user) || g.gui().IsPlayerForcingAggression()) && actor.IsAlive() && user != actor {
			nearbyEnemies = append(nearbyEnemies, actor)
		}
	}
	slices.SortStableFunc(nearbyEnemies, func(i, j *Actor) int {
		distI := currentMap.MoveDistance(user.MapPosition(), i.MapPosition())
		distJ := currentMap.MoveDistance(user.MapPosition(), j.MapPosition())
		return cmp.Compare(distI, distJ)
	})
	return nearbyEnemies
}

func (g *Factory) NewConditionFromString(condition string) func(user *Actor) bool {
	if condition == "can_replicate" {
		return g.canReplicate
	}
	if condition == "can_enchant_armor" {
		return g.hasEnchentableArmor
	}
	if condition == "can_enchant_weapon" {
		return g.hasEnchentableWeapon
	}
	g.logError(fmt.Sprintf("no condition found for name %s", condition))
	return func(user *Actor) bool { return true }
}

func (g *Factory) hasEnchentableArmor(user *Actor) bool {
	if !user.GetInventory().HasEnchantableArmor() {
		g.logError("You have no armor to enchant")
		return false
	}
	return true
}


func (g *Factory) hasEnchentableWeapon(user *Actor) bool {
	if !user.GetInventory().HasEnchantableWeapon() {
		g.logError("You have no weapon to enchant")
		return false
	}
	return true
}

func (g *Factory) canReplicate(user *Actor) bool {
	currentMap := g.state.GetMap()

	copyCount := 0
	// check for map limit
	for _, actor := range currentMap.Actors() {
		if actor.GetInternalName() == user.GetInternalName() {
			copyCount++
		}
	}

	if copyCount >= 60 { // currently hardcoded
		return false
	}

	origin := user.MapPosition()
	cloneNeighbors := currentMap.GetFilteredNeighbors(origin, func(p geometry.Point) bool {
		if !currentMap.Contains(p) || !currentMap.IsActorAt(p) {
			return false
		}
		actorAt := currentMap.ActorAt(p)
		return actorAt.GetInternalName() == user.GetInternalName()
	})

	if len(cloneNeighbors) != 1 {
		// conditions for replication not met
		return false
	}

	// find position for spawning
	freePositions := currentMap.GetFreeCellsForDistribution(origin, 1, func(p geometry.Point) bool {
		return currentMap.IsCurrentlyPassable(p)
	})

	if len(freePositions) == 0 { // no space on map
		return false
	}

	return true
}

type EffectSource string
