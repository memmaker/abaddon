package common

import (
	"fmt"
)

type OnDieHook struct {
	ItemHook
	effectDescription string
}

func (d *OnDieHook) ToString() string {
	return fmt.Sprintf("On Kill: %s", d.effectDescription)
}

func (d *OnDieHook) Name() string {
	return "OnDieHook"
}

func NewOnDieHook(effectDescription string, afterHook func(itemContext *ItemContext, action Action, engine ActionContext) []Action) *OnDieHook {
	return &OnDieHook{
		ItemHook: ItemHook{
			HookedAction: ActionNameDie,
			AfterHook:    afterHook,
		},
		effectDescription: effectDescription,
	}
}

type OnKillHook struct {
	BaseActionHook
}

func (d *OnKillHook) IsBefore() bool {
	return false
}

func (d *OnKillHook) IsAfter() bool {
	return true
}

func NewMainOnKillHook() *OnKillHook {
	return &OnKillHook{
		BaseActionHook: BaseActionHook{
			hookedAction: ActionNameDie,
		},
	}
}

func (d *OnKillHook) After(action Action, context ActionContext) []Action {
	dieAction := action.(*DieAction)
	killer := dieAction.Cause().GetActor()
	victim := dieAction.GetActor()

	state := context.GameState
	state.GetVisibility().RemoveActor(victim)
	if killer != nil && victim != nil {
		return d.onKill(context, killer, victim)
	}
	return nil
}

func (d *OnKillHook) onKill(context ActionContext, killer, victim *Actor) []Action {
	// should technically be its own action, if we wanted to hook this..
	state := context.GameState
	if killer == state.GetPlayer() {
		flags := state.GetFlags()
		flags.Increment(fmt.Sprintf("killed:%s", victim.GetInternalName()))

		currentMap := state.GetMap()
		livingEnemies := currentMap.GetFilteredActors(func(actor *Actor) bool {
			return actor.IsAlive() && !actor.IsVendor() && actor != killer
		})
		currentMapName := currentMap.GetName()
		if len(livingEnemies) == 0 {
			flags.SetTrue(fmt.Sprintf("all_clear:%s", currentMapName))
		}
	}
	return nil
}
