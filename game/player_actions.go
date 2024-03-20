package game

import (
	"cmp"
	"fmt"
	"math"
	"math/rand"
	"rpg/audio"
	"rpg/common"
	"rpg/defs"
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/stats"
	"rpg/util"
	"slices"
)

type Actions struct {
	player       *common.Actor
	actionBroker common.ActionBroker
	state        *common.State
	animator     common.Animator
	ui           func() common.UserInterface
	config       common.Configuration
	audio        *audio.Player
	definitions  defs.DataContainer

	getTimeKeeper  func() *common.TimeKeeper
	msg            func(message string)
	autoExplore    bool
	playerPath     []geometry.Point
	isResting      bool
	actorSpeak     func(actor *common.Actor, trigger defs.ActorSpeechTrigger)
	queueAfterTurn func(call func())
	playerMoves    []geometry.Point
}

func (a *Actions) WaitOneTurn() {
	state := a.state
	player := state.GetPlayer()

	isMounted := player.HasLeader()

	if isMounted {
		freeManeuverActions := a.maneuverAction(geometry.Point{}, false)
		if len(freeManeuverActions) == 0 && rand.Intn(30) < 0  {
			a.audio.PlayCue("black_death")
		} else {
			for _, action := range freeManeuverActions {
				a.actionBroker.Execute(action)
			}
		}
	}

	a.EndTurn(10)
}

func (a *Actions) MountToggle() {
	state := a.state
	player := state.GetPlayer()

	isMounted := player.HasLeader()

	if isMounted {
		a.unmountCompanion()
	} else {
		a.mountCompanion()
	}

	a.player = a.state.GetPlayer()
	a.updateFoV()
}
func (a *Actions) unmountCompanion() {
	state := a.state
	companion := state.GetPlayer()

	if !companion.IsAlive() {
		return
	}
	player := companion.GetLeader()

	currentMap := state.GetMap()
	freePos := currentMap.GetFreeCellsForDistribution(companion.MapPosition(), 1, func(point geometry.Point) bool {
		return currentMap.CurrentlyPassableAndSafeForActor(player)(point)
	})
	if len(freePos) == 0 {
		a.msg("No space to unmount")
		return
	}
	// add the player actor to the map
	spawnPos := freePos[0]
	currentMap.AddActor(player, spawnPos)
	player.SetDrawPosition(spawnPos.ToPointF())
	state.SetPlayer(player)
	state.SetCompanion(companion)
	a.getTimeKeeper().AddTimed(companion)
	a.ui().UpdateStatsLabel()
}
func (a *Actions) mountCompanion() {
	state := a.state
	player := state.GetPlayer()

	if !player.IsAlive() {
		return
	}
	if !state.HasCompanion() {
		a.msg(fmt.Sprintf("You don't have a companion"))
		return
	}
	companion := state.GetCompanion()
	distance := state.GetMap().MoveDistance(player.MapPosition(), companion.MapPosition())
	if distance > 1 {
		a.msg(fmt.Sprintf("Your companion is too far away"))
		return
	}
	// approach
	// remove the player actor from the map
	currentMap := state.GetMap()
	currentMap.RemoveActor(player)

	state.SetPlayer(companion)
	state.SetCompanion(nil)
	a.getTimeKeeper().Remove(companion)
	companion.SetInventory(player.GetInventory())
	companion.SetEquipment(player.GetEquipment())


	a.ui().UpdateStatsLabel()
}

func (a *Actions) BeginActionTargeting(element *common.TargetedAction) {
	a.beginActionTargeting(element, nil)
}

func (a *Actions) EndTurn(autTaken int) {
	a.actionBroker.Execute(common.NewEndPlayerTurnAction(common.AbstractCause("Wait"), autTaken))
}

func (a *Actions) DropItem(item *common.Item) {
	dropItemAction := common.NewDropAction(common.AbstractCause("DropItem"), a.player, item)
	a.actionBroker.Execute(dropItemAction)
}

func (a *Actions) ActivateUntargetedAction(action *common.UntargetedAction, source common.EffectSource) {
	a.activateUntargetedAction(a.player, action, source)
}
func (a *Actions) beginActionTargeting(action *common.TargetedAction, chargedItem *common.Item) {
	state := a.state
	if !state.GetPlayer().IsAlive() {
		return
	}
	act := a
	act.CancelAutoMove()

	rangedAttackRules := action.GetTargetingRules(state.GetPlayer(), state.GetMap())
	onHit := &common.HitEffect{
		Activate: func(attacker *common.Actor, target geometry.Point, hitPositions []geometry.Point) {
			if chargedItem != nil {
				chargedItem.UseCharge()
				if !chargedItem.HasChargesLeft() && chargedItem.SelfDestructs() {
					state.GetPlayer().GetInventory().Remove(chargedItem)
				}
			}
			act.CancelAutoMove()
			action.OnSelected(attacker, target, hitPositions)
		},
		Description: action.GetDescription(),
	}
	a.ui().BeginTargeting(rangedAttackRules, onHit)
}

func (a *Actions) PickUpItem() {
	state := a.state
	player := state.GetPlayer()
	if !player.IsAlive() || !player.CanPickup() {
		return
	}
	if !state.GetMap().IsItemAt(player.MapPosition()) {
		return
	}
	pickUpAction := common.NewPickUpAction(common.AbstractCause("PickUpAction"), player)
	a.actionBroker.Execute(pickUpAction)
	a.actionBroker.Execute(common.NewEndPlayerTurnAction(pickUpAction.Cause(), pickUpAction.GetDelayInAut()))
}

