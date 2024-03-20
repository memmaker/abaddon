package common

import (
	"fmt"
)

type OnMeleeHook struct {
	ItemHook
	matchAttacker     *Actor
	effectDescription string
}

func (d *OnMeleeHook) ToString() string {
	return fmt.Sprintf("On Melee Hit: %s", d.effectDescription)
}

func (d *OnMeleeHook) SetMatchAttacker(actor *Actor) {
	d.matchAttacker = actor
}
func (d *OnMeleeHook) Name() string {
	return "OnMeleeHook"
}

func (d *OnMeleeHook) Before(action Action, context ActionContext) []Action {
	meleeAction := action.(*MeleeAttackAction)
	if d.matchAttacker != nil && d.matchAttacker != meleeAction.GetAttacker() {
		return nil
	}
	return d.ItemHook.Before(action, context)
}
func NewOnAfterMeleeHook(effectDescription string, afterHook func(*ItemContext, Action, ActionContext) []Action) *OnMeleeHook {
	return &OnMeleeHook{
		ItemHook: ItemHook{
			HookedAction: ActionNameMeleeAttack,
			AfterHook:    afterHook,
		},
		effectDescription: effectDescription,
	}
}
func NewOnBeforeMeleeHook(effectDescription string, beforeHook func(*ItemContext, Action, ActionContext) []Action) *OnMeleeHook {
	return &OnMeleeHook{
		ItemHook: ItemHook{
			HookedAction: ActionNameMeleeAttack,
			BeforeHook:   beforeHook,
		},
		effectDescription: effectDescription,
	}
}

func NewMainOnMeleeHook() *OnMeleeHook {
	return &OnMeleeHook{
		ItemHook: ItemHook{
			HookedAction: ActionNameMeleeAttack,
			AfterHook: func(itemContext *ItemContext, action Action, actionContext ActionContext) []Action {
				return nil
			},
		},
	}
}

type DamagingAction interface {
	GetTotalDamageDealt() int
}
