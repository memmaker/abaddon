package console

import "C"
import (
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"image/color"
	"log"
	"regexp"
	"rpg/common"
	"rpg/defs"
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/stats"
	"rpg/util"
	"slices"
	"strings"
)

type Console struct {
	context         common.UIContext
	terminationCode int
	lightStyle      tcell.Style
	darkStyle       tcell.Style
	screen          tcell.Screen
	mapWindow       *tview.Box
	messageLabel    *tview.TextView
	statusBar       *tview.TextView
	rootPages       *tview.Pages
	focus           tview.Primitive
}

func (c *Console) OpenCenteredTextWindowWithCallback(texts []util.LabelText, onClose func()) {

}

func (c *Console) FadeToBlack() {

}

func (c *Console) FadeFromBlack() {

}

func (c *Console) GetVisibleMap() geometry.Rect {
	return geometry.Rect{}
}

func (c *Console) IsPlayerForcingAggression() bool {
	return false
}

func (c *Console) HasOverlays() bool {
	return false
}

func (c *Console) ShowLogMessage(message string) {
	c.messageLabel.SetText(convertColorCodes(message))
}

func convertColorCodes(message string) string {
	regexPattern := `\[:([\w,]+)\]` // our color codes look like this [:red] [:blue] [:green] [:yellow] [:white] [:24,54,222]
	rp, _ := regexp.Compile(regexPattern)
	return rp.ReplaceAllStringFunc(message, func(match string) string {
		colorCode := match[2 : len(match)-1]
		if strings.Contains(colorCode, ",") {
			return fmt.Sprintf("[#%s]", colorCode)
		}
		return toTcellColorString(defs.ColorFromName(colorCode))
	})
}

func (c *Console) Cancel() {

}

func (c *Console) PopAll() {

}

func (c *Console) ShowGameOverMenu() {

}

func (c *Console) OpenMenu(options []util.MenuItem) {
	list := tview.NewList()
	list.ShowSecondaryText(false)
	list.SetBorder(true)
	longestOption := 0
	for index, i := range options {
		item := i
		list.AddItem(item.MainText, "", util.GetShortCutForIndex(index), func() {
			c.rootPages.HidePage("menu")
			c.focus = nil
			item.Action()
		})
		longestOption = max(longestOption, len(item.MainText))
	}
	w, h := c.screen.Size()
	height := min(h-4, len(options)+2)
	width := min(longestOption+6, w-4)
	modalContainer := makeModal(list, width, height)
	list.SetInputCapture(c.popOnEscape("menu"))
	c.rootPages.AddPage("menu", modalContainer, true, true)
	c.rootPages.ShowPage("menu")
	c.focus = list
}

func (c *Console) OpenTextWindow(msg []util.LabelText) {
	c.openTextWindow(toTcellColorCodes(msg))
}

func toTcellColorCodes(msg []util.LabelText) []string { // DOESN'T DO COLOR FOR NOW
	result := make([]string, len(msg))
	for i, labelText := range msg {
		if len(labelText.TextWithColorCodes) == 0 {
			result[i] = labelText.Text
			continue
		}
		result[i] = toTcellColorCodedText(labelText.TextWithColorCodes)
	}
	return result
}

func toTcellColorCodedText(codes []util.ColoredTextPart) string {
	var result strings.Builder
	for i, part := range codes {
		if i != 0 {
			result.WriteString(" ")
		}
		result.WriteString(toTcellColorString(part.Color))
		result.WriteString(part.Text)
	}
	return result.String()
}

func toTcellColorString(c color.RGBA) string {
	//[#8080ff]
	return fmt.Sprintf("[#%02x%02x%02x]", c.R, c.G, c.B)
}

func (c *Console) openTextWindow(msg []string) {
	textView := tview.NewTextView()
	textView.SetBorder(true)
	longestLine := MaxLenWithoutColorCodes(msg)
	textView.SetText(strings.Join(msg, "\n"))
	w, h := c.screen.Size()
	height := min(h-4, len(msg)+2)
	width := min(longestLine+2, w-4)
	modalContainer := makeModal(textView, width, height)
	textView.SetInputCapture(c.popOnEscape("text"))
	textView.SetDynamicColors(true)
	c.rootPages.AddPage("text", modalContainer, true, true)
	c.rootPages.ShowPage("text")
	c.focus = textView
}
func RemoveColorCodes(colorCodedText string) string {
	// color codes look like this [:red] [:blue] [:green] [:yellow] [:white] [:24,54,222]
	regexPattern := `\[([\w,]+)\]`
	rp, _ := regexp.Compile(regexPattern)
	return rp.ReplaceAllString(colorCodedText, "")
}