type PlayerActionsContext struct {
	Player         *common.Actor
	ActionBroker   common.ActionBroker
	State          *common.State
	UI             func() common.UserInterface
	Speak          func(actor *common.Actor, trigger defs.ActorSpeechTrigger)
	GetTimeKeeper  func() *common.TimeKeeper
	Msg            func(message string)
	Animator       common.Animator
	Configuration  common.Configuration
	QueueAfterTurn func(call func())
	AudioPlayer    *audio.Player
}

func NewActions(context PlayerActionsContext) *Actions {
	return &Actions{
		player:         context.Player,
		actionBroker:   context.ActionBroker,
		state:          context.State,
		ui:             context.UI,
		actorSpeak:     context.Speak,
		animator:       context.Animator,
		getTimeKeeper:  context.GetTimeKeeper,
		msg:            context.Msg,
		config:         context.Configuration,
		queueAfterTurn: context.QueueAfterTurn,
		audio: context.AudioPlayer,
	}
}
func (a *Actions) EquipItem(item *common.Item) {
	a.actionBroker.Execute(common.NewEquipItemAction(common.AbstractCause("equip"), a.player, item))
}

func (a *Actions) UnequipItem(item *common.Item) {
	a.actionBroker.Execute(common.NewUnequipItemAction(common.AbstractCause("unequip"), a.player, item))
}

func (a *Actions) ActivateItemActionForPlayer(itemWithAction *common.Item) {
	msg := a.msg
	user := a.player // an AI actor needs to provide a target pos for targeted common.. handle this elsewhere..
	if !user.CanUseItems() || !itemWithAction.HasAction() {
		return
	}

	if !itemWithAction.HasChargesLeft() {
		msg(fmt.Sprintf("%s is out of charges", itemWithAction.Name()))
		return
	}

	if itemWithAction.HasTargetedAction() {
		a.beginActionTargeting(itemWithAction.GetTargetedAction(), itemWithAction)
	} else if itemWithAction.HasUntargetedAction() {

		itemWithAction.UseCharge()
		if !itemWithAction.HasChargesLeft() && itemWithAction.SelfDestructs() {
			user.GetInventory().Remove(itemWithAction)
		}

		untargetedAction := itemWithAction.GetUntargetedAction()
		a.activateUntargetedAction(user, untargetedAction, common.EffectSource(itemWithAction.GetInternalName()))
		/* Actions should end the turn themselves
		autTaken := untargetedAction.GetTimeNeededInAut()
		if autTaken > 0 {
			a.actionBroker.Execute(common.NewEndPlayerTurnAction(common.AbstractCause("ActiveItemUse"), autTaken))
		}
		 */
	}
}

func (a *Actions) activateUntargetedAction(actor *common.Actor, action *common.UntargetedAction, source common.EffectSource) {
	if action.IsDelayed() {
		// add to timekeeper
		action.Reset()
		a.msg(fmt.Sprintf("%s is delayed for %d turns", action.Name(), action.GetDelayInTurns()))
		action.SetCurrentUserAndSource(actor, source)
		timeKeeper := a.getTimeKeeper()
		common.RemoveTimedEffectsBySource(timeKeeper, source)
		timeKeeper.AddTimed(action)
	} else if action.CanActivate(actor) {
		action.Invoke(actor, source)
	}
}
func (a *Actions) playerThrow(thrownItem *common.Item) func(player *common.Actor, targetPos geometry.Point) {
	return func(player *common.Actor, targetPos geometry.Point) {
		throwAction := common.NewThrowAction(common.NewActorCause("Throw", player), player, thrownItem, targetPos)
		a.actionBroker.Execute(throwAction)
		a.actionBroker.Execute(common.NewEndPlayerTurnAction(throwAction.Cause(), throwAction.GetDelayInAut()))
	}
}
func (a *Actions) GetContextActionsForItem(item *common.Item) []util.MenuItem {
	// TODO:-..
	var menuForActions []util.MenuItem
	if item.HasUntargetedAction() {
		menuForActions = append(menuForActions, util.MenuItem{
			MainText: item.GetUntargetedAction().Name(),
			Action: func() {
				a.ActivateItemActionForPlayer(item)
			},
		})
	}
	if item.HasTargetedAction() {
		menuForActions = append(menuForActions, util.MenuItem{
			MainText: item.GetTargetedAction().Name(),
			Action: func() {
				a.ActivateItemActionForPlayer(item)
			},
		})
	}
	if item.IsThrowable() {
		menuForActions = append(menuForActions, util.MenuItem{
			MainText: "Throw",
			Action: func() {
				//a.
			},
		})
	}
	return menuForActions
}

func (a *Actions) ManualMovePlayer(direction geometry.Point) {
	state := a.state

	if state.GetFlags().IsSet("player_fell_into_chasm") {
		return
	}

	ui := a.ui()
	player := state.GetPlayer()
	if !player.IsAlive() {
		return
	}
	ui.CancelTargeting()
	a.CancelAutoMove()
	//g.actionBroker.CancelAllAnimations()
	if player.HasStatusEffect(common.StatusConfused) {
		if rand.Intn(3) != 0 { // one in three chance to succeed with movement
			// pick a different random direction..
			possibleDirectionsRemaining := state.GetMap().GetFilteredDirectionsForMovement(func(dir geometry.Point) bool {
				if dir == direction {
					return false
				}
				return true
			})
			direction = possibleDirectionsRemaining[rand.Intn(len(possibleDirectionsRemaining))]
		}
	}

	playerAction := common.PlayerMoveToAttack(player, direction, a.ui().IsPlayerForcingAggression(), a.getActionContext())
	if playerAction == nil {
		return
	}

	destination := player.MapPosition().Add(direction)

	a.actionBroker.Execute(playerAction)

	didMove := player.MapPosition() == destination
	freeManeuverActions := a.maneuverAction(direction, didMove)
	if player.HasLeader() && len(freeManeuverActions) == 0 && rand.Intn(30) == 0 {
		a.audio.PlayCue("black_death")
	} else {
		for _, action := range freeManeuverActions {
			a.actionBroker.Execute(action)
		}
	}

	a.actionBroker.Execute(common.NewEndPlayerTurnAction(playerAction.Cause(), playerAction.GetDelayInAut()))
}

