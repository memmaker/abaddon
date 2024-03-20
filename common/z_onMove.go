package common

import (
	"fmt"

	"rpg/geometry"
)

type OnMoveHook struct {
	BaseActionHook
	autoPickup bool
}

func (d OnMoveHook) IsBefore() bool {
	return false
}

func (d OnMoveHook) IsAfter() bool {
	return true
}

func (d OnMoveHook) Name() string {
	return "OnMoveHook"
}

func NewMainOnMoveHook(autoPickup bool) *OnMoveHook {
	return &OnMoveHook{
		BaseActionHook: BaseActionHook{
			hookedAction: ActionNameMove,
		},
		autoPickup: autoPickup,
	}
}

type Mover interface {
	GetActor() *Actor
	GetNewPos() geometry.Point
}

func (d OnMoveHook) After(action Action, context ActionContext) []Action {
	var actions []Action
	state := context.GameState
	moveAction := action.(Mover)
	actor := moveAction.GetActor()
	newPos := moveAction.GetNewPos()
	currentMap := state.GetMap()

	followUp, _ := TriggerCellEffectsDuringMovementForActor(state, action, actor, newPos, AnimateNow())
	actions = append(actions, followUp...)

	if actor != state.GetPlayer() {
		actions = append(actions, NewUpdateFoVAction(action.Cause(), actor))
		return actions
	}

	// player only
	transitionAtLocation, hasTransition := currentMap.GetTransitionAt(newPos)
	if hasTransition {
		actions = append(actions, NewTransitionAction(action.Cause(), transitionAtLocation))
	}

	actions = append(actions, []Action{
		NewUpdateFoVAction(action.Cause(), actor),
	}...)
	d.logObjectsAtLocation(context, newPos)
	return actions
}

func (d OnMoveHook) logObjectsAtLocation(context ActionContext, pos geometry.Point) *Item {
	state := context.GameState
	if state.GetMap().IsItemAt(pos) {
		item := state.GetMap().ItemAt(pos)
		context.Msg(fmt.Sprintf("You see [:darkText]%s", item.Name()))
		return item
	}
	return nil
}
