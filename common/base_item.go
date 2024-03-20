package common

import (
	"cmp"
	"fmt"
	"image/color"
	"rpg/defs"
	"rpg/geometry"
	"rpg/stats"
	"rpg/util"
	"slices"
)

type ItemRarity int

const (
	ItemRarityCommon ItemRarity = iota
	ItemRarityMagic
	ItemRarityUnique
)

type ArmorStats struct {
	BaseArmor  int
	Type       defs.ArmorType
	Plus       int
}

func (s ArmorStats) ToString(player *Actor) []util.LabelText {
	return []util.LabelText{
		util.LabelText{
			Text:      fmt.Sprintf("Armor Type: %s", s.Type.ToString()),
			TextColor: defs.OffWhite,
		},
		util.LabelText{
			Text:      fmt.Sprintf("Armor Class: %d", s.BaseArmor),
			TextColor: defs.OffWhite,
		},
		util.LabelText{
			Text:      fmt.Sprintf("Enchantment: +%d", s.Plus),
			TextColor: defs.OffWhite,
		},
	}
}

func (s ArmorStats) NameParts() string {
	// angband style [3, +3]
	return fmt.Sprintf("[%d, +%d]", s.BaseArmor, s.Plus)
}

func (s ArmorStats) GetTotalArmorClass() int {
	return s.BaseArmor + s.Plus
}

func (s ArmorStats) Clone() *ArmorStats {
	return &ArmorStats{
		BaseArmor:  s.BaseArmor,
		Type:       s.Type,
		Plus:       s.Plus,
	}
}

func (s ArmorStats) CanBeFurtherEnchanted() bool {
	return s.Plus < 3
}

type LightStats struct {
	LightColor        color.RGBA
	FlickerPattern    []float64
	FlickerFrameDelay float64
}

func (s LightStats) Clone() *LightStats {
	return &LightStats{
		LightColor:        s.LightColor,
		FlickerPattern:    s.FlickerPattern,
		FlickerFrameDelay: s.FlickerFrameDelay,
	}
}

type Item struct {
	name string

	pos  geometry.Point
	icon int32

	weaponComponent *WeaponStats
	armorComponent  *ArmorStats
	lightComponent  *LightStats
	throwComponent  *ThrowStats

	qualityLevel      int
	equippedModifiers []*stats.Modifier
	getPlayer         func() *Actor
	category          defs.ItemCategory
	rarity            ItemRarity

	droppedByPlayer bool
	isHidden        bool
	shouldBeDrawn   bool
	baseCost        int

	forcedTooltip      []util.LabelText
	itemHooks          []ItemActionHook
	targetedAction     *TargetedAction
	untargetedAction   *UntargetedAction
	actionCharges      int
	destroyAfterUse    bool
	nameOnlyInTooltip  bool
	internalName       string
	isQuestItem        bool
	dynamicDescription func(player *Actor) []util.LabelText
	isMetal            bool
}

func NewItem(getPlayerRef func() *Actor, name string, icon int32) *Item {
	return &Item{
		getPlayer:     getPlayerRef,
		name:          name,
		icon:          icon,
		shouldBeDrawn: true,
	}
}

func (i *Item) SetDrawnToScreen(shouldBeDrawn bool) {
	i.shouldBeDrawn = shouldBeDrawn
}

func (i *Item) IsDrawn() bool {
	return i.shouldBeDrawn
}

func (i *Item) SetIcon(icon int32) {
	i.icon = icon
}
func (i *Item) SetName(name string) {
	i.name = name
}

func (i *Item) Hide() {
	i.isHidden = true
}
func (i *Item) Show() {
	i.isHidden = false
}
func (i *Item) IsHidden() bool {
	return i.isHidden
}
func (i *Item) DroppedByPlayer() {
	i.droppedByPlayer = true
}

