package stats

import (
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"image/color"
	"math"
	"strings"
)

func StatFromString(statName string) Name {
	return Name(strings.ToLower(statName))
}

type Name string

const (
	// baseline for the game
	BaseSpeed  Name = "speed"
	ArmorClass         Name = "armor_class"
	AttackSpeedFactor  Name = "attack_speed"
	MoveSpeedFactor    Name = "move_speed"

	MaxHealth Name = "max_health"

	Health Name = "health"

	AggroRange Name = "aggro_range"

	VisionRange Name = "vision_range"

	// 7drl specific
	Blood      Name = "blood"
)

type ModificationMode bool

const (
	ModifyAdditive       ModificationMode = false
	ModifyMultiplicative ModificationMode = true
)

type Modifier struct {
	Attribute      Name
	Description    string
	StackID        string
	Value          float64
	Mode           ModificationMode
	IsPercentBased bool
}

func (m Modifier) ToString() string {
	attributeName := strings.Replace(string(m.Attribute), "_", " ", -1)
	caser := cases.Title(language.English)
	niceName := caser.String(attributeName)
	if m.Mode == ModifyAdditive {
		sign := "+"
		if m.Value < 0 {
			sign = "-"
		}
		if m.IsPercentBased {
			intValue := int(math.Ceil(m.Value * 100))
			return fmt.Sprintf("%s%d%% %s", sign, intValue, niceName)
		}
		intValue := int(m.Value)
		return fmt.Sprintf("%s%d %s", sign, intValue, niceName)
	}
	// m.Mode == ModifyMultiplicative
	if m.Value < 1 {
		minusInPercent := int(math.Ceil((1.0 - m.Value) * 100))
		return fmt.Sprintf("-%d%% %s", minusInPercent, niceName)
	}
	plusPercent := int(math.Ceil((m.Value - 1) * 100))
	return fmt.Sprintf("+%d%% %s", plusPercent, niceName)
}

func (m Modifier) GetMonetaryValue() int {
	return int(m.Value * 65)
}

type AttributeHolder struct {
	attributes map[Name]float64
	resources  map[Name]*Resource
	modifiers  map[Name][]*Modifier
	onChanged  func()
}

func NewAttributeHolder() AttributeHolder {
	holder := AttributeHolder{
		attributes: map[Name]float64{
			Blood: 0,

			BaseSpeed:  10, // we now would need an attack speed and a movement speed modifier
			ArmorClass: 0,

			AttackSpeedFactor: 1,
			MoveSpeedFactor:   1,

			MaxHealth: 9,

			VisionRange:      20,

			AggroRange: 0, // No range limit by default, means aggro on sight
		},

		modifiers: make(map[Name][]*Modifier),
	}

	resources := map[Name]*Resource{
		Health: NewResource(-9, holder.getMaxForResource(MaxHealth), 9),
	}
	holder.resources = resources
	return holder
}
func (a *AttributeHolder) getMaxForResource(attr Name) func() int {
	return func() int {
		return a.GetAttribute(attr)
	}
}
func (a *AttributeHolder) GetAttributeBaseValue(name Name) float64 {
	return a.attributes[name]
}
func (a *AttributeHolder) GetAttribute(attr Name) int {
	return int(a.getModifiedValue(attr, a.GetAttributeBaseValue(attr)))
}

type ModifierImpact int

func (i ModifierImpact) ToColor() color.Color {
	switch i {
	case ModifierImpactPositive:
		return color.RGBA{0x00, 0xff, 0x00, 0xff}
	case ModifierImpactNegative:
		return color.RGBA{0xff, 0x00, 0x00, 0xff}
	default:
		return color.RGBA{0xff, 0xff, 0xff, 0xff}
	}
}

const (
	ModifierImpactNone ModifierImpact = iota
	ModifierImpactPositive
	ModifierImpactNegative
)