func (a *Actions) getActionContext() common.ActionContext {
	return common.ActionContext{
		GameState:     a.state,
		UserInterface: a.ui,
		AudioPlayer:   a.audio,
		Data:          a.definitions,
		ActionBroker:  a.actionBroker,
		Speak:         a.actorSpeak,
	}
}

func (a *Actions) nextStepOnPlayerPath() {
	state := a.state
	msg := a.msg
	player := state.GetPlayer()
	if !player.IsAlive() {
		a.stopAutoMove()
		return
	}
	if a.IsPlayerAutoMoving() {
		if state.GetVisibility().GetVisibleEnemyCount(func(actor *common.Actor) bool {
			return actor.IsHostileTo(player)
		}) > 0 {
			a.stopAutoMove()
			msg("You can see an enemy")
			return
		}

		// pop first element
		nextStep := a.playerPath[0]

		currentMap := state.GetMap()
		if currentMap.IsActorAt(nextStep) {
			actorAt := currentMap.ActorAt(nextStep)
			if actorAt.IsAlly(player) {
				// SWAP
				a.playerPath = a.playerPath[1:]
				autoMoveAction := common.NewSwapPositionsAction(common.AbstractCause("Player Movement"), player, actorAt, true)
				autoMoveAction.SetAnimationDelay(0.08)
				a.actionBroker.QueueForExecution(autoMoveAction)
				a.actionBroker.QueueForExecution(common.NewEndPlayerTurnAction(autoMoveAction.Cause(), autoMoveAction.GetDelayInAut()))
				return
			} else {
				a.stopAutoMove()
				if !a.autoExplore { // never explore into another actor, only deliberately
					direction := nextStep.Sub(player.MapPosition())
					a.ManualMovePlayer(direction)
				}
				return
			}
		}

		if currentMap.IsObjectAt(nextStep) {
			if a.state.IsClosedDoorAt(nextStep) {
				openDoorAction := common.NewOpenDoorAction(common.AbstractCause("Player Movement"), player, nextStep)
				a.actionBroker.QueueForExecution(openDoorAction)
				a.actionBroker.QueueForExecution(common.NewEndPlayerTurnAction(openDoorAction.Cause(), openDoorAction.GetDelayInAut()))
				return
			} else if !a.autoExplore {
				objectAt := currentMap.ObjectAt(nextStep)
				if !objectAt.IsWalkable(player) {
					a.stopAutoMove()
					msg(fmt.Sprintf("You bump into %s", objectAt.Name()))
					objectAt.ActorBumpedInto(player, common.AnimateNow())
					return
				}
			}
		}

		if !currentMap.IsCurrentlyPassable(nextStep) {
			a.stopAutoMove()
			msg("You can't move there")
			return
		}

		a.playerPath = a.playerPath[1:]
		autoMoveAction := common.NewMoveAction(common.AbstractCause("Player Movement"), player, nextStep, true)
		autoMoveAction.SetAnimationDelay(0.08)
		a.actionBroker.QueueForExecution(autoMoveAction)
		a.actionBroker.QueueForExecution(common.NewEndPlayerTurnAction(autoMoveAction.Cause(), autoMoveAction.GetDelayInAut()))
	} else if a.autoExplore {
		if target, hasTarget := a.getNextAutoExploreTarget(); hasTarget {
			a.continuePlayerAutoMovement(target)
		} else {
			msg("Nothing left to explore")
			a.stopAutoMove()
		}
	}
}
func (a *Actions) continuePlayerAutoMovement(targetPos geometry.Point) {
	currentMap := a.state.GetMap()
	var isWalkable func(point geometry.Point) bool
	if a.autoExplore {
		isWalkable = a.allowedForAutoExplore
	} else {
		isWalkable = a.allowedForAutoMove(targetPos)
	}
	pathToTarget := currentMap.GetJPSPath(a.state.GetPlayer().MapPosition(), targetPos, isWalkable)
	if len(pathToTarget) <= 1 {
		return
	}
	a.playerPath = pathToTarget[1:]
	a.nextStepOnPlayerPath()
}
func (a *Actions) allowedForAutoExplore(point geometry.Point) bool {
	if a.state.HasCompanion() && a.state.GetCompanion().MapPosition() == point {
		return true
	}
	currentMap := a.state.GetMap()
	isWalkable := currentMap.CurrentlyPassableAndSafeForActor(a.state.GetPlayer())

	return isWalkable(point) && !currentMap.IsStairsAt(point)
}
func (a *Actions) allowedForAutoMove(targetPos geometry.Point) func(point geometry.Point) bool {
	return func(point geometry.Point) bool {
		if a.state.HasCompanion() && a.state.GetCompanion().MapPosition() == point {
			return true
		}
		currentMap := a.state.GetMap()
		isWalkable := currentMap.CurrentlyPassableAndSafeForActor(a.state.GetPlayer())
		return isWalkable(point) || point == targetPos
	}
}