func (i *Item) IsDroppedByPlayer() bool {
	return i.droppedByPlayer
}
func (i *Item) GetRarity() ItemRarity {
	return i.rarity
}
func (i *Item) SetRarity(rarity ItemRarity) {
	i.rarity = rarity
}
func (i *Item) GetItemType() defs.ItemCategory {
	return i.category
}
func (i *Item) GetEquippedModifiers() []*stats.Modifier {
	return i.equippedModifiers
}
func (i *Item) Counter() int {
	return -1
}
func (i *Item) BuyPrice() int {
	itemPrice := i.baseCost
	if i.IsArmor() {
		pricePerPlus := 6
		plusValue := i.armorComponent.Plus * pricePerPlus
		itemPrice += plusValue
	}
	if i.IsWeapon() {
		pricePerPlus := 10
		plusValue := i.weaponComponent.Plus * pricePerPlus
		itemPrice += plusValue
	}
	chargeMultiplier := 7 // for infinite charges
	if !i.HasInfiniteCharges() {
		chargeMultiplier = i.GetCharges()
	}
	if i.HasTargetedAction() {
		itemPrice += i.targetedAction.GetMonetaryValue()*chargeMultiplier
	}
	if i.HasUntargetedAction() {
		itemPrice += i.untargetedAction.GetMonetaryValue()*chargeMultiplier
	}
	if i.HasModifiers() {
		for _, m := range i.equippedModifiers {
			itemPrice += m.GetMonetaryValue()
		}
	}
	return geometry.Abs(itemPrice)
}
func (i *Item) InventoryIcon() int32 {
	return i.icon
}

func (i *Item) GetLightStats() *LightStats {
	return i.lightComponent
}

func (i *Item) TintColor() color.RGBA {
	switch i.rarity {
	case ItemRarityCommon:
		return i.categoryColor()
/*	case ItemRarityMagic:
		return defs.MagicHighlightColor
 */
	case ItemRarityUnique:
		return defs.InvColorUnique
	}
	return defs.InvColorMisc
}

func (i *Item) categoryColor() color.RGBA {
	if i.IsRing() {
		return defs.InvColorRing
	}

	if i.IsAmulet() {
		return defs.InvColorAmulet
	}

	if i.IsPotion() {
		return defs.InvColorPotion
	}

	if i.IsScroll() {
		return defs.InvColorScroll
	}

	if i.IsWand() {
		return defs.InvColorWand
	}

	if i.IsArmor() {
		return defs.InvColorArmor
	}

	if i.IsRangedWeapon() {
		return defs.InvColorRangedWeapon
	}

	if i.IsMeleeWeapon() {
		return defs.InvColorMeleeWeapon
	}

	return defs.InvColorMisc
}

func (i *Item) Name() string {
	name := i.name
	return name
}
func (i *Item) String() string {
	return i.Name()
}

func (i *Item) IsWeapon() bool {
	return i.weaponComponent != nil
}

func (i *Item) IsArmor() bool {
	return i.armorComponent != nil
}

func (i *Item) IsMeleeWeapon() bool {
	return i.IsWeapon() && i.weaponComponent.IsMeleeWeapon()
}

func (i *Item) IsRangedWeapon() bool {
	return i.IsWeapon() && i.weaponComponent.IsRangedWeapon()
}

func (i *Item) SlotName() defs.SlotName {
	if i.IsRangedWeapon() {
		return defs.SlotNameMissileLauncher
	}
	if i.IsArmor() {
		return i.armorComponent.Type.GetInventorySlot()
	}
	if i.IsWeapon() {
		return i.weaponComponent.GetInventorySlot()
	}
	if i.IsThrowable() {
		return defs.SlotNameQuiver
	}
	if i.IsRing() {
		return defs.SlotNameRing
	}
	if i.IsAmulet() {
		return defs.SlotNameAmulet
	}
	return defs.SlotNameNotEquippable
}

func (i *Item) MapPosition() geometry.Point {
	return i.pos
}

func (i *Item) Icon(tick uint64) int32 {
	return i.icon
}

func (i *Item) SetMapPosition(pos geometry.Point) {
	i.pos = pos
}

func (i *Item) GetWeaponStats() *WeaponStats {
	return i.weaponComponent
}

func (i *Item) GetArmorStats() *ArmorStats {
	return i.armorComponent
}

func (i *Item) SetQualityLevel(level int) {
	i.qualityLevel = level
}

func (i *Item) SetWeaponComponent(component *WeaponStats) {
	i.weaponComponent = component
}

func (i *Item) SetEquippedModifiers(modifiers []*stats.Modifier) {
	i.equippedModifiers = modifiers
	i.sortModifiers()
}
func (i *Item) sortModifiers() {
	slices.SortStableFunc(i.equippedModifiers, func(i, j *stats.Modifier) int {
		return cmp.Compare(i.Attribute, j.Attribute)
	})
}

func (i *Item) RemoveEquippedModifier(modifier *stats.Modifier) {
	for index, m := range i.equippedModifiers {
		if m == modifier {
			i.equippedModifiers = append(i.equippedModifiers[:index], i.equippedModifiers[index+1:]...)
			return
		}
	}
}