func (a *AttributeHolder) GetModifierType(attr Name) ModifierImpact {
	unmodified := a.GetAttributeBaseValue(attr)
	modified := a.getModifiedValue(attr, unmodified)
	if modified > unmodified {
		return ModifierImpactPositive
	}
	if modified < unmodified {
		return ModifierImpactNegative
	}
	return ModifierImpactNone
}
func (a *AttributeHolder) GetAttributeAsFloat(attr Name) float64 {
	return a.getModifiedValue(attr, a.GetAttributeBaseValue(attr))
}

func (a *AttributeHolder) GetResource(name Name) *Resource {
	return a.resources[name]
}

func (a *AttributeHolder) getModifiedValue(name Name, baseValue float64) float64 {
	modifierFactor := float64(1)
	modifierSummand := float64(0)

	allModifiers := a.modifiers[name]
	for i := len(allModifiers) - 1; i >= 0; i-- {
		mod := allModifiers[i]
		if mod.Mode == ModifyAdditive {
			modifierSummand += mod.Value
		} else {
			modifierFactor *= mod.Value
		}
	}

	return (baseValue * modifierFactor) + modifierSummand
}

func (a *AttributeHolder) GetAllModifiersForAttribute(name Name) []*Modifier {
	return a.modifiers[name]
}

func (a *AttributeHolder) AddModifier(modifier *Modifier) {
	defer a.statsChanged()
	attr := modifier.Attribute
	a.modifiers[attr] = append(a.modifiers[attr], modifier)
}
func (a *AttributeHolder) RemoveModifier(modifier *Modifier) {
	defer a.statsChanged()
	attr := modifier.Attribute
	for i, m := range a.modifiers[attr] {
		if m == modifier {
			a.modifiers[attr] = append(a.modifiers[attr][:i], a.modifiers[attr][i+1:]...)
			return
		}
	}
}
func (a *AttributeHolder) Increment(attributeName Name) {
	a.attributes[attributeName]++
	a.statsChanged()
}

func (a *AttributeHolder) Decrement(name Name) {
	a.attributes[name]--
	a.statsChanged()
}

func (a *AttributeHolder) SetAttribute(attributeName Name, value int) {
	if _, ok := a.attributes[attributeName]; !ok {
		//println("attribute not found", attributeName)
		return
	}
	a.attributes[attributeName] = float64(value)
	a.statsChanged()
}

func (a *AttributeHolder) SetAttributeAsFloat(attr Name, value float64) {
	if _, ok := a.attributes[attr]; !ok {
		println("attribute not found", attr)
		return
	}
	a.attributes[attr] = value
	a.statsChanged()
}

func (a *AttributeHolder) IncrementBy(attributeName Name, amount int) {
	a.attributes[attributeName] += float64(amount)
	a.statsChanged()
}

func (a *AttributeHolder) IncrementResourceBy(name Name, amount float64) {
	a.resources[name].Increase(amount)
	a.statsChanged()
}

func (a *AttributeHolder) DecrementBy(name Name, amount int) {
	a.attributes[name] -= float64(amount)
	a.statsChanged()
}

func (a *AttributeHolder) Clone() AttributeHolder {
	clone := NewAttributeHolder()
	for k, v := range a.attributes {
		clone.attributes[k] = v
	}
	for k, v := range a.resources {
		clone.resources[k] = v.Clone()
	}
	for k, v := range a.modifiers {
		clone.modifiers[k] = v
	}
	return clone
}

func (a *AttributeHolder) SetOnChangeHandler(onChanged func()) {
	a.onChanged = onChanged
}

func (a *AttributeHolder) statsChanged() {
	if a.onChanged != nil {
		a.onChanged()
	}
}

func (a *AttributeHolder) IsBuffed(statName Name) bool {
	if len(a.modifiers[statName]) == 0 {
		return false
	}
	totalModifier := 0.0
	for _, mod := range a.modifiers[statName] {
		totalModifier += mod.Value
	}
	return totalModifier > 0
}
