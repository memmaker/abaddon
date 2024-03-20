package common

import "rpg/defs"

type Equipment struct {
	slots     map[defs.SlotName]*Item
	onChanged func()
}

func NewEquipment() *Equipment {
	return &Equipment{
		slots: make(map[defs.SlotName]*Item),
	}
}

func (e *Equipment) IsEquipped(item *Item) bool {
	slotName := item.SlotName()
	if slotName == defs.SlotNameTwoHandedWeapon || slotName == defs.SlotNameOneHandedWeapon{
		slotName = defs.SlotNameMainHand
	}

	if slotName == defs.SlotNameRing {
		if slotItem, exists := e.slots[defs.SlotNameRingLeft]; exists && slotItem == item {
			return true
		}
		if slotItem, exists := e.slots[defs.SlotNameRingRight]; exists && slotItem == item {
			return true
		}
	}
	if slotItem, exists := e.slots[slotName]; exists && slotItem == item {
		return true
	}
	return false
}

func (e *Equipment) Equip(item *Item) {
	defer e.changed()
	slotName := item.SlotName()
	if slotName == defs.SlotNameTwoHandedWeapon {
		e.slots[defs.SlotNameMainHand] = item
		delete(e.slots, defs.SlotNameOffHand)
		return
	}

	if slotName == defs.SlotNameOneHandedWeapon {
		e.slots[defs.SlotNameMainHand] = item
		return
	}

	if slotName == defs.SlotNameRing {
		if _, exists := e.slots[defs.SlotNameRingLeft]; !exists {
			e.slots[defs.SlotNameRingLeft] = item
			return
		}
		if _, exists := e.slots[defs.SlotNameRingRight]; !exists {
			e.slots[defs.SlotNameRingRight] = item
			return
		}
	}

	e.slots[slotName] = item
}

func (e *Equipment) UnEquip(item *Item) {
	defer e.changed()
	slotName := item.SlotName()
	if slotName == defs.SlotNameTwoHandedWeapon {
		e.unEquipBySlot(defs.SlotNameMainHand)
		e.unEquipBySlot(defs.SlotNameOffHand)
		return
	}
	if slotName == defs.SlotNameOneHandedWeapon {
		slotName = defs.SlotNameMainHand
	}
	if slotName == defs.SlotNameRing {
		if slotItem, exists := e.slots[defs.SlotNameRingLeft]; exists && slotItem == item {
			e.unEquipBySlot(defs.SlotNameRingLeft)
			return
		}
		if slotItem, exists := e.slots[defs.SlotNameRingRight]; exists && slotItem == item {
			e.unEquipBySlot(defs.SlotNameRingRight)
			return
		}
	}

	e.unEquipBySlot(slotName)
}

func (e *Equipment) unEquipBySlot(hand defs.SlotName) {
	delete(e.slots, hand)
}

func (e *Equipment) GetBySlot(hand defs.SlotName) *Item {
	return e.slots[hand]
}

func (e *Equipment) GetWeapon(attackType AttackType) *Item {
	isCorrectWeaponType := func(item *Item) bool {
		if attackType == MeleeAttack {
			return item.IsMeleeWeapon()
		}
		if attackType == RangedAttack {
			return item.IsRangedWeapon()
		}
		return false
	}
	if mainHand, exists := e.slots[defs.SlotNameMainHand]; exists && isCorrectWeaponType(mainHand) {
		return mainHand
	}
	if ranged, exists := e.slots[defs.SlotNameMissileLauncher]; exists && isCorrectWeaponType(ranged) {
		return ranged
	}
	if twoHanded, exists := e.slots[defs.SlotNameTwoHandedWeapon]; exists && isCorrectWeaponType(twoHanded) {
		return twoHanded
	}
	if offHand, exists := e.slots[defs.SlotNameOffHand]; exists && isCorrectWeaponType(offHand) {
		return offHand
	}

	return nil
}

func (e *Equipment) GetArmor() []*Item {
	var armor []*Item
	for _, item := range e.slots {
		if item.IsArmor() {
			armor = append(armor, item)
		}
	}
	return armor

}