func (i *Item) AddEquippedModifier(modifier *stats.Modifier) {
	i.equippedModifiers = append(i.equippedModifiers, modifier)
	i.sortModifiers()
}
func (i *Item) SetLightComponent(component *LightStats) {
	i.lightComponent = component
}
func (i *Item) SetArmorComponent(component *ArmorStats) {
	i.armorComponent = component
}

func (i *Item) GetThrowStats() *ThrowStats {
	return i.throwComponent
}

func (i *Item) GetWeaponStatsAsString() []util.LabelText {
	return i.GetWeaponStats().ToString(i.getPlayer())
}
func (i *Item) GetArmorStatsAsString() []util.LabelText {
	return i.GetArmorStats().ToString(i.getPlayer())
}

func (i *Item) SetCategory(category defs.ItemCategory) {
	i.category = category
}

func (i *Item) SetThrowComponent(throwComponent *ThrowStats) {
	i.throwComponent = throwComponent
}

func (i *Item) IsThrowable() bool {
	return i.throwComponent != nil
}

func (i *Item) SetBaseCost(cost int) {
	i.baseCost = cost
}

func (i *Item) Clone() *Item {
	clone := &Item{
		name:              i.name,
		internalName: 	i.internalName,
		pos:               i.pos,
		icon:              i.icon,
		weaponComponent:   nil,
		armorComponent:    nil,
		lightComponent:    nil,
		throwComponent:    nil,
		qualityLevel:      i.qualityLevel,
		equippedModifiers: i.equippedModifiers,
		getPlayer:         i.getPlayer,
		category:          i.category,
		rarity:            i.rarity,
		droppedByPlayer:   false,
		isHidden:          i.isHidden,
		shouldBeDrawn:     i.shouldBeDrawn,
		baseCost:          i.baseCost,
	}
	if i.weaponComponent != nil {
		clone.weaponComponent = i.weaponComponent.Clone()
	}
	if i.armorComponent != nil {
		clone.armorComponent = i.armorComponent.Clone()
	}
	if i.lightComponent != nil {
		clone.lightComponent = i.lightComponent.Clone()
	}
	if i.throwComponent != nil {
		clone.throwComponent = i.throwComponent.Clone()
	}

	var clonedUntargetedAction *UntargetedAction
	if i.untargetedAction != nil {
		clonedUntargetedAction = i.untargetedAction.Clone()
	}
	clone.forcedTooltip = i.forcedTooltip
	clone.itemHooks = i.itemHooks
	clone.targetedAction = i.targetedAction
	clone.untargetedAction = clonedUntargetedAction
	clone.actionCharges = i.actionCharges
	clone.destroyAfterUse = i.destroyAfterUse
	clone.nameOnlyInTooltip = i.nameOnlyInTooltip

	return clone
}

func (i *Item) SetCharges(charges int) {
	i.actionCharges = charges
}
func (i *Item) CanStackWith(other *Item) bool {
	if i.SlotName() != other.SlotName() {
		return false
	}

	if i.SlotName() != defs.SlotNameNotEquippable && i.SlotName() != defs.SlotNameQuiver {
		return false
	}

	if i.GetItemType() != other.GetItemType() {
		return false
	}

	if i.Name() != other.Name() {
		return false
	}

	if i.InventoryIcon() != other.InventoryIcon() {
		return false
	}

	if i.GetCharges() != other.GetCharges() {
		return false
	}
	if i.GetUntargetedAction() != nil && other.GetUntargetedAction() != nil {
		if i.GetUntargetedAction().Name() != other.GetUntargetedAction().Name() {
			return false
		}
	} else if !(i.GetUntargetedAction() == nil && other.GetUntargetedAction() == nil) {
		return false
	}

	if i.GetTargetedAction() != nil && other.GetTargetedAction() != nil {
		if i.GetTargetedAction().Name() != other.GetTargetedAction().Name() {
			return false
		}
	} else if !(i.GetTargetedAction() == nil && other.GetTargetedAction() == nil) {
		return false
	}
	return true
}

func (i *Item) SetUnlimitedCharges() {
	i.actionCharges = -1
}
func (i *Item) GetCharges() int {
	return i.actionCharges
}

func (i *Item) HasAction() bool {
	return i.targetedAction != nil || i.untargetedAction != nil
}

