package common

import (
	"fmt"
	"rpg/defs"
)

type FallIntoChasmAction struct {
	BaseAction
	victim 		  *Actor
}

func NewFallIntoChasmAction(source Cause, victim *Actor) *FallIntoChasmAction {
	return &FallIntoChasmAction{
		BaseAction: BaseAction{
			cause: source,
			name:  ActionNameSpawnObject,
		},
		victim: victim,
	}
}
func (d *FallIntoChasmAction) Execute(context ActionContext) []Action {
	personPart := "You fall"
	currentMap := context.GameState.GetMap()
	currentMap.RemoveActor(d.victim)
	context.AudioPlayer.PlayCue("chasm_fall")
	victimIsPlayer := d.victim == context.GameState.GetPlayer()

	if !victimIsPlayer {
		context.GameState.GetTimeKeeper().Remove(d.victim)
		personPart = fmt.Sprintf("%s falls", d.victim.Name())
	}

	currentLevel := context.GameState.GetCurrentDungeonLevel()
	// DEBUG HACK
	if currentLevel == 0 {
		currentLevel = 2
	}
	var actionList []Action

	levelBelow := defs.LevelNameFromDepth(currentLevel - 1)

	if currentLevel < 2 {
		levelBelow = string(defs.LevelNameTownHell)
	} else if currentMap.GetName() == string(defs.LevelNameTown) {
		levelBelow = defs.LevelNameFromDepth(5)
	}

	if victimIsPlayer {
		transition := NewTransitionToRandom(d.cause, levelBelow)
		actionList = append(actionList, transition)
	} else {
		personPart = fmt.Sprintf("%s falls", d.victim.Name())
		context.GameState.PlaceOnMap(d.victim, levelBelow)
	}
	context.Msg(fmt.Sprintf("%s into a chasm!", personPart))
	return actionList
}

func (d *FallIntoChasmAction) GetAnimations() []Animation {
	return nil
}