func (a *Actions) StartPlayerAutoMovement(targetPos geometry.Point) {
	player := a.state.GetPlayer()
	if !player.IsAlive() {
		return
	}
	currentMap := a.state.GetMap()

	// single step, like a manual move..
	if currentMap.MoveDistance(player.MapPosition(), targetPos) == 1 {
		// this seems to be an intended action..
		direction := targetPos.Sub(player.MapPosition())
		a.ManualMovePlayer(direction)
		return
	}

	if len(a.state.GetVisibility().GetVisibleEnemies(player)) > 0 {
		// player is trying to move while enemies are visible..
		a.msg("You can't run while enemies are visible")
		return
	}

	if !player.CanAutoMove() {
		a.msg("You can't move in a coordinated fashion right now")
		return
	}

	playerPosition := player.MapPosition()

	// check for item directly beneath us.. if so, pick it up
	if a.config.AutoPickup {
		if currentMap.IsItemAt(playerPosition) {
			itemAt := currentMap.ItemAt(playerPosition)
			if !itemAt.IsDroppedByPlayer() {
				pickUpAction := common.NewPickUpAction(common.AbstractCause("PickUp"), player)
				a.actionBroker.Execute(pickUpAction)
				a.actionBroker.Execute(common.NewEndPlayerTurnAction(pickUpAction.Cause(), pickUpAction.GetDelayInAut()))
			}
		}
	}

	var isWalkable func(point geometry.Point) bool
	if a.autoExplore {
		isWalkable = a.allowedForAutoExplore
	} else {
		isWalkable = a.allowedForAutoMove(targetPos)
	}
	pathToTarget := currentMap.GetJPSPath(playerPosition, targetPos, isWalkable)
	if len(pathToTarget) <= 1 {
		a.stopAutoMove()
		return
	}
	a.animator.CancelAllAnimations()
	a.actionBroker.ClearQueuedActions(func(action common.Action) bool {
		if moveAction, ok := action.(*common.MoveAction); ok {
			return moveAction.GetActor() == player
		}
		return false
	})
	a.playerPath = pathToTarget[1:]
	a.nextStepOnPlayerPath()
}

func (a *Actions) getNextAutoExploreTarget() (geometry.Point, bool) {
	state := a.state
	player := state.GetPlayer()
	playerInv := player.GetInventory()
	playerPos := player.MapPosition()

	if !playerInv.IsFull() && a.config.AutoPickup {
		nearestItem := state.GetMap().GetNearestExploredItem(playerPos, func(item *common.Item) bool {
			return !item.IsDroppedByPlayer() && !item.IsHidden()
		})
		if nearestItem != playerPos {
			return nearestItem, true
		}
	}

	unexplored := state.GetMap().GetNearestUnexploredWalkablePosition(playerPos)
	if unexplored != playerPos {
		return unexplored, true
	}
	return geometry.Point{}, false
}

func (a *Actions) CancelAutoMove() {
	a.stopAutoMove()
	a.animator.CancelAllAnimations()

}
func (a *Actions) stopAutoMove() {
	state := a.state
	a.actionBroker.ClearQueuedActions(func(action common.Action) bool {
		if moveAction, ok := action.(*common.MoveAction); ok {
			return moveAction.GetActor() == state.GetPlayer()
		}
		return false
	})

	a.autoExplore = false
	a.playerPath = nil
}

func (a *Actions) StartAutoExplore() {
	if !a.state.GetPlayer().CanAutoMove() {
		a.msg("You can't explore right now")
		return
	}
	target, hasTarget := a.getNextAutoExploreTarget()
	if a.autoExplore {
		a.stopAutoMove()
		a.msg("You stop exploring")
		return
	}
	if hasTarget {
		a.msg("You start exploring")
		a.autoExplore = true
		a.StartPlayerAutoMovement(target)
	} else {
		a.msg("Nothing to explore")
	}
}

func (a *Actions) goOneLevelUp() {
	state := a.state
	if !state.GetPlayer().IsAlive() {
		return
	}
	currentMap := state.GetMap()
	mapName := currentMap.GetName()
	if mapName == string(defs.LevelNameTown) {
		return
	}

	ourPos := state.GetPlayer().MapPosition()
	nearestStairsUp := currentMap.GetNearestTile(ourPos, func(tile gridmap.Tile) bool {
		return tile.IsStairsUp()
	})

	if nearestStairsUp == ourPos {
		a.msg("No known path")
		return
	}
	a.CancelAutoMove()
	a.StartPlayerAutoMovement(nearestStairsUp)
}

func (a *Actions) goOneLevelDown() {
	state := a.state
	if !state.GetPlayer().IsAlive() {
		return
	}
	currentMap := state.GetMap()
	//mapName := currentMap.Action()

	ourPos := state.GetPlayer().MapPosition()
	nearestStairsDown := currentMap.GetNearestTile(ourPos, func(tile gridmap.Tile) bool {
		return tile.IsStairsDown()
	})
	if nearestStairsDown == ourPos {
		a.msg("No known path")
		return
	}
	a.CancelAutoMove()
	a.StartPlayerAutoMovement(nearestStairsDown)
}

func (a *Actions) QuickRangedAttack() {
	state := a.state
	if !state.GetPlayer().IsAlive() {
		return
	}
	if !state.GetPlayer().GetEquipment().HasRangedWeaponEquipped() {
		a.msg("You don't have a ranged weapon equipped")
		return
	}
	rangedAttackRuleProvider := common.NewRangedAttackTargetingRules(util.Arguments{})
	rangedAttackRules := rangedAttackRuleProvider.GetTargetRules(state.GetPlayer(), common.OneTargetTile(state.GetMap, util.Arguments{}), state.GetMap())
	target := rangedAttackRules.GetPreSelectedTarget()
	if target == state.GetPlayer().MapPosition() {
		a.msg("You can't attack yourself like that")
		return
	}
	a.playerRangedAttack(state.GetPlayer(), target)
	return
}