func (i *Item) GetTargetedAction() *TargetedAction {
	return i.targetedAction
}
func (i *Item) GetUntargetedAction() *UntargetedAction {
	return i.untargetedAction
}
func (i *Item) GetItemHooks() []ItemActionHook {
	return i.itemHooks
}
func (i *Item) GetTooltipLines(func(string) []util.ColoredTextPart) []util.LabelText {
	info := []util.LabelText{util.LabelTextFromStringWithColor(i.Name(), i.TintColor())}

	if i.nameOnlyInTooltip {
		return info
	}

	if len(i.forcedTooltip) > 0 {
		return append(info, i.forcedTooltip...)
	}

	if i.IsWeapon() {
		handedNess := "One-Handed"
		if i.SlotName() == defs.SlotNameTwoHandedWeapon {
			handedNess = "Two-Handed"
		}
		mrType := "melee"
		info = append(info, util.LabelTextFromString(fmt.Sprintf("%s %s weapon", handedNess, mrType)))
		info = append(info, util.LabelTextFromString(""))
		info = append(info, i.GetWeaponStatsAsString()...)
	}

	if i.IsArmor() {
		info = append(info, util.LabelTextFromString(""))
		info = append(info, i.GetArmorStatsAsString()...)
	}

	if len(i.GetEquippedModifiers()) > 0 {
		info = append(info, util.LabelTextFromString(""))
		for _, m := range i.GetEquippedModifiers() {
			info = append(info, util.LabelTextFromString(fmt.Sprintf("%s", m.ToString())))
		}
	}
	if len(i.GetItemHooks()) > 0 {
		info = append(info, util.LabelTextFromString(""), util.LabelTextFromString("Effects"))
		for _, h := range i.GetItemHooks() {
			info = append(info, util.LabelTextFromString(fmt.Sprintf("%s", h.ToString())))
		}
	}
	if i.HasAction() { // wands
		if i.HasInfiniteCharges() {
			//info = append(info, common.LabelTextFromString(fmt.Sprintf("Has infinite charges")))
			if i.targetedAction != nil {
				info = append(info, i.targetedAction.GetDescription()...)
			}
			if i.untargetedAction != nil {
				info = append(info, i.untargetedAction.GetDescription()...)
			}
		} else {
			if i.SelfDestructs() {
				info = append(info, util.LabelTextFromString(fmt.Sprintf("Is destroyed after use")))
			}

			if i.targetedAction != nil {
				if !i.SelfDestructs() || i.GetCharges() > 1 {
					info = append(info, util.LabelTextFromString(fmt.Sprintf("Has %d charges", i.GetCharges())))
				}
				info = append(info, i.targetedAction.GetDescription()...)
			}
			if i.untargetedAction != nil {
				if !i.SelfDestructs() || i.GetCharges() > 1 {
					info = append(info, util.LabelTextFromString(fmt.Sprintf("Has %d charges", i.GetCharges())))
				}
				info = append(info, i.untargetedAction.GetDescription()...)
			}
		}

	}

	if i.IsThrowable() {
		info = append(info, util.LabelTextFromString(""))
		info = append(info, util.LabelTextFromString(fmt.Sprintf("Can be thrown")))
		info = append(info, util.LabelTextFromString(""))
		info = append(info, i.GetThrowStats().ToLabelText()...)
	}

	if i.dynamicDescription != nil {
		info = append(info, util.LabelTextFromString(""))
		info = append(info,i.dynamicDescription(i.getPlayer())...)
	}

	return info
}
func (i *Item) GetAllEnchantments() []Enchantment {
	var result []Enchantment
	for _, m := range i.GetEquippedModifiers() {
		result = append(result, m)
	}
	for _, h := range i.GetItemHooks() {
		result = append(result, h)
	}
	return result
}

func (i *Item) RemoveEnchantment(enchantment Enchantment) {
	switch e := enchantment.(type) {
	case *stats.Modifier:
		i.RemoveEquippedModifier(e)
	case ItemActionHook:
		i.RemoveItemHook(e)
	}
}

func (i *Item) AddEnchantment(enchantment Enchantment) {
	switch e := enchantment.(type) {
	case *stats.Modifier:
		i.AddEquippedModifier(e)
	case ItemActionHook:
		i.AddItemhook(e)
	}
}

func (i *Item) RemoveItemHook(hookToRemove ItemActionHook) {
	for index, hook := range i.itemHooks {
		if hook == hookToRemove {
			i.itemHooks = append(i.itemHooks[:index], i.itemHooks[index+1:]...)
			return
		}
	}
}

func (i *Item) AddItemhook(hookToAdd ItemActionHook) {
	i.itemHooks = append(i.itemHooks, hookToAdd)
}

func (i *Item) IsEquippable() bool {
	return i.IsArmor() || i.IsWeapon() || i.IsAccessory()
}

