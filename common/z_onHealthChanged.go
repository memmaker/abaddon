package common

import (
	"rpg/stats"
)

type OnHealthChangedHook struct {
	BaseActionHook
}

func (d OnHealthChangedHook) IsBefore() bool {
	return false
}

func (d OnHealthChangedHook) IsAfter() bool {
	return true
}

func NewOnHealthChangedHook() *OnHealthChangedHook {
	return &OnHealthChangedHook{
		BaseActionHook: BaseActionHook{
			hookedAction: ActionNameChangeResource,
		},
	}
}

func (d OnHealthChangedHook) After(action Action, context ActionContext) []Action {
	changeResource := action.(*ChangeResourceAction)
	actorFromAction := changeResource.GetActor()

	if changeResource.GetResourceName() == stats.Health && changeResource.GetChangeAmount() < 0 {

		state := context.GameState
		if actorFromAction.IsVendor() {
			actorFromAction.SetMonsterFaction()
			actorFromAction.SetVendorType(VendorTypeNone)
			for _, otherActor := range state.GetMap().Actors() {
				if otherActor.IsVendor() {
					otherActor.SetMonsterFaction()
					otherActor.SetVendorType(VendorTypeNone)
				}
			}
		}
	}

	if changeResource.GetResourceName() == stats.Health && !actorFromAction.IsAlive() {
		return OneAction(NewDieAction(action.Cause(), actorFromAction))
	}

	return nil
}
