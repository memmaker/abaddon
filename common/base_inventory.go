package common

import (
	"cmp"
	"fmt"
	"image/color"
	"rpg/defs"
	"rpg/util"
	"slices"
)

type Inventory struct {
	items         []*Item
	maxItemStacks int
	letterToItem  map[rune]*Item
	itemToLetter  map[*Item]rune
	onChanged     func()
}

func NewInventory(maxItemStacks int) *Inventory {
	return &Inventory{
		items:         make([]*Item, 0),
		maxItemStacks: maxItemStacks,
		itemToLetter:  make(map[*Item]rune),
		letterToItem:  make(map[rune]*Item),
	}
}

func (i *Inventory) Items() []*Item {
	return i.items
}

type InventoryStack []*Item

func (i InventoryStack) Name() string {
	return appendStacks(i[0].Name(), len(i))
}

func appendStacks(name string, stackCount int) string {
	if stackCount > 1 {
		name = fmt.Sprintf("%s (x%d)", name, stackCount)
	}
	return name
}

func (i InventoryStack) PlusName() string {
	return appendStacks(i[0].PlusName(), len(i))
}

func (i InventoryStack) BandStyleName() string {
	return appendStacks(i[0].BandStyleName(), len(i))
}

func (i InventoryStack) String() string {
	return i.Name()
}

func (i InventoryStack) GetTooltipLines(textParser func(string) []util.ColoredTextPart) []util.LabelText {
	return i[0].GetTooltipLines(textParser)
}

func (i InventoryStack) InventoryIcon() int32 {
	return i[0].InventoryIcon()
}

func (i InventoryStack) TintColor() color.RGBA {
	return i[0].TintColor()
}

func (i InventoryStack) Counter() int {
	return len(i)
}

func (i *Inventory) StackedItems() []InventoryStack {
	return i.StackedItemsWithFilter(func(item *Item) bool { return true })
}
func (i *Inventory) StackedItemsWithFilter(filter func(*Item) bool) []InventoryStack {
	stacks := make([]InventoryStack, 0)
	for _, item := range i.items {
		if !filter(item) {
			continue
		}
		found := false
		for stackIndex, stack := range stacks {
			if stack[0].CanStackWith(item) {
				stack = append(stack, item)
				stacks[stackIndex] = stack
				found = true
				break
			}
		}
		if !found {
			stacks = append(stacks, []*Item{item})
		}
	}
	return stacks
}

func (i *Inventory) Remove(item *Item) {
	defer i.changed()
	for idx, invItem := range i.items {
		if invItem == item {
			i.items = append(i.items[:idx], i.items[idx+1:]...)
			i.updateLetterMapping()
			return
		}
	}
}

func (i *Inventory) Has(item *Item) bool {
	for _, invItem := range i.items {
		if invItem == item {
			return true
		}
	}
	return false
}

func (i *Inventory) Add(item *Item) {
	defer i.changed()
	i.items = append(i.items, item)
	i.updateLetterMapping()
}

func (i *Inventory) IsEmpty() bool {
	return len(i.items) == 0
}

func (i *Inventory) IsFull() bool {
	return len(i.StackedItems()) == i.maxItemStacks
}

func (i *Inventory) updateLetterMapping() {
	i.letterToItem = make(map[rune]*Item)
	sortedItems := i.StackedItems()
	SortInventory(sortedItems)
	for idx, item := range sortedItems {
		firstItem := item[0]
		assignedLetter := rune('a' + idx)
		if idx >= 26 {
			assignedLetter = rune('A' + idx - 26)
		}
		i.letterToItem[assignedLetter] = firstItem
		i.itemToLetter[firstItem] = assignedLetter
	}
}

func (i *Inventory) GetItemByLetter(letter rune) *Item {
	if i.letterToItem == nil {
		i.updateLetterMapping()
	}
	return i.letterToItem[letter]
}

func (i *Inventory) GetLetterForItemStack(item InventoryStack) rune {
	firstItem := item[0]
	return i.GetLetterForItem(firstItem)
}

func (i *Inventory) GetLetterForItem(firstItem *Item) rune {
	if i.itemToLetter == nil {
		i.updateLetterMapping()
	}
	return i.itemToLetter[firstItem]
}