func (i *Item) HasTargetedAction() bool {
	return i.targetedAction != nil
}

func (i *Item) HasUntargetedAction() bool {
	return i.untargetedAction != nil
}

func (i *Item) UseCharge() {
	if i.actionCharges > 0 {
		i.actionCharges--
	}
}

func (i *Item) HasInfiniteCharges() bool {
	return i.actionCharges == -1
}

func (i *Item) SelfDestructs() bool {
	return i.destroyAfterUse
}

func (i *Item) SetItemHooks(hooks []ItemActionHook) {
	i.itemHooks = hooks
}

func (i *Item) ShowNameOnlyInTooltip() {
	i.nameOnlyInTooltip = true
}
func (i *Item) SetTooltip(forcedTooltip []util.LabelText) {
	i.forcedTooltip = forcedTooltip
}

func (i *Item) HasChargesLeft() bool {
	return i.actionCharges != 0
}

func flatFunc(i int) func() int {
	return func() int {
		return i
	}
}

func (i *Item) BandStyleName() string {
	name := i.Name()
	if i.IsWeapon() {
		weaponValues := i.GetWeaponStats().NameParts()
		name = fmt.Sprintf("%s %s", name, weaponValues)
	}

	if i.IsArmor() {
		armorValues := i.GetArmorStats().NameParts()
		name = fmt.Sprintf("%s %s", name, armorValues)
	}

	return name
}

func (i *Item) PlusName() string {
	name := i.Name()
	if i.IsWeapon() && i.IsArmor() {
		weaponPlus := fmt.Sprintf("+%d", i.GetWeaponStats().Plus)
		armorPlus := fmt.Sprintf("+%d", i.GetArmorStats().Plus)
		return fmt.Sprintf("%s %s %s", name, weaponPlus, armorPlus)
	}
	if i.IsWeapon() {
		weaponPlus := fmt.Sprintf("+%d", i.GetWeaponStats().Plus)
		return fmt.Sprintf("%s %s", name, weaponPlus)
	}
	if i.IsArmor() {
		armorPlus := fmt.Sprintf("+%d", i.GetArmorStats().Plus)
		return fmt.Sprintf("%s %s", name, armorPlus)
	}
	return name
}

func (i *Item) SetDestroyAfterUse(value bool) {
	i.destroyAfterUse = value
}

func (i *Item) SetUntargetedAction(action *UntargetedAction) {
	i.untargetedAction = action
}

func (i *Item) SetTargetedAction(action *TargetedAction) {
	i.targetedAction = action
}

func (i *Item) IsRing() bool {
	return i.GetItemType() == defs.ItemCategoryRing
}

func (i *Item) IsAmulet() bool {
	return i.GetItemType() == defs.ItemCategoryAmulet
}

func (i *Item) IsAccessory() bool {
	return i.IsRing() || i.IsAmulet()
}
func (i *Item) IsPotion() bool {
	return i.GetItemType() == defs.ItemCategoryPotion
}

func (i *Item) IsScroll() bool {
	return i.GetItemType() == defs.ItemCategoryScroll
}

func (i *Item) IsWand() bool {
	return i.GetItemType() == defs.ItemCategoryWand
}

func (i *Item) IsTwoHandedWeapon() bool {
	return i.IsWeapon() && i.SlotName() == defs.SlotNameTwoHandedWeapon
}

func (i *Item) IsBow() bool {
	return i.IsRangedWeapon() && i.GetWeaponStats().Type == defs.WeaponTypeBow
}

func (i *Item) IsCrossbow() bool {
	return i.IsRangedWeapon() && i.GetWeaponStats().Type == defs.WeaponTypeCrossbow
}

func (i *Item) IsSlingshot() bool {
	return i.IsRangedWeapon() && i.GetWeaponStats().Type == defs.WeaponTypeSling
}

func (i *Item) GetInternalName() string {
	return i.internalName
}

func (i *Item) SetInternalName(name string) {
	i.internalName = name
}

func (i *Item) IsShield() bool {
	return i.IsArmor() && i.GetArmorStats().Type == defs.ArmorTypeShield
}


func (i *Item) HasModifiers() bool {
	return len(i.equippedModifiers) > 0
}

func (i *Item) IsQuestItem() bool {
	return i.isQuestItem
}

func (i *Item) SetDynamicDescription(getDescription func(player *Actor) []util.LabelText) {
	i.dynamicDescription = getDescription
}

func (i *Item) IsConductor() bool {
	return i.isMetal
}

type Enchantment interface {
	ToString() string
}
