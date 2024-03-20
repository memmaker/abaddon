package common

import (
	"rpg/geometry"
	"rpg/util"
)

type SwapPositionsAction struct {
	BaseAction
	actorOne                  *Actor
	actorTwo         *Actor
	isPlayerInvolved bool
	animationDelay   float64
	isMovementVisibleToPlayer bool
	animationDisabled         bool
	newPosForOne              geometry.Point
	newPosForTwo              geometry.Point
}

func NewSwapPositionsAction(source Cause, actorOne, actorTwo *Actor, isPlayer bool) *SwapPositionsAction {
	return &SwapPositionsAction{
		BaseAction: BaseAction{
			cause: source,
			name:  ActionNameSwap,
		},
		actorOne:         actorOne,
		actorTwo:         actorTwo,
		isPlayerInvolved: isPlayer,
		animationDelay:   0.2,
	}
}

func (d *SwapPositionsAction) SetAnimationDelay(delay float64) {
	d.animationDelay = delay
}
func (d *SwapPositionsAction) Execute(context ActionContext) []Action {
	state := context.GameState
	currentMap := state.GetMap() // read state

	d.newPosForOne = d.actorTwo.MapPosition()
	d.newPosForTwo = d.actorOne.MapPosition()
	currentMap.SwapPositions(d.actorOne, d.actorTwo) // write state

	d.isMovementVisibleToPlayer = false
	if d.isPlayerInvolved {
		d.isMovementVisibleToPlayer = true
	} else {
		player := state.GetPlayer()
		d.isMovementVisibleToPlayer = player.CanSee(d.actorOne.MapPosition()) || player.CanSee(d.actorTwo.MapPosition())
	}
	if d.isNotAnimated() {
		d.actorOne.SetDrawPosition(d.newPosForOne.ToPointF())
		d.actorTwo.SetDrawPosition(d.newPosForTwo.ToPointF())
	}

	// player only
	if !d.isPlayerInvolved {
		return nil
	}
	var actions []Action
	player := state.GetPlayer()
	transitionAtLocation, hasTransition := currentMap.GetTransitionAt(player.MapPosition())
	if hasTransition {
		actions = append(actions, NewTransitionAction(d.Cause(), transitionAtLocation))
	}
	return actions
}

func (d *SwapPositionsAction) GetAnimations() []Animation {
	if d.isNotAnimated() {
		return nil
	}
	ticksNeeded := util.SecondsToTicks(d.animationDelay)
	animOne := &MovementAnimation{
		actor:       d.actorOne,
		path:        []geometry.Point{d.newPosForTwo, d.newPosForOne},
		ticksNeeded: ticksNeeded,
		isForPlayer: d.isPlayerInvolved,
	}
	animTwo := &MovementAnimation{
		actor:       d.actorTwo,
		path:        []geometry.Point{d.newPosForOne, d.newPosForTwo},
		ticksNeeded: ticksNeeded,
		isForPlayer: d.isPlayerInvolved,
	}
	return []Animation{animOne, animTwo}
}

func (d *SwapPositionsAction) isNotAnimated() bool {
	return !d.isMovementVisibleToPlayer || d.animationDisabled
}

func (d *SwapPositionsAction) GetActors() []*Actor {
	return []*Actor{d.actorOne, d.actorTwo}
}

func (d *SwapPositionsAction) DisableAnimation() {
	d.animationDisabled = true
}