func (a *Actions) AimedRangedAttack() {
	state := a.state
	if !state.GetPlayer().IsAlive() {
		return
	}
	if !state.GetPlayer().GetEquipment().HasRangedWeaponEquipped() {
		a.msg("You don't have a ranged weapon equipped")
		return
	}
	a.defaultRangedTargeting(a.playerRangedAttack) // execute ranged attack action and end turn)
	return
}
func (a *Actions) defaultRangedTargeting(onTargetSelected func(player *common.Actor, targetPos geometry.Point)) {
	state := a.state
	rangedAttackRuleProvider := common.NewRangedAttackTargetingRules(util.Arguments{})
	rangedAttackRules := rangedAttackRuleProvider.GetTargetRules(state.GetPlayer(), common.OneTargetTile(state.GetMap, util.Arguments{}), state.GetMap())
	a.beginRangedTargeting(rangedAttackRules, onTargetSelected)
}
func (a *Actions) beginRangedTargeting(targetRules common.TargetRules, onSelected func(attacker *common.Actor, target geometry.Point)) {
	state := a.state
	if !state.GetPlayer().IsAlive() {
		return
	}
	onHit := &common.HitEffect{
		Activate: func(attacker *common.Actor, target geometry.Point, hitPositions []geometry.Point) {
			origin := state.GetPlayer().MapPosition()
			if origin == target {
				a.msg("You can't attack yourself like that")
				return
			}
			onSelected(attacker, target)
		},
		Description: []util.LabelText{
			util.LabelTextFromString(fmt.Sprintf("Ranged attack")),
		},
	}
	a.ui().BeginTargeting(targetRules, onHit)
}

func (a *Actions) UseItem() {
	state := a.state
	player := state.GetPlayer()
	if !player.IsAlive() {
		return
	}

	inventory := player.GetInventory()

	userInterface := a.ui()
	useFilter := func(element *common.Item) bool {
		return element.HasAction()
	}
	onSelected := func(itemKey rune) {
		userInterface.PopAll()
		useItem := inventory.GetItemByLetter(itemKey)
		a.ActivateItemActionForPlayer(useItem)
	}
	userInterface.OpenListInventoryForSelection(useFilter, onSelected)
}

func (a *Actions) AimedThrow() {
	state := a.state
	player := state.GetPlayer()
	if !player.IsAlive() {
		return
	}

	equipment := player.GetEquipment()
	inventory := player.GetInventory()
	if !equipment.HasThrowableQuivered() {
		userInterface := a.ui()
		throwFilter := func(element *common.Item) bool {
			return element.IsThrowable()
		}
		onSelected := func(itemKey rune) {
			userInterface.PopAll()
			thrownItem := inventory.GetItemByLetter(itemKey)
			throwComponent := thrownItem.GetThrowStats()
			targetRulesProvider := common.NewRangedAttackTargetingRules(util.Arguments{"throw"})
			rangedAttackRules := targetRulesProvider.GetTargetRules(player, throwComponent.HitZone, state.GetMap())
			a.beginRangedTargeting(rangedAttackRules, a.playerThrow(thrownItem))
		}
		userInterface.OpenListInventoryForSelection(throwFilter, onSelected)
	} else {
		thrownItem := equipment.GetNextQuiveredThrowable()
		throwComponent := thrownItem.GetThrowStats()
		targetRulesProvider := common.NewRangedAttackTargetingRules(util.Arguments{"throw"})
		rangedAttackRules := targetRulesProvider.GetTargetRules(player, throwComponent.HitZone, state.GetMap())

		a.beginRangedTargeting(rangedAttackRules, a.playerThrow(thrownItem))
	}
}

func (a *Actions) AutoAttack() {
	a.stopAutoMove()
	state := a.state
	gui := a.ui()
	player := state.GetPlayer()
	if !player.IsAlive() {
		return
	}
	var visibleEnemies []*common.Actor

	if gui.IsPlayerForcingAggression() {
		for actor, _ := range state.GetVisibility().GetVisibleActors() {
			visibleEnemies = append(visibleEnemies, actor)
		}
	} else {
		visibleEnemies = state.GetVisibility().GetVisibleEnemies(player)
	}

	if len(visibleEnemies) == 0 {
		return
	}

	hasRanged := player.GetEquipment().HasRangedWeaponEquipped()
	hasMelee := player.GetEquipment().HasMeleeWeaponEquipped()
	slices.SortStableFunc(visibleEnemies, func(i, j *common.Actor) int {
		distI := i.MapPosition().ManhattanDistanceTo(player.MapPosition())
		distJ := j.MapPosition().ManhattanDistanceTo(player.MapPosition())
		return cmp.Compare(distI, distJ)
	})
	currentMap := state.GetMap()
	for _, nearestEnemy := range visibleEnemies {
		dist := currentMap.MoveDistance(player.MapPosition(), nearestEnemy.MapPosition())
		if dist <= 1 {
			if hasMelee || !hasRanged {
				a.playerMeleeAttack(player, nearestEnemy)
				return
			} else if hasRanged {
				a.playerRangedAttack(player, nearestEnemy.MapPosition())
				return
			}
		}

		if hasRanged {
			a.playerRangedAttack(player, nearestEnemy.MapPosition())
			return
		}

		// do melee
		freeAttackPositions := currentMap.GetFilteredNeighborsForMovement(nearestEnemy.MapPosition(), func(point geometry.Point) bool {
			return currentMap.IsCurrentlyPassable(point)
		})

		if len(freeAttackPositions) == 0 {
			continue
		}
		distMap := make([]util.Tuple[geometry.Point, int], len(freeAttackPositions))
		for _, freeAttackPosition := range freeAttackPositions {
			distanceTo := currentMap.MoveDistance(player.MapPosition(), freeAttackPosition)
			distMap = append(distMap, util.NewTuple(freeAttackPosition, distanceTo))
		}
		slices.SortStableFunc(distMap, func(i, j util.Tuple[geometry.Point, int]) int {
			return cmp.Compare(i.GetItem2(), j.GetItem2())
		})

		for _, distTuple := range distMap {
			targetPos := distTuple.GetItem1()
			walkPath := currentMap.GetJPSPath(player.MapPosition(), targetPos, currentMap.CurrentlyPassableAndSafeForActor(player))
			if len(walkPath) <= 1 {
				continue
			}
			// remove first pos
			walkPath = walkPath[1:]
			firstStep := walkPath[0]

			a.playerMoveOneStep(player, firstStep)
			return
		}
	}

}

