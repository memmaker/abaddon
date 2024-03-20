package common

import (
	"rpg/geometry"
	"rpg/gridmap"
)

type TransitionAction struct {
	BaseAction
	transitionTo gridmap.Transition

	targetLocation geometry.Point
	targetMapName    string
	isRandomLocation bool
}

func (t TransitionAction) Execute(context ActionContext) []Action {
	if !t.transitionTo.IsEmpty() {
		context.GameState.MapTransition(t.transitionTo)
	} else if t.targetMapName != "" {
		if t.isRandomLocation {
			context.GameState.MapTransitionToRandomLocation(t.targetMapName)
		} else {
			context.GameState.MapTransitionTo(t.targetMapName, t.targetLocation)
		}
	}

	return nil
}

func NewTransitionAction(source Cause, transition gridmap.Transition) *TransitionAction {
	return &TransitionAction{
		BaseAction: BaseAction{
			cause: source,
			name:  ActionNameTransition,
		},
		transitionTo: transition,
	}
}

func NewTransitionToAction(source Cause, targetMapName string, targetLocation geometry.Point) *TransitionAction {
	return &TransitionAction{
		BaseAction: BaseAction{
			cause: source,
			name:  ActionNameTransition,
		},
		targetMapName: targetMapName,
		targetLocation: targetLocation,
	}
}

func NewTransitionToRandom(source Cause, targetMapName string) *TransitionAction {
	return &TransitionAction{
		BaseAction: BaseAction{
			cause: source,
			name:  ActionNameTransition,
		},
		targetMapName:    targetMapName,
		isRandomLocation: true,
	}
}