func MaxLenWithoutColorCodes(text []string) int {
	maxLength := 0
	for _, line := range text {
		asRunes := []rune(RemoveColorCodes(line))
		if len(asRunes) > maxLength {
			maxLength = len(asRunes)
		}
	}
	return maxLength
}
func (c *Console) CancelTargeting() {

}

func (c *Console) ShowFullStats() {
	player := c.context.GameState().GetPlayer()
	fullStats := player.GetFullStats()
	c.OpenTextWindow(fullStats)
	//label.SetText(fullStats)
}

func (c *Console) UpdateStatsLabel() {
	player := c.context.GameState().GetPlayer()
	pStats := player.GetStats()

	gold := pStats.GetAttribute(stats.Blood)
	goldStr := fmt.Sprintf("Blood: %-5d", gold)

	hp := pStats.GetResource(stats.Health).Value()
	hpValString := fmt.Sprintf("%d", hp)
	hpStr := fmt.Sprintf("Life: %-7s", hpValString)

	armor := player.GetArmorClass()
	armorStr := fmt.Sprintf("Armor: %-3d", armor)

	statusStr := fmt.Sprintf("%s %s %s", goldStr, hpStr, armorStr)

	c.statusBar.SetText(statusStr)
}
func (c *Console) OpenStash(inv *common.Inventory, town *common.Inventory, drop func(item *common.Item)) {

}

func (c *Console) OnMapChanged() {
	//c.screen.Clear()
	c.screen.Sync()
}

func (c *Console) OpenClassSelectionMenu() {

}

func (c *Console) Say(actor *common.Actor, linesOfText []string) {

}

func (c *Console) BeginTargeting(rules common.TargetRules, onHit *common.HitEffect) {

}

func (c *Console) OpenShopMenu(npc *common.Actor) {
	c.openShopMenuWithHandler(npc.GetInventory(), func(item *common.Item) {
		c.tryToBuy(item, npc)
	})
}
func (c *Console) tryToBuy(item *common.Item, seller *common.Actor) {
	player := c.context.GameState().GetPlayer()
	goldAmount := player.GetStats().GetAttribute(stats.Blood)
	msg := c.ShowLogMessage // TODO, this is wrong..
	cost := item.BuyPrice()
	cannotPayForItem := cost > goldAmount

	if cannotPayForItem {
		msg("You don't have enough blood")
		c.OpenShopMenu(seller)
		return
	}
	player.GetStats().DecrementBy(stats.Blood, cost)
	seller.GetStats().IncrementBy(stats.Blood, cost)

	alwaysItems := c.context.Definitions.VendorDefinitions[string(seller.GetVendorType())].Always
	seller.GetInventory().Remove(item)
	player.GetInventory().Add(item)
	itemName := item.GetInternalName()
	if slices.Contains(alwaysItems, itemName) {
		seller.GetInventory().Add(item.Clone())
	}

	msg(fmt.Sprintf("You bought [:darkText]%s[:white] for [:darkText]%d[:white] blood", item.Name(), cost))
	c.OpenShopMenu(seller)
}

func (c *Console) openShopMenuWithHandler(inventory *common.Inventory, handler func(item *common.Item)) {
	if inventory.IsEmpty() {
		return
	}
	player := c.context.GameState().GetPlayer()
	goldAmount := player.GetStats().GetAttribute(stats.Blood)

	inventoryStacks := inventory.StackedItems()
	common.SortInventory(inventoryStacks)

	table := tview.NewTable()
	table.SetWrapSelection(true, false)
	table.SetSelectedStyle(tcell.StyleDefault.Background(tcell.ColorGreen))
	table.SetBorders(false)
	table.SetBorder(true)
	table.SetSelectable(true, false)
	table.SetSelectedFunc(func(row, column int) {
		item := inventoryStacks[row][0]
		handler(item)
	})
	table.SetTitle(fmt.Sprintf("You have %d blood", goldAmount))
	longestName := 0
	longestPrice := 0
	for itemIndex, i := range inventoryStacks {

		item := i[0]
		cost := item.BuyPrice()
		cannotPayForItem := cost > goldAmount

		colorCodeForPrice := "darkText"
		if cannotPayForItem {
			colorCodeForPrice = "negative"
		}
		/*
			itemName := item.Name()
			if c.configuration.ItemNamesInShop == common.ItemNamingSchemeEnchantments {
				itemName = item.PlusName()
			} else if c.configuration.ItemNamesInShop == common.ItemNamingSchemeBandStyle {

		*/
		itemName := item.BandStyleName()
		//}
		if len(i) > 1 {
			itemName = fmt.Sprintf("%s (x%d)", itemName, len(i))
		}
		nameString := fmt.Sprintf("%s", itemName)
		//nameStringWithColor := fmt.Sprintf("%s", itemName)
		priceString := fmt.Sprintf("%d", cost)

		nameCell := &tview.TableCell{
			Text:          nameString,
			Color:         toTcellColor(item.TintColor()),
			Align:         tview.AlignLeft,
			NotSelectable: false,
		}
		priceCell := &tview.TableCell{
			Text:          priceString,
			Color:         toTcellColor(defs.ColorFromName(colorCodeForPrice)),
			Align:         tview.AlignRight,
			NotSelectable: false,
		}
		table.SetCell(itemIndex, 0, nameCell)
		table.SetCell(itemIndex, 1, priceCell)
		longestName = max(longestName, len(nameString))
		longestPrice = max(longestPrice, len(priceString))
	}

	longestItem := longestName + longestPrice + 1
	w, h := c.screen.Size()
	height := min(h-4, len(inventoryStacks)+2)
	width := min(longestItem+2, w-4)
	modalContainer := makeModal(table, width, height)
	table.SetInputCapture(c.popOnEscape("shop"))
	c.rootPages.AddPage("shop", modalContainer, true, true)
	c.rootPages.ShowPage("shop")
	c.focus = table
}
func (c *Console) BeginPlayerSellingTo(vendor *common.Actor) {

}

func (c *Console) OpenTransferEnchantmentMenu() {

}
func (c *Console) UpdateMiniMap() {

}

func (c *Console) OpenPlayerStashInTown() {

}

func (c *Console) OnPlayerTurnTaken() {
	c.UpdateStatsLabel()
}

func (c *Console) SetPlayerMessage(message func(message string)) {

}

func (c *Console) SetDebugActions(actions func() common.DebugActions) {
	c.context.DebugActions = actions
}

func (c *Console) HandleInput() (terminationCode int) {
	ev := c.screen.PollEvent()
	if ev == nil {
		return 1
	}

	switch typedEvent := ev.(type) {
	case *tcell.EventResize:
		c.onScreenSizeChanged()
	case *tcell.EventKey:
		if c.focus != nil {
			handler := c.focus.InputHandler()
			handler(typedEvent, func(p tview.Primitive) {
				c.focus = p
			})
			return c.terminationCode
		}
		typedRune := typedEvent.Rune()
		//mod, key, ch := typedEvent.Modifiers(), typedEvent.Key(), typedRune
		//println(fmt.Sprintf("EventKey Modifiers: %d Key: %d Rune: %d", mod, key, ch))
		actions := c.context.PlayerAction()
		if typedRune == 'w' {
			actions.ManualMovePlayer(geometry.Point{X: 0, Y: -1})
		} else if typedRune == 's' {
			actions.ManualMovePlayer(geometry.Point{X: 0, Y: 1})
		} else if typedRune == 'a' {
			actions.ManualMovePlayer(geometry.Point{X: -1, Y: 0})
		} else if typedRune == 'd' {
			actions.ManualMovePlayer(geometry.Point{X: 1, Y: 0})
		} else if typedRune == 'i' {
			c.OpenListInventory()
		} else if typedRune == 'l' {
			c.showLog()
		} else if typedRune == 'b' {
			c.OpenListInventoryForDropping()
		} else if typedRune == 'g' {
			actions.PickUpItem()
		} else if typedRune == 't' {
			actions.MountToggle()
		} else if typedRune == 'S' {
			c.OpenSkillsMenu()
		} else if typedRune == 'C' {
			c.ShowFullStats()
		} else if typedRune == 'Q' {
			c.screen.Clear()
			c.screen.Fini()
			c.terminationCode = 1
		} else if typedRune == 'o' {
			actions.StartAutoExplore()
		} else if typedEvent.Key() == tcell.KeyTab {
			actions.AutoAttack()
		} else if typedEvent.Key() == tcell.KeyF10 {
			c.context.DebugActions().OpenDebugMenu()
		}
	}
	return c.terminationCode
}

