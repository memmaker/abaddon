package common

import (
	"fmt"
	"rpg/geometry"
	"rpg/gridmap"
)

type DieAction struct {
	BaseAction
	actor *Actor
}

func NewDieAction(source Cause, actor *Actor) *DieAction {
	return &DieAction{
		BaseAction: BaseAction{
			cause: source,
			name:  ActionNameDie,
		},
		actor: actor,
	}
}
func (d *DieAction) Execute(context ActionContext) []Action {
	state := context.GameState
	gameMap := state.GetMap() // read state
	msg := context.Msg
	actor := d.actor
	ui := context.UserInterface()
	var actions []Action
	if actor == state.GetPlayer() {
		msg(fmt.Sprintf("You die. CoD: %s", d.cause.ToString()))
		ui.ShowGameOverMenu()
	} else {
		gameMap.RemoveActor(actor)
		if actor.HasCorpse() {
			corpsePos := d.getCorpsePos(gameMap, actor.MapPosition())
			gameMap.AddDownedActor(actor, corpsePos) // write state
		}
		if actor.HasStatusEffect(StatusExplodeOnDeath) {
			explosion := NewExplosionAction(d.cause, actor.MapPosition(), 1, 1, 3)
			explosion.SetGory()
			actions = append(actions, explosion)
		}
		msg(fmt.Sprintf(actor.GetDeathStringFormat(), actor.Name()))
		//engine.UpdateMiniMap()
	}

	return actions
}
func (d *DieAction) getCorpsePos(gameMap *gridmap.GridMap[*Actor, *Item, Object], actorPos geometry.Point) geometry.Point {
	isFree := func(pos geometry.Point) bool {
		return gameMap.Contains(pos) && gameMap.IsWalkable(pos) && !gameMap.IsDownedActorAt(pos)
	}
	if isFree(actorPos) {
		return actorPos
	}
	freeCells := gameMap.GetFreeCellsForDistribution(actorPos, 1, isFree)
	if len(freeCells) == 0 {
		return actorPos
	}
	dropPos := freeCells[0]
	return dropPos
}
func (d *DieAction) GetAnimations() []Animation {
	return nil
}

func (d *DieAction) GetActor() *Actor {
	return d.actor
}
