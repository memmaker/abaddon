package common

import (
	"fmt"
)

type OnRangedHook struct {
	ItemHook
	hookDescription string
}

func (d OnRangedHook) ToString() string {
	return fmt.Sprintf("On Ranged: %s", d.hookDescription)
}

func (d OnRangedHook) Name() string {
	return "OnRangedHook"
}

func NewMainOnRangedHook() *OnRangedHook {
	return &OnRangedHook{
		ItemHook: ItemHook{
			HookedAction: ActionNameRangedAttack,
			AfterHook: func(context *ItemContext, action Action, actionContext ActionContext) []Action {
				return nil
			},
		},
	}
}

func NewOnRangedHook(hookDescription string, afterHook func(*ItemContext, Action, ActionContext) []Action) *OnRangedHook {
	return &OnRangedHook{
		ItemHook: ItemHook{
			HookedAction: ActionNameRangedAttack,
			AfterHook:    afterHook,
		},
		hookDescription: hookDescription,
	}
}