func (c *Console) putRune(screen tcell.Screen, x, y int, style tcell.Style, r rune) {
	screen.SetContent(x, y, r, nil, style)
}

func (c *Console) Update(dt float64) {
	// Poll event
	/*
		gameErr := c.context.Update()
		if gameErr != nil {
			c.terminationCode = 1
		}

	*/
	//c.context.Animator().CancelAllAnimations()
	c.context.Animator().Tick() // this is animations
	for _, actor := range c.context.GameState().GetMap().Actors() {
		actor.Tick()
	}
}

func (c *Console) OpenListInventory() {
	noFilter := func(element *common.Item) bool { return true }
	c.OpenListInventoryForSelection(noFilter, func(invItemKey rune) {
		inv := c.context.GameState().GetPlayer().GetInventory()
		c.onItemInInventoryClicked(inv.GetItemByLetter(invItemKey))
		c.closeInventory()
	})
}
func (c *Console) OpenListInventoryForDropping() {
	noFilter := func(element *common.Item) bool { return true }
	c.OpenListInventoryForSelection(noFilter, func(invItemKey rune) {
		inv := c.context.GameState().GetPlayer().GetInventory()
		c.context.PlayerAction().DropItem(inv.GetItemByLetter(invItemKey))
		c.closeInventory()
	})
}
func (c *Console) onItemInInventoryClicked(typedItem *common.Item) {
	state := c.context.GameState()
	if !state.GetPlayer().IsAlive() {
		return
	}
	act := c.context.PlayerAction()
	// shift state

	player := state.GetPlayer()

	wasEquipped := player.GetEquipment().IsEquipped(typedItem)
	/*
		if c.shiftIsBeingPressedThisTick { // drop
			act.DropItem(typedItem)
			if wasEquipped {
				c.ShowFullStats()
			}
			return
		}

	*/

	// equip
	if typedItem.IsEquippable() {
		if wasEquipped {
			act.UnequipItem(typedItem)
		} else {
			act.EquipItem(typedItem)
		}
		//c.ShowFullStats()
		return
	}

	// use
	contextActions := c.context.PlayerAction().GetContextActionsForItem(typedItem)
	if len(contextActions) == 0 {
		return // maybe open a description? or at least a menu with drop or sth?
	}

	if len(contextActions) == 1 {
		actionFunc := contextActions[0].Action
		actionFunc()
		// hmm, this will trigger even if unsuc
		return
	}
	c.OpenMenu(contextActions)
}
func (c *Console) OpenListInventoryForSelection(filter func(element *common.Item) bool, onSelected func(invItemKey rune)) {
	inventory := c.context.GameState().GetPlayer().GetInventory()
	items := common.SortStacksForTextUITyped(inventory.StackedItemsWithFilter(filter), inventory.GetLetterForItemStack, onSelected)
	//msg := c.
	if len(items) == 0 {
		c.ShowLogMessage("You are not carrying anything.")
		return
	}
	list := tview.NewList()
	list.SetBorder(true)
	longestItem := 0
	for _, item := range items {
		firstItem := item.InventoryStack[0]
		wornString := ""
		if firstItem.IsEquippable() && c.context.GameState().GetPlayer().GetEquipment().IsEquipped(firstItem) {
			wornString = fmt.Sprintf("(worn: %s)", firstItem.SlotName().ToString())
		}
		list.AddItem(item.BandStyleName(), wornString, item.Shortcut(), item.OnSelected)
		longestItem = max(longestItem, len(item.BandStyleName()))
	}
	w, h := c.screen.Size()
	height := min(h-4, len(items)*2+2)
	width := min(longestItem+6, w-4)
	modalContainer := makeModal(list, width, height)
	list.SetInputCapture(c.popOnEscape("inventory"))
	c.rootPages.AddPage("inventory", modalContainer, true, true)
	c.rootPages.ShowPage("inventory")
	c.focus = list
}
func (c *Console) popOnEscape(currentPage string) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEscape {
			c.rootPages.HidePage(currentPage)
			c.focus = nil
		}
		return event
	}
}
func (c *Console) Draw() {
	screen := c.screen

	c.rootPages.Draw(screen)

	screen.Show()
}

func NewConsole(gameContext common.UIContext) *Console {
	// Initialize screen
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := s.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	lightStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite)
	darkStyle := tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorGray)
	s.SetStyle(lightStyle)
	//s.EnableMouse()
	s.Clear()

	c := &Console{
		context:    gameContext,
		screen:     s,
		lightStyle: lightStyle,
		darkStyle:  darkStyle,
	}
	c.setupUI()
	return c
}
func (c *Console) setupUI() {
	c.mapWindow = tview.NewBox()
	c.mapWindow.SetDrawFunc(c.drawMap)

	c.messageLabel = tview.NewTextView()
	c.messageLabel.SetDynamicColors(true)
	c.messageLabel.SetText("Welcome, [green]Stranger!")

	c.statusBar = tview.NewTextView()
	//c.statusBar.SetText("Status Bar")

	grid := tview.NewGrid().
		SetRows(1, 0, 1).
		//SetColumns(30, 0, 30).
		//SetBorders(true).
		AddItem(c.messageLabel, 0, 0, 1, 1, 0, 0, false).
		AddItem(c.mapWindow, 1, 0, 1, 1, 5, 5, true).
		AddItem(c.statusBar, 2, 0, 1, 1, 0, 0, false)

	c.rootPages = tview.NewPages().
		AddPage("main", grid, true, true)

	width, height := c.screen.Size()

	c.rootPages.SetRect(0, 0, width, height)
	c.UpdateStatsLabel()
}

func (c *Console) drawMap(screen tcell.Screen, xOff int, yOff int, sWidth int, sHeight int) (int, int, int, int) {
	state := c.context.GameState()
	player := state.GetPlayer()
	currentMap := state.GetMap()
	scrollX := 0
	scrollY := 0
	if sWidth < currentMap.GetWidth() {
		scrollX = max(0, player.MapPosition().X-sWidth/2)
	}
	if sHeight < currentMap.GetHeight() {
		scrollY = max(0, player.MapPosition().Y-sHeight/2)
	}

	for y := 0; y < sHeight; y++ {
		for x := 0; x < sWidth; x++ {
			//drawInfos := c.context.StaticMapLookup(x, y, c.context.GetTicks())
			//lastInfo := drawInfos[len(drawInfos)-1]
			//lastInfo.

			mapPos := geometry.Point{X: scrollX + x, Y: scrollY + y}
			drawX := x + xOff
			drawY := y + yOff
			drawStyle := c.darkStyle

			if !currentMap.Contains(mapPos) || !currentMap.IsExplored(mapPos) {
				c.putRune(screen, drawX, drawY, drawStyle, ' ')
				continue
			}

			canSeeLocation := player.CanSee(mapPos)
			if canSeeLocation {
				drawStyle = c.lightStyle
			}
			drawRune := ' '
			if canSeeLocation && currentMap.IsActorAt(mapPos) {
				actorAt := currentMap.ActorAt(mapPos)
				if mapPos == player.MapPosition() {
					drawStyle = drawStyle.Foreground(toTcellColor(actorAt.GetStats().GetResource(stats.Health).AsColor()))
					drawRune = '@'
				} else {
					drawRune = actorAt.AsciiIcon(0)
					drawStyle = drawStyle.Foreground(toTcellColor(actorAt.GetColor()))
				}
				c.putRune(screen, drawX, drawY, drawStyle, drawRune)
			} else if canSeeLocation && currentMap.IsItemAt(mapPos) {
				//itemAt := currentMap.ItemAt(mapPos)
				c.putRune(screen, drawX, drawY, drawStyle, '!')
			} else if currentMap.IsObjectAt(mapPos) {
				objectAt := currentMap.ObjectAt(mapPos)
				if door, ok := objectAt.(*common.Door); ok {
					if door.IsClosed() {
						c.putRune(screen, drawX, drawY, drawStyle, '+')
					} else {
						c.putRune(screen, drawX, drawY, drawStyle, '´')
					}
				} else {
					c.putRune(screen, drawX, drawY, drawStyle, rune(objectAt.Name()[0]))
				}
			} else if currentMap.IsTileWithSpecialAt(mapPos, gridmap.SpecialTileStairsUp) {
				c.putRune(screen, drawX, drawY, drawStyle, '<')
			} else if currentMap.IsTileWithSpecialAt(mapPos, gridmap.SpecialTileStairsDown) {
				c.putRune(screen, drawX, drawY, drawStyle, '>')
			} else if currentMap.IsTileWithSpecialAt(mapPos, gridmap.SpecialTileWater) {
				c.putRune(screen, drawX, drawY, drawStyle.Foreground(toTcellColor(c.context.GameState().GetWaterColor())), '~')
			} else if currentMap.IsTileWithSpecialAt(mapPos, gridmap.SpecialTileMountain) {
				c.putRune(screen, drawX, drawY, drawStyle.Foreground(toTcellColor(defs.Brown)), '^')
			} else if currentMap.IsTileWithSpecialAt(mapPos, gridmap.SpecialTileLava) {
				c.putRune(screen, drawX, drawY, drawStyle.Foreground(toTcellColor(defs.Red)), '~')
			} else if currentMap.IsTileWithSpecialAt(mapPos, gridmap.SpecialTileForest) {
				c.putRune(screen, drawX, drawY, drawStyle.Foreground(toTcellColor(defs.Green)), ';')
			} else if currentMap.IsTileWalkable(mapPos) {
				if _, hasDecal := currentMap.GetDecal(mapPos); hasDecal {
					drawStyle = drawStyle.Foreground(toTcellColor(defs.Red))
				}
				c.putRune(screen, drawX, drawY, drawStyle, '·')
			} else {
				drawStyle = drawStyle.Reverse(true)
				c.putRune(screen, drawX, drawY, drawStyle, ' ')
			}
		}
	}
	return xOff, yOff, sWidth, sHeight
}

func toTcellColor(rgba color.RGBA) tcell.Color {
	return tcell.NewRGBColor(int32(rgba.R), int32(rgba.G), int32(rgba.B))
}

func (c *Console) closeInventory() {
	c.rootPages.HidePage("inventory")
	c.focus = nil
}

func (c *Console) onScreenSizeChanged() {
	c.screen.Sync()
	width, height := c.screen.Size()
	c.rootPages.SetRect(0, 0, width, height)
}

func (c *Console) showLog() {
	buffer := c.context.GetLog()
	if len(buffer) > 20 {
		buffer = buffer[len(buffer)-20:]
	}
	for i := range buffer {
		buffer[i] = convertColorCodes(buffer[i])
	}
	c.openTextWindow(buffer)
}
func (c *Console) GetActiveSkillsForUI(player *common.Actor) []fmt.Stringer {
	var activeSkills []fmt.Stringer
	for _, skill := range player.GetTargetedSkills() {
		activeSkills = append(activeSkills, skill)
	}
	for _, skill := range player.GetUntargetedSkills() {
		activeSkills = append(activeSkills, skill)
	}
	return activeSkills
}
func (c *Console) OpenSkillsMenu() {
	player := c.context.GameState().GetPlayer()
	allSkills := c.GetActiveSkillsForUI(player)
	var options []util.MenuItem
	for _, s := range allSkills {
		skill := s
		options = append(options, util.MenuItem{
			MainText: skill.String(),
			Action: func() {
				switch typedSkill := skill.(type) {
				case *common.TargetedAction:
					c.context.PlayerAction().BeginActionTargeting(typedSkill)
				case *common.UntargetedAction:
					c.context.PlayerAction().ActivateUntargetedAction(typedSkill, common.EffectSource("Player"))
				}
			},
		})

	}
	c.OpenMenu(options)
}

func makeModal(p tview.Primitive, width, height int) tview.Primitive {
	return tview.NewFlex().
		AddItem(nil, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(nil, 0, 1, false).
			AddItem(p, height, 1, true).
			AddItem(nil, 0, 1, false), width, 1, true).
		AddItem(nil, 0, 1, false)
}