func (a *Actions) getRangedAttackRules(player *common.Actor) *common.RangedTargetingRules {
	state := a.state
	targetRulesProvider := common.NewRangedAttackTargetingRules(util.Arguments{})
	targetRules := targetRulesProvider.GetTargetRules(player, common.OneTargetTile(state.GetMap, util.Arguments{}), state.GetMap())
	rangedAttackRules := targetRules.(*common.RangedTargetingRules)
	return rangedAttackRules
}

func (a *Actions) playerMoveOneStep(player *common.Actor, targetPos geometry.Point) {
	moveAction := common.NewMoveAction(common.NewActorCause("Move", player), player, targetPos, true)
	a.actionBroker.Execute(moveAction)
	a.actionBroker.Execute(common.NewEndPlayerTurnAction(moveAction.Cause(), moveAction.GetDelayInAut()))
}
func (a *Actions) playerMeleeAttack(player *common.Actor, defender *common.Actor) {
	meleeAction := common.NewMeleeAttackAction(common.NewActorCause("Melee attack", player), player, defender)
	a.actionBroker.Execute(meleeAction)
	a.actionBroker.Execute(common.NewEndPlayerTurnAction(meleeAction.Cause(), meleeAction.GetDelayInAut()))
}
func (a *Actions) playerRangedAttack(player *common.Actor, targetPos geometry.Point) {
	rangedAttackAction := common.NewRangedAttackAction(common.NewActorCause("Ranged attack", player), player, targetPos)
	a.actionBroker.Execute(rangedAttackAction)
	a.actionBroker.Execute(common.NewEndPlayerTurnAction(rangedAttackAction.Cause(), rangedAttackAction.GetDelayInAut()))
}


