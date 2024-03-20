package common

import (
	"rpg/util"
)

type TalkToAction struct {
	BaseAction
	npc *Actor
}

func NewTalkToAction(source Cause, npc *Actor) *TalkToAction {
	return &TalkToAction{
		BaseAction: BaseAction{
			cause: source,
			name:  ActionNameTalkTo,
		},
		npc: npc,
	}
}
func (d *TalkToAction) Execute(context ActionContext) []Action {
	gui := context.UserInterface()
	npc := d.npc

	options := getVendorOptions(context, npc)
	if len(options) == 0 {
		return nil
	}

	gui.OpenMenu(options)
	return nil
}

func (d *TalkToAction) GetAnimations() []Animation {
	return nil
}

func getVendorOptions(context ActionContext, vendor *Actor) []util.MenuItem {
	vendorType := vendor.GetVendorType()
	var menuItems []util.MenuItem

	switch vendorType {
	case VendorTypeWands:
		menuItems = getWandVendorOptions(context, vendor)
	case VendorTypeWeapons:
		menuItems = getWeaponVendorOptions(context, vendor)
	case VendorTypeArmor:
		menuItems = getArmorVendorOptions(context, vendor)
	case VendorTypeGeneral:
		menuItems = getGeneralVendorOptions(context, vendor)
	}
	return menuItems
}
func getWandVendorOptions(context ActionContext, vendor *Actor) []util.MenuItem {
	gui := context.UserInterface()
	vendorOptions := []util.MenuItem{
		{
			MainText: "BUY",
			Action: func() {
				gui.OpenShopMenu(vendor)
			},
		},
		{
			MainText: "TRANSFER",
			Action:   gui.OpenTransferEnchantmentMenu,
		},
	}

	return vendorOptions
}

func getGeneralVendorOptions(context ActionContext, vendor *Actor) []util.MenuItem {
	gui := context.UserInterface()
	vendorOptions := []util.MenuItem{
		{
			MainText: "BUY",
			Action: func() {
				gui.OpenShopMenu(vendor)
			},
		},
	}

	return vendorOptions
}

func getWeaponVendorOptions(context ActionContext, vendor *Actor) []util.MenuItem {
	gui := context.UserInterface()
	vendorOptions := []util.MenuItem{
		{
			MainText: "BUY",
			Action: func() {
				gui.OpenShopMenu(vendor)
			},
		},
	}

	return vendorOptions
}

func getArmorVendorOptions(context ActionContext, vendor *Actor) []util.MenuItem {
	gui := context.UserInterface()
	vendorOptions := []util.MenuItem{
		{
			MainText: "BUY",
			Action: func() {
				gui.OpenShopMenu(vendor)
			},
		},
	}

	return vendorOptions
}