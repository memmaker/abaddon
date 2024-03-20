package common

import "rpg/defs"

type UpdateFoVAction struct {
	BaseAction
	actor *Actor
}

func NewUpdateFoVAction(source Cause, actor *Actor) *UpdateFoVAction {
	return &UpdateFoVAction{
		BaseAction: BaseAction{
			cause: source,
			name:  "ActionNameUpdateFoV",
		},
		actor: actor,
	}
}
func (d *UpdateFoVAction) Execute(context ActionContext) []Action {
	state := context.GameState
	actor := d.actor
	currentMap := state.GetMap()

	if currentMap == nil {
		return nil
	}

	visionRange := 0
	if actor == state.GetPlayer() {
		visionRange = context.GameState.GetPlayerVisionRange()
	} else {
		visionRange = actor.VisionRange()
	}

	currentMap.UpdateFieldOfView(actor.GetFOV(), actor.MapPosition(), visionRange)

	updatePlayerKnowledge(context)

	if actor == state.GetPlayer() {
		state.updatePlayerDijkstra()
	}
	return nil
}

func updatePlayerKnowledge(context ActionContext) {
	state := context.GameState
	player := state.GetPlayer()
	currentMap := state.GetMap()
	vis := state.GetVisibility()

	currentlyVisible := make(map[*Actor]bool)
	cameIntoView := make(map[*Actor]bool)
	wentOutOfView := make(map[*Actor]bool)

	for _, p := range player.VisionMap() {
		currentMap.SetExplored(p)
		if state.GetMap().IsActorAt(p) {
			visibleActor := state.GetMap().ActorAt(p)
			if visibleActor == player {
				continue
			}
			currentlyVisible[visibleActor] = true
			if !vis.IsVisible(visibleActor) {
				cameIntoView[visibleActor] = true
			}
		}
	}

	if state.PlayerCanSeeNeighbors() {
		neighbors := currentMap.GetFilteredNeighbors(player.MapPosition(), currentMap.Contains)
		for _, p := range neighbors {
			currentMap.SetExplored(p)
		}
	}

	for previouslyVisible, _ := range vis.GetVisibleActors() {
		if _, exists := currentlyVisible[previouslyVisible]; !exists {
			wentOutOfView[previouslyVisible] = true
		}
	}
	vis.Set(currentlyVisible)
	for newActor, _ := range cameIntoView {
		if !newActor.IsHostileTo(player) {
			continue
		}
		//context.Msg(fmt.Sprintf("%s comes into view", newActor.Name()))
		if newActor.HasSpeech(defs.SpeechTriggerSight) {
			context.Speak(newActor, defs.SpeechTriggerSight)
		}
	}
}

func (d *UpdateFoVAction) GetAnimations() []Animation {
	return nil
}

