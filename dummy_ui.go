package main

import (
	"rpg/common"
	"rpg/geometry"
	"rpg/util"
)

type DummyUI struct {
}

func (d DummyUI) OpenCenteredTextWindowWithCallback(texts []util.LabelText, onClose func()) {

}

func (d DummyUI) FadeToBlack() {

}

func (d DummyUI) FadeFromBlack() {

}

func (d DummyUI) SetDebugActions(actions func()common.DebugActions) {

}

func (d DummyUI) GetVisibleMap() geometry.Rect {
	//TODO implement me
	return geometry.NewRect(0, 0, 0, 0)
}

func (d DummyUI) IsPlayerForcingAggression() bool {
	//TODO implement me
	return false
}

func (d DummyUI) HasOverlays() bool {
	//TODO implement me
	return true
}

func (d DummyUI) ShowLogMessage(message string) {
	//TODO implement me
	//println(message)
}

func (d DummyUI) Cancel() {
	//TODO implement me

}

func (d DummyUI) PopAll() {
	//TODO implement me

}

func (d DummyUI) ShowGameOverMenu() {
	//TODO implement me

}

func (d DummyUI) OpenMenu(options []util.MenuItem) {
	//TODO implement me

}

func (d DummyUI) OpenTextWindow(msg []util.LabelText) {
	//TODO implement me

}

func (d DummyUI) CancelTargeting() {
	//TODO implement me

}

func (d DummyUI) ShowFullStats() {
	//TODO implement me

}

func (d DummyUI) UpdateStatsLabel() {
	//TODO implement me

}

func (d DummyUI) OpenStash(inv *common.Inventory, town *common.Inventory, drop func(item *common.Item)) {
	//TODO implement me

}

func (d DummyUI) OnMapChanged() {
	//TODO implement me

}

func (d DummyUI) OpenClassSelectionMenu() {
	//TODO implement me

}

func (d DummyUI) Say(actor *common.Actor, linesOfText []string) {
	//TODO implement me

}

func (d DummyUI) BeginTargeting(rules common.TargetRules, onHit *common.HitEffect) {
	//TODO implement me

}

func (d DummyUI) OpenShopMenu(npc *common.Actor) {
	//TODO implement me

}

func (d DummyUI) BeginPlayerSellingTo(vendor *common.Actor) {
	//TODO implement me

}

func (d DummyUI) OpenTransferEnchantmentMenu() {
	//TODO implement me

}

func (d DummyUI) OpenListInventoryForSelection(filter func(element *common.Item) bool, selected func(itemKey rune)) {
	//TODO implement me

}

func (d DummyUI) UpdateMiniMap() {
	//TODO implement me

}

func (d DummyUI) OpenPlayerStashInTown() {
	//TODO implement me

}

func (d DummyUI) OnPlayerTurnTaken() {
	//TODO implement me

}

func (d DummyUI) SetPlayerMessage(message func(message string)) {
	//TODO implement me

}

func NewDummyUI() *DummyUI {
	return &DummyUI{}
}
