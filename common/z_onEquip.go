package common

type OnEquipHook struct {
	BaseActionHook
}

func (d OnEquipHook) IsBefore() bool {
	return true
}

func (d OnEquipHook) IsAfter() bool {
	return true
}

func (d OnEquipHook) Name() string {
	return "OnEquipHook"
}

func NewMainOnEquipHook() *OnEquipHook {
	return &OnEquipHook{
		BaseActionHook: BaseActionHook{
			hookedAction: ActionNameChangeEquipment,
		},
	}
}
func (d OnEquipHook) Before(action Action, engine ActionContext) []Action {
	changeEquipmentAction := action.(*ChangeEquipmentAction)

	actor := changeEquipmentAction.GetActor()
	item := changeEquipmentAction.GetItem()

	if changeEquipmentAction.GetMode() == EquipmentChangeModeUnequip {
		return nil
	}

	// we only care about equipping stuff here
	// we want to identify the equipment that needs to unequipped
	// and issue the unequip actions for them

	equipment := actor.GetEquipment()

	blockingItems := equipment.GetItemsToReplace(item)

	var actions []Action
	for _, blockingItem := range blockingItems {
		unequipAction := NewUnequipItemAction(action.Cause(), actor, blockingItem)
		actions = append(actions, unequipAction)
	}

	return actions
}

func (d OnEquipHook) After(action Action, context ActionContext) []Action {
	changeEquipmentAction := action.(*ChangeEquipmentAction)

	actor := changeEquipmentAction.GetActor()
	item := changeEquipmentAction.GetItem()

	actorStats := actor.GetStats()

	isEquipAction := changeEquipmentAction.GetMode() == EquipmentChangeModeEquip
	isUnequipAction := changeEquipmentAction.GetMode() == EquipmentChangeModeUnequip

	for _, modifier := range item.GetEquippedModifiers() {
		if isEquipAction {
			actorStats.AddModifier(modifier)
		} else if isUnequipAction {
			actorStats.RemoveModifier(modifier)
		}
	}

	for _, itemHook := range item.GetItemHooks() {
		itemHook.SetItemContext(&ItemContext{
			Wearer: actor,
			Item:   item,
		})

		broker := context.ActionBroker

		if isEquipAction {
			if itemHook.IsBefore() {
				broker.RegisterHook(itemHook)
			}
			if itemHook.IsAfter() {
				broker.RegisterHook(itemHook)
			}
		} else if isUnequipAction {
			if itemHook.IsBefore() {
				broker.RemoveBeforeHook(itemHook)
			}
			if itemHook.IsAfter() {
				broker.RemoveAfterHook(itemHook)
			}
		}
	}

	//context.UserInterface().ShowFullStats() // problem, this is a modal in the console UI..

	//actor := changeEquipmentAction.GetAttacker()
	/*if actor == state.GetPlayer() {
	      return OneAction(NewEndPlayerTurnAction(d, changeEquipmentAction.GetTimeNeededInAut()))
	  }
	*/
	return nil
}