func (a *Actions) SetQueueAfterTurn(afterTurn func(call func())) {
	a.queueAfterTurn = afterTurn
}
func (a *Actions) ShowActorInfoAt(mapPos geometry.Point) {
	state := a.state
	player := state.GetPlayer()
	actor := state.GetMap().ActorAt(mapPos)
	infoText := util.ToLabelText(actor.GetInfoText())
	if !actor.IsVendor() && actor.GetFaction() != player.GetFaction() {
		otherHasRanged := actor.GetEquipment().HasRangedWeaponEquipped()
		playerHasRanged := player.GetEquipment().HasRangedWeaponEquipped()

		otherMeleeDamage := actor.GetDamageDice(common.MeleeAttack, "")

		expectedMeleeDamage := otherMeleeDamage.ExpectedValue()
		maxMeleeDamage := otherMeleeDamage.Max()

		expectedMeleeDamageReduced := common.CalculateMeleeDamage(expectedMeleeDamage, 1.0, actor, player)
		maxMeleeDamageReduced := common.CalculateMeleeDamage(maxMeleeDamage, 1.0, actor, player)

		if maxMeleeDamageReduced == 0 {
			infoText = append(infoText, util.LabelTextFromString(""),
				util.LabelTextFromStringWithColor("This creature will not be able to penetrate", defs.Red),
				util.LabelTextFromStringWithColor("your armor with a regular attack.", defs.Red),
			)
		} else if expectedMeleeDamageReduced == 0 {
			infoText = append(infoText, util.LabelTextFromString(""),
				util.LabelTextFromStringWithColor("It will be hard for this creature", defs.Red),
				util.LabelTextFromStringWithColor("to penetrate your armor with a regular attack.", defs.Red),
			)
			infoText = append(infoText, util.LabelTextFromString(""), util.LabelTextFromStringWithColor(fmt.Sprintf("Enemy could at worst hit for %d damage per melee attack,", maxMeleeDamageReduced), defs.Red))
			infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("that's %d%% of your current health,", maxMeleeDamageReduced*100/player.GetStats().GetResource(stats.Health).Value()), defs.Red))
			hitsForMeleeKill := int(math.Ceil(float64(player.GetStats().GetResource(stats.Health).Value()) / float64(maxMeleeDamageReduced)))
			infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("and at worst could defeat you in %d hits.", hitsForMeleeKill), defs.Red))
		} else {
			infoText = append(infoText, util.LabelTextFromString(""), util.LabelTextFromStringWithColor(fmt.Sprintf("Enemy typically hits for %d damage per melee attack,", expectedMeleeDamageReduced), defs.Red))
			infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("that's %d%% of your current health,", expectedMeleeDamageReduced*100/player.GetStats().GetResource(stats.Health).Value()), defs.Red))
			hitsForMeleeKill := int(math.Ceil(float64(player.GetStats().GetResource(stats.Health).Value()) / float64(maxMeleeDamageReduced)))
			infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("and at worst could defeat you in %d hits.", hitsForMeleeKill), defs.Red))
		}

		if otherHasRanged {
			otherRangedDamage := actor.GetDamageDice(common.RangedAttack, "")

			expectedRangedDamage := otherRangedDamage.ExpectedValue()
			maxRangedDamage := otherRangedDamage.Max()

			expectedRangedDamageReduced := common.CalculateRangedDamage(expectedRangedDamage, 1.0, actor, player)
			maxRangedDamageReduced := common.CalculateRangedDamage(maxRangedDamage, 1.0, actor, player)

			if maxRangedDamageReduced == 0 {
				infoText = append(infoText, util.LabelTextFromString(""),
					util.LabelTextFromStringWithColor("This creature will not be able to penetrate", defs.Red),
					util.LabelTextFromStringWithColor("your armor with a ranged attack.", defs.Red),
				)
			} else if expectedRangedDamageReduced == 0 {
				infoText = append(infoText, util.LabelTextFromString(""),
					util.LabelTextFromStringWithColor("It will be hard for this creature", defs.Red),
					util.LabelTextFromStringWithColor("to penetrate your armor with a ranged attack.", defs.Red),
				)
				infoText = append(infoText, util.LabelTextFromString(""), util.LabelTextFromStringWithColor(fmt.Sprintf("Enemy could at worst hit for %d damage per ranged attack,", maxRangedDamageReduced), defs.Red))
				infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("that's %d%% of your current health,", maxRangedDamageReduced*100/player.GetStats().GetResource(stats.Health).Value()), defs.Red))
				hitsForRangedKill := int(math.Ceil(float64(player.GetStats().GetResource(stats.Health).Value()) / float64(maxRangedDamageReduced)))
				infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("and at worst could defeat you in %d hits.", hitsForRangedKill), defs.Red))
			} else {
				infoText = append(infoText, util.LabelTextFromString(""), util.LabelTextFromStringWithColor(fmt.Sprintf("Enemy typically hits for %d damage per ranged attack,", expectedRangedDamageReduced), defs.Red))
				infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("that's %d%% of your current health,", expectedRangedDamageReduced*100/player.GetStats().GetResource(stats.Health).Value()), defs.Red))
				hitsForRangedKill := int(math.Ceil(float64(player.GetStats().GetResource(stats.Health).Value()) / float64(maxRangedDamageReduced)))
				infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("and at worst could defeat you in %d hits.", hitsForRangedKill), defs.Red))
			}
		}

		playerMeleeDamage := player.GetDamageDice(common.MeleeAttack, actor.GetInternalName())

		playerExpectedMeleeDamage := playerMeleeDamage.ExpectedValue()
		playerMaxMeleeDamage := playerMeleeDamage.Max()

		playerExpectedMeleeDamageReduced := common.CalculateMeleeDamage(playerExpectedMeleeDamage, 1.0, player, actor)
		playerMaxMeleeDamageReduced := common.CalculateMeleeDamage(playerMaxMeleeDamage, 1.0, player, actor)


		if playerMaxMeleeDamageReduced == 0 {
			infoText = append(infoText, util.LabelTextFromString(""),
				util.LabelTextFromStringWithColor("You will not be able to penetrate", defs.DarkGreen),
				util.LabelTextFromStringWithColor("the enemy's armor with a regular attack.", defs.DarkGreen),
			)
		} else if playerExpectedMeleeDamageReduced == 0 {
			infoText = append(infoText, util.LabelTextFromString(""),
				util.LabelTextFromStringWithColor("It will be hard for you", defs.DarkGreen),
				util.LabelTextFromStringWithColor("to penetrate the enemy's armor with a regular attack.", defs.DarkGreen),
			)
			infoText = append(infoText, util.LabelTextFromString(""), util.LabelTextFromStringWithColor(fmt.Sprintf("You could at best hit for %d damage per melee attack,", playerMaxMeleeDamageReduced), defs.DarkGreen))
			infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("that's %d%% of the enemy's current health,", playerMaxMeleeDamageReduced*100/actor.GetStats().GetResource(stats.Health).Value()), defs.DarkGreen))
			hitsForMeleeKill := int(math.Ceil(float64(actor.GetStats().GetResource(stats.Health).Value()) / float64(playerMaxMeleeDamageReduced)))
			infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("and at best could defeat the enemy in %d hits.", hitsForMeleeKill), defs.DarkGreen))
		} else {
			infoText = append(infoText, util.LabelTextFromString(""), util.LabelTextFromStringWithColor(fmt.Sprintf("You typically hit for %d damage per melee attack,", playerExpectedMeleeDamageReduced), defs.DarkGreen))
			infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("that's %d%% of the enemy's current health,", playerExpectedMeleeDamageReduced*100/actor.GetStats().GetResource(stats.Health).Value()), defs.DarkGreen))
			hitsForMeleeKill := int(math.Ceil(float64(actor.GetStats().GetResource(stats.Health).Value()) / float64(playerMaxMeleeDamageReduced)))
			infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("and at best could defeat the enemy in %d hits.", hitsForMeleeKill), defs.DarkGreen))
		}

		if playerHasRanged {

			playerHitChance := common.CalculateRangedWeaponHitChance(player, actor)

			infoText = append(infoText, util.LabelTextFromString(""), util.LabelTextFromStringWithColor(fmt.Sprintf("You have a %d%% chance to hit with a ranged attack.", int(playerHitChance*100)), defs.DarkGreen))

			playerRangedDamage := player.GetDamageDice(common.RangedAttack, actor.GetInternalName())

			playerExpectedRangedDamage := playerRangedDamage.ExpectedValue()
			playerMaxRangedDamage := playerRangedDamage.Max()

			playerExpectedRangedDamageReduced := common.CalculateRangedDamage(playerExpectedRangedDamage, 1.0, player, actor)
			playerMaxRangedDamageReduced := common.CalculateRangedDamage(playerMaxRangedDamage, 1.0, player, actor)

			if playerMaxRangedDamageReduced == 0 {
				infoText = append(infoText, util.LabelTextFromString(""),
					util.LabelTextFromStringWithColor("You will not be able to penetrate", defs.DarkGreen),
					util.LabelTextFromStringWithColor("the enemy's armor with a ranged attack.", defs.DarkGreen),
				)
			} else if playerExpectedRangedDamageReduced == 0 {
				infoText = append(infoText, util.LabelTextFromString(""),
					util.LabelTextFromStringWithColor("It will be hard for you", defs.DarkGreen),
					util.LabelTextFromStringWithColor("to penetrate the enemy's armor with a ranged attack.", defs.DarkGreen),
				)
				infoText = append(infoText, util.LabelTextFromString(""), util.LabelTextFromStringWithColor(fmt.Sprintf("You could at best hit for %d damage per ranged attack,", playerMaxRangedDamageReduced), defs.DarkGreen))
				infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("that's %d%% of the enemy's current health,", playerMaxRangedDamageReduced*100/actor.GetStats().GetResource(stats.Health).Value()), defs.DarkGreen))
				hitsForRangedKill := int(math.Ceil(float64(actor.GetStats().GetResource(stats.Health).Value()) / float64(playerMaxRangedDamageReduced)))
				infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("and at best could defeat the enemy in %d hits.", hitsForRangedKill), defs.DarkGreen))
			} else {
				infoText = append(infoText, util.LabelTextFromString(""), util.LabelTextFromStringWithColor(fmt.Sprintf("You typically hit for %d damage per ranged attack,", playerExpectedRangedDamageReduced), defs.DarkGreen))
				infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("that's %d%% of the enemy's current health,", playerExpectedRangedDamageReduced*100/actor.GetStats().GetResource(stats.Health).Value()), defs.DarkGreen))
				hitsForRangedKill := int(math.Ceil(float64(actor.GetStats().GetResource(stats.Health).Value()) / float64(playerMaxRangedDamageReduced)))
				infoText = append(infoText, util.LabelTextFromStringWithColor(fmt.Sprintf("and at best could defeat the enemy in %d hits.", hitsForRangedKill), defs.DarkGreen))
			}
		}
	}

	a.ui().OpenTextWindow(infoText)
}

