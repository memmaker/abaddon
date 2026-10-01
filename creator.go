//go:build !js


package main

import (
	"bufio"
	"fmt"
	"os"
	"rpg/defs"
	"strings"
)

type sitem struct {
	label  string
	action func()
}
type Creator struct {
	data        defs.DataContainer
	currentItem defs.ItemDefinition
}

func (c Creator) Run() {
	c.currentItem = defs.ItemDefinition{
		InventoryCategory: defs.ItemCategoryInvalid,
	}
	showMenu("What to create?", []sitem{
		{"Create Unique Item", c.createItem},
	})
}

func (c Creator) createItem() {
	var choices []sitem

	currentName := "(None)"
	if c.currentItem.Name != "" {
		currentName = c.currentItem.Name
	}
	nameLabel := fmt.Sprintf("Name: %s", currentName)
	choices = append(choices, sitem{
		label:  nameLabel,
		action: c.changeItemName,
	})

	var modifiersLabel string
	if len(c.currentItem.Modifiers) == 0 {
		modifiersLabel = "Modifiers: (None)"
	} else {
		modifiersLabel = fmt.Sprintf("Modifiers:\n %s", strings.Join(c.currentItem.Modifiers, "\n "))
	}
	choices = append(choices, sitem{
		label:  modifiersLabel,
		action: c.changeItemModifiers,
	})

	choices = append(choices, sitem{
		label:  "Category: " + c.currentItem.InventoryCategory.ToString(),
		action: c.changeItemCategory,
	})

	showMenu("Item Creation", choices)
}

func (c Creator) changeItemCategory() {
	itemCategories := c.data.GetItemCategories()
	var choices []sitem
	for _, cat := range itemCategories {
		category := cat
		choices = append(choices, sitem{
			label: category.ToString(),
			action: func() {
				c.currentItem.InventoryCategory = category
				c.createItem()
			},
		})
	}
	choices = append(choices, sitem{
		label:  "Back",
		action: c.Run,
	})
	showMenu("Item Category", choices)
}

func getString(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(prompt)
	answer, _ := reader.ReadString('\n')
	return answer
}

func (c Creator) changeItemName() {
	c.currentItem.Name = getString("Enter new name: ")
	c.createItem()
}

func (c Creator) changeItemModifiers() {
	var choices []sitem

	// add a modifier
	choices = append(choices, sitem{
		label:  "Add Modifier",
		action: c.addModifier,
	})
	// remove a modifier
	if len(c.currentItem.Modifiers) > 0 {
		choices = append(choices, sitem{
			label:  "Remove Modifier",
			action: c.removeModifier,
		})
	}

	choices = append(choices, sitem{
		label:  "Back",
		action: c.createItem,
	})

	showMenu("Modifiers", choices)
}

func (c Creator) addModifier() {
	availableModifiers := c.data.Modifiers
	var choices []sitem
	for k, _ := range availableModifiers {
		key := k
		choices = append(choices, sitem{
			label: key,
			action: func() {
				c.currentItem.Modifiers = append(c.currentItem.Modifiers, key)
				c.createItem()
			},
		})
	}

	choices = append(choices, sitem{
		label:  "Back",
		action: c.changeItemModifiers,
	})

	showMenu("Add Modifier", choices)
}

func (c Creator) removeModifier() {
	currentModifiers := c.currentItem.Modifiers
	var choices []sitem
	for _, m := range currentModifiers {
		modifier := m
		choices = append(choices, sitem{
			label: modifier,
			action: func() {
				for i, m := range c.currentItem.Modifiers {
					if m == modifier {
						c.currentItem.Modifiers = append(c.currentItem.Modifiers[:i], c.currentItem.Modifiers[i+1:]...)
						c.createItem()
						return
					}
				}
			},
		})
	}
}

func showMenu(prompt string, sitems []sitem) {
	validAnswer := false
	for validAnswer == false {
		println(prompt)
		for i, item := range sitems {
			println(i+1, item.label)
		}
		var choice int
		_, _ = fmt.Scanln(&choice)
		i := choice - 1
		if i < 0 || i >= len(sitems) {
			println("Invalid choice")
			continue
		}
		validAnswer = true
		sitems[i].action()
	}
}

func NewCreator(gameData defs.DataContainer) *Creator {
	return &Creator{
		data: gameData,
	}
}