func (i *Inventory) AsColorCodedStrings() []string {
	items := i.StackedItems()
	SortInventory(items)
	var asStrings []string
	for _, item := range items {
		itemColor := defs.ColorCodeFromColor(item.TintColor())
		itemLetter := i.GetLetterForItemStack(item)
		itemLine := fmt.Sprintf("[:white]%c) %s%s", itemLetter, itemColor, item.BandStyleName())
		if item.Counter() > 1 {
			itemLine = fmt.Sprintf("[:white]%c) %s%s (x%d)", itemLetter, itemColor, item.BandStyleName(), item.Counter())
		}
		asStrings = append(asStrings, itemLine)
	}
	return asStrings
}
func (i *Inventory) IsValidLetter(r rune) bool {
	return i.GetItemByLetter(r) != nil
}

func (i *Inventory) Clone() *Inventory {
	clonedInv := NewInventory(i.maxItemStacks)
	for _, item := range i.items {
		clonedInv.Add(item.Clone())
	}
	return clonedInv
}

func (i *Inventory) HasEnchantableArmor() bool {
	for _, item := range i.items {
		if item.IsArmor() && item.GetArmorStats().CanBeFurtherEnchanted() {
			return true
		}
	}
	return false
}

func (i *Inventory) HasEnchantableWeapon() bool {
	for _, item := range i.items {
		if item.IsWeapon() && item.GetWeaponStats().CanBeFurtherEnchanted() {
			return true
		}
	}
	return false
}

func (i *Inventory) SetOnChangeHandler(onChanged func()) {
	i.onChanged = onChanged
}

func (i *Inventory) changed() {
	if i.onChanged != nil {
		i.onChanged()
	}
}


func SortInventory(stacks []InventoryStack) {
	slices.SortStableFunc(stacks, func(i, j InventoryStack) int {
		itemI := i[0]
		itemTypeOne := itemI.GetItemType()
		itemJ := j[0]
		itemTypeTwo := itemJ.GetItemType()
		if itemTypeOne != itemTypeTwo {
			return cmp.Compare(itemTypeOne, itemTypeTwo)
		} else {
			switch itemTypeOne {
			case defs.ItemCategoryWeapon:
				weaponPartOne := itemI.GetWeaponStats()
				weaponPartTwo := itemJ.GetWeaponStats()
				if weaponPartTwo.Type != weaponPartOne.Type {
					return cmp.Compare(weaponPartOne.Type, weaponPartTwo.Type)
				}
				if weaponPartOne.Damage.Max() != weaponPartTwo.Damage.Max() {
					return cmp.Compare(weaponPartOne.Damage.Max(), weaponPartTwo.Damage.Max())
				}
				if weaponPartOne.Damage.Min() != weaponPartTwo.Damage.Min() {
					return cmp.Compare(weaponPartOne.Damage.Min(), weaponPartTwo.Damage.Min())
				}
				if weaponPartOne.Plus != weaponPartTwo.Plus {
					return cmp.Compare(weaponPartOne.Plus, weaponPartTwo.Plus)
				}
			case defs.ItemCategoryArmor:
				armorPartOne := itemI.GetArmorStats()
				armorPartTwo := itemJ.GetArmorStats()
				if armorPartTwo.Type != armorPartOne.Type {
					return cmp.Compare(armorPartOne.Type, armorPartTwo.Type)
				}
				if armorPartOne.BaseArmor != armorPartTwo.BaseArmor {
					return cmp.Compare(armorPartOne.BaseArmor, armorPartTwo.BaseArmor)
				}
				if armorPartOne.Plus != armorPartTwo.Plus {
					return cmp.Compare(armorPartOne.Plus, armorPartTwo.Plus)
				}
			case defs.ItemCategoryWand:
				fallthrough
			case defs.ItemCategoryPotion:
				fallthrough
			case defs.ItemCategoryScroll:
				fallthrough
			default:
				return cmp.Compare(itemI.Name(), itemJ.Name())
			}
		}
		return cmp.Compare(itemI.Name(), itemJ.Name())
	})
}

type TextualStack struct {
	InventoryStack
	letter   rune
	selected func()
}

func (t *TextualStack) Shortcut() rune {
	return t.letter
}

func (t *TextualStack) OnSelected() {
	t.selected()
}

func NewTextualStack(stack InventoryStack, letter rune, selected func(invItemKey rune)) *TextualStack {
	return &TextualStack{
		InventoryStack: stack,
		letter:         letter,
		selected: func() {
			selected(letter)
		},
	}
}

func (t *TextualStack) Name() string {
	return fmt.Sprintf("%c) %s", t.letter, t.InventoryStack.BandStyleName())
}

func (t *TextualStack) String() string {
	return t.Name()
}