func (a *Actions) updateFoV() {
	state := a.state
	player := state.GetPlayer()
	a.actionBroker.Execute(common.NewUpdateFoVAction(common.AbstractCause("FoV Update"), player))
}

func (a *Actions) IsPlayerAutoMoving() bool {
	return len(a.playerPath) > 0
}

func (a *Actions) maneuverAction(direction geometry.Point, playerDidMove bool) []common.Action {
	var zeroDir geometry.Point
	state := a.state

	player := state.GetPlayer()
	isMounted := player.HasLeader()

	if !isMounted {
		return nil
	}

	isWaitAction := direction == zeroDir

	if playerDidMove || isWaitAction {
		a.playerMoves = append(a.playerMoves, direction)
	} else {
		// this is a bump into something..
		// if we trigger our pounding move, then this is a valid input.
		// otherwise we'll reset our moves
		if len(a.playerMoves) < 2 {
			a.playerMoves = nil
			return nil
		}

		if a.playerMoves[0] != a.playerMoves[1] || a.playerMoves[0] != direction {
			a.playerMoves = nil
			return nil
		}

		currentMap := state.GetMap()
		currentPlayerPos := player.MapPosition()
		poundPos := currentPlayerPos.Add(direction)
		if !currentMap.IsActorAt(poundPos) {
			a.playerMoves = nil
			return nil
		}

		hitActor := currentMap.ActorAt(poundPos)

		currentMap.SetActorToDowned(hitActor) // must this be an action?
		a.msg(fmt.Sprintf("You pound %s to the ground", hitActor.Name()))
		return []common.Action{
			common.NewApplyStatusAction(common.AbstractCause("Knockback"), hitActor, common.StatusStunned, rand.Intn(3)+2),
			//common.NewChangeResourceAction(common.AbstractCause("Knockback"), hitActor, stats.Health, -4),
		}
	}

	if len(a.playerMoves) < 2 {
		return nil
	}
	if len(a.playerMoves) > 2 {
		a.playerMoves = a.playerMoves[1:]
	}
	lastDir := a.playerMoves[len(a.playerMoves)-1]
	nextToLastDir := a.playerMoves[len(a.playerMoves)-2]

	currentMap := state.GetMap()
	if lastDir == zeroDir && nextToLastDir == zeroDir {
		hitPositions := currentMap.NeighborsAll(player.MapPosition(), func(p geometry.Point) bool {
			return currentMap.IsActorAt(p)
		})
		if len(hitPositions) > 0 {
			a.msg("The hooves of your steed clap, like thunder!")
			damage := rand.Intn(5) + 3
			return []common.Action{
				common.NewLightningAction(common.AbstractCause("Thunderclap"), player, hitPositions, damage),
			}
		}
	}

	if isWaitAction {
		a.playerMoves = nil
		return nil
	}
	// patterns
	// wait, wait
	// direction, reversed-direction
	// direction, direction

	if lastDir.IsOpposite(nextToLastDir) {
		currentPlayerPos := player.MapPosition()
		kickPos := currentPlayerPos.Add(nextToLastDir.Mul(2))
		if currentMap.IsActorAt(kickPos) {
			hitActor := currentMap.ActorAt(kickPos)
			a.msg(fmt.Sprintf("You kick %s", hitActor.Name()))
			a.audio.PlayCue("black_death_kick")
			return []common.Action{
				common.NewKnockbackAction(common.AbstractCause("Knockback"), hitActor, nextToLastDir, 4),
				common.NewChangeResourceAction(common.AbstractCause("Knockback"), hitActor, stats.Health, -3),
			}
		}
	}
	return nil
}