func (e *Equipment) GetItemsToReplace(item *Item) []*Item {
	var items []*Item
	if item.SlotName() == defs.SlotNameOffHand {
		mainHandItem, hasItemInMainHand := e.slots[defs.SlotNameMainHand]
		if hasItemInMainHand && mainHandItem.IsTwoHandedWeapon() {
			items = appendIfNotNil(items, e.slots[defs.SlotNameMainHand])
			items = appendIfNotNil(items, e.slots[defs.SlotNameOffHand])
			return items
		}
		items = appendIfNotNil(items, e.slots[defs.SlotNameOffHand])
		return items
	}
	if item.SlotName() == defs.SlotNameTwoHandedWeapon {
		items = appendIfNotNil(items, e.slots[defs.SlotNameMainHand])
		items = appendIfNotNil(items, e.slots[defs.SlotNameOffHand])
		return items
	}
	if item.SlotName() == defs.SlotNameOneHandedWeapon {
		items = appendIfNotNil(items, e.slots[defs.SlotNameMainHand])
		return items
	}
	if item.SlotName() == defs.SlotNameRing {
		if e.isOneSlotAvailable(defs.SlotNameRingLeft, defs.SlotNameRingRight) {
			return items
		} else {
			items = appendIfNotNil(items, e.slots[defs.SlotNameRingLeft])
			return items
		}
	}
	items = appendIfNotNil(items, e.slots[item.SlotName()])
	return items
}

func (e *Equipment) isOneSlotAvailable(slotOne defs.SlotName, slotTwo defs.SlotName) bool {
	_, slotOneExists := e.slots[slotOne]
	_, slotTwoExists := e.slots[slotTwo]
	return !slotOneExists || !slotTwoExists
}

func appendIfNotNil(items []*Item, item *Item) []*Item {
	if item != nil {
		return append(items, item)
	}
	return items
}

func (e *Equipment) HasRangedWeaponEquipped() bool {
	slotMain := e.GetBySlot(defs.SlotNameMissileLauncher)
	return slotMain != nil && slotMain.IsRangedWeapon()
}

func (e *Equipment) HasMeleeWeaponEquipped() bool {
	slotMain := e.GetBySlot(defs.SlotNameMainHand)
	slotOff := e.GetBySlot(defs.SlotNameOffHand)
	return (slotMain != nil && slotMain.IsMeleeWeapon()) || (slotOff != nil && slotOff.IsMeleeWeapon())
}

func (e *Equipment) AllItems() []*Item {
	var items []*Item
	for _, item := range e.slots {
		items = append(items, item)
	}
	return items
}

func (e *Equipment) GetSlot(item *Item) defs.SlotName {
	for slot, slotItem := range e.slots {
		if slotItem == item {
			return slot
		}
	}
	return defs.SlotNameNotEquippable
}

func (e *Equipment) HasThrowableQuivered() bool {
	slot := e.GetBySlot(defs.SlotNameQuiver)
	return slot != nil && slot.IsThrowable()
}

func (e *Equipment) GetNextQuiveredThrowable() *Item {
	slot := e.GetBySlot(defs.SlotNameQuiver)
	if slot == nil {
		return nil
	}
	return slot
}

func (e *Equipment) HasWeaponEquipped() bool {
	slot := e.GetBySlot(defs.SlotNameMainHand)
	return slot != nil && slot.IsWeapon()
}

func (e *Equipment) HasArmorEquipped() bool {
	slot := e.GetBySlot(defs.SlotNameArmorTorso)
	return slot != nil && slot.IsArmor()
}

func (e *Equipment) SetOnChangeHandler(onChanged func()) {
	e.onChanged = onChanged
}

func (e *Equipment) changed() {
	if e.onChanged != nil {
		e.onChanged()
	}
}

func (e *Equipment) HasShieldEquipped() bool {
	slot := e.GetBySlot(defs.SlotNameOffHand)
	return slot != nil && slot.IsShield()
}

func (e *Equipment) GetShield() *Item {
	return e.GetBySlot(defs.SlotNameOffHand)
}