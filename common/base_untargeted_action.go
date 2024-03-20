package common

import (
	"fmt"
	"image/color"
	"rpg/defs"
	"rpg/util"
)

type UntargetedActionName string
type UntargetedAction struct {
	*BaseTimedBeing
	name                   string
	activationDelayInTurns int
	effect                 *UntargetedEffect
	user                   *Actor
	turnsUntilRemoval      int
	icon                   int32
	timeNeededInAut        int
	monetaryValue          int
	condition              func(user *Actor) bool
	effectSource           EffectSource
	internalName           string
}

func NewUntargetedAction(name string, icon int32) *UntargetedAction {
	return &UntargetedAction{
		name:            name,
		icon:            icon,
		timeNeededInAut: 10,
	}
}

func (a *UntargetedAction) GetDescription() []util.LabelText {
	var fullDescription []util.LabelText
	effectDescription := a.effect.GetDescription()
	fullDescription = append(fullDescription, util.LabelTextFromString(""))
	fullDescription = append(fullDescription, effectDescription...)

	if a.activationDelayInTurns > 0 {
		fullDescription = append(fullDescription, util.LabelTextFromString(""))
		fullDescription = append(fullDescription,
			util.LabelText{
				Text:          fmt.Sprintf("Delayed: Will activate after %d turns", a.turnsUntilRemoval),
				TextColor:     defs.DarkTextColor,
				UseColorCodes: false,
				//TextWithColorCodes: nil,
			},
		)
	}

	return fullDescription
}
func (a *UntargetedAction) SetMonetaryValue(value int) {
	a.monetaryValue = value
}

func (a *UntargetedAction) GetTooltipLines(func(string) []util.ColoredTextPart) []util.LabelText {
	return a.GetDescription()
}

func (a *UntargetedAction) InventoryIcon() int32 {
	return a.icon
}

func (a *UntargetedAction) TintColor() color.RGBA {
	return defs.White
}

func (a *UntargetedAction) IsAlive() bool {
	return a.activationDelayInTurns > 0
}

func (a *UntargetedAction) Turn() {
	a.activationDelayInTurns--
	if a.activationDelayInTurns == 0 {
		a.Invoke(a.user, a.effectSource)
	}
}

func (a *UntargetedAction) IsDelayed() bool {
	return a.turnsUntilRemoval > 0
}

func (a *UntargetedAction) Name() string {
	return string(a.name)
}
func (a *UntargetedAction) String() string {
	return string(a.name)
}

func (a *UntargetedAction) SetCurrentUserAndSource(user *Actor, source EffectSource) { // only needed for time delayed effects
	a.user = user
	a.effectSource = source
}

func (a *UntargetedAction) Invoke(actor *Actor, source EffectSource) {
	a.effect.Invoke(actor, source)
}

func (a *UntargetedAction) GetDelayInTurns() int {
	return a.activationDelayInTurns
}

func (a *UntargetedAction) Reset() {
	a.activationDelayInTurns = a.turnsUntilRemoval
}
func (a *UntargetedAction) Counter() int {
	if a.IsDelayed() {
		return a.GetDelayInTurns()
	}
	return -1
}

func (a *UntargetedAction) SetEffect(f *UntargetedEffect) {
	a.effect = f
}

func (a *UntargetedAction) SetDelay(delayToActivationInTurns int) {
	a.BaseTimedBeing = NewBaseTimedBeing()
	a.activationDelayInTurns = delayToActivationInTurns
	a.turnsUntilRemoval = delayToActivationInTurns
}

func (a *UntargetedAction) GetTimeNeededInAut() int {
	return a.timeNeededInAut
}

func (a *UntargetedAction) SetNoTimeNeeded() {
	a.timeNeededInAut = 0
}

func (a *UntargetedAction) GetMonetaryValue() int {
	return a.monetaryValue
}

func (a *UntargetedAction) Clone() *UntargetedAction {
	return &UntargetedAction{
		name:                   a.name,
		activationDelayInTurns: a.activationDelayInTurns,
		effect:                 a.effect,
		user:                   a.user,
		turnsUntilRemoval:      a.turnsUntilRemoval,
		icon:                   a.icon,
		timeNeededInAut:        a.timeNeededInAut,
		monetaryValue:          a.monetaryValue,
	}
}

func (a *UntargetedAction) CanActivate(user *Actor) bool {
	if a.condition == nil {
		return true
	}
	return a.condition(user)
}

func (a *UntargetedAction) SetCondition(condition func(user *Actor) bool) {
	a.condition = condition
}

func (a *UntargetedAction) SetInternalName(name string) {
	a.internalName = name
}

func (a *UntargetedAction) GetInternalName() string {
	return a.internalName
}
