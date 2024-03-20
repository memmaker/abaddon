package common

import (
	"fmt"
)

type ChangeEquipmentAction struct {
	BaseAction
	actor   *Actor
	mode    EquipmentChangeMode
	item    *Item
	noSound bool
}

type EquipmentChangeMode bool

const (
	EquipmentChangeModeEquip   EquipmentChangeMode = true
	EquipmentChangeModeUnequip EquipmentChangeMode = false
)

func NewEquipItemAction(source Cause, actor *Actor, item *Item) *ChangeEquipmentAction {
	return &ChangeEquipmentAction{
		BaseAction: BaseAction{
			cause: source,
			name:  ActionNameChangeEquipment,
		},
		actor: actor,
		mode:  EquipmentChangeModeEquip,
		item:  item,
	}
}

func NewUnequipItemAction(source Cause, actor *Actor, item *Item) *ChangeEquipmentAction {
	return &ChangeEquipmentAction{
		BaseAction: BaseAction{
			cause: source,
			name:  ActionNameChangeEquipment,
		},
		actor: actor,
		mode:  EquipmentChangeModeUnequip,
		item:  item,
	}
}
func (d *ChangeEquipmentAction) Execute(context ActionContext) []Action {

	actor := d.actor
	equipment := actor.GetEquipment()
	item := d.item
	var message string

	if d.mode == EquipmentChangeModeEquip {
		equipment.Equip(item)

		name := actor.Name()
		verb := "equips"
		if context.GameState.GetPlayer() == actor {
			name = "You"
			verb = "equip"
		}
		message = fmt.Sprintf("%s %s %s", name, verb, item.Name())
		if !d.noSound  {context.AudioPlayer.PlayCue("item_equip")}
	} else if d.mode == EquipmentChangeModeUnequip {
		equipment.UnEquip(item)

		name := actor.Name()
		verb := "unequips"
		if context.GameState.GetPlayer() == actor {
			name = "You"
			verb = "unequip"
		}
		message = fmt.Sprintf("%s %s %s", name, verb, item.Name())
		if !d.noSound  {context.AudioPlayer.PlayCue("item_unequip")}
	}

	if context.GameState.GetPlayer() == actor {
		context.Msg(message)
	}

	return nil
}

func (d *ChangeEquipmentAction) GetAnimations() []Animation {
	return nil
}

func (d *ChangeEquipmentAction) GetMode() EquipmentChangeMode {
	return d.mode
}

func (d *ChangeEquipmentAction) GetItem() *Item {
	return d.item
}

func (d *ChangeEquipmentAction) GetActor() *Actor {
	return d.actor
}

func (d *ChangeEquipmentAction) NoSound() {
	d.noSound = true
}
