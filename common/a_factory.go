package common

import (
	"fmt"
	"math/rand"
	"rpg/audio"
	"rpg/defs"
	"rpg/geometry"
	"rpg/stats"
	"rpg/util"
)

// Will turn definitions into actual game objects
type Factory struct {
	definitions  defs.DataContainer
	actionBroker ActionBroker
	randomSource *rand.Rand
	state        *State
	gui          func() UserInterface

	giveAndEquipItemToActor func(actor *Actor, item *Item)
	getTimeKeeper           func() *TimeKeeper
	logError                func(msg string)
	queueAfterTurn          func(func())
	audioPlayer             *audio.Player
}

type FactoryContext struct {
	Definitions  defs.DataContainer
	ActionBroker ActionBroker
	RandomSource *rand.Rand
	State        *State

	GetTimeKeeper func() *TimeKeeper

	GiveAndEquipItemToActor func(actor *Actor, item *Item)
	Msg                     func(msg string)
}

func NewFactory(context FactoryContext) *Factory {
	return &Factory{
		definitions:             context.Definitions,
		randomSource:            context.RandomSource,
		state:                   context.State,
		giveAndEquipItemToActor: context.GiveAndEquipItemToActor,
		actionBroker:            context.ActionBroker,
		getTimeKeeper:           context.GetTimeKeeper,
		logError:                context.Msg,
	}
}
func (g *Factory) NewRandomEnemy(dungeonLevel int) (*Actor, []defs.EnemyDefinition) {
	// get defs
	enemyDefinitions := g.definitions.EnemyDefinitions
	if dungeonLevel == 0 {
		dungeonLevel = 6
	}
	// filter for dungeon level
	var filteredEnemies []defs.EnemyDefinition
	for _, enemyDef := range enemyDefinitions {
		if !enemyDef.CanAppearAt(g.randomSource, dungeonLevel) {
			continue
		}
		filteredEnemies = append(filteredEnemies, enemyDef)
	}

	if len(filteredEnemies) == 0 {
		g.logError(fmt.Sprintf("no enemies found for dungeon level %d", dungeonLevel))
		return nil, nil
	}
	// pick random enemy definition
	enemyDef := filteredEnemies[g.randomSource.Intn(len(filteredEnemies))]

	// wants to spawn as a group?
	var groupDefs []defs.EnemyDefinition
	if len(enemyDef.GroupSpawn) > 0 {
		for _, otherDef := range enemyDef.GroupSpawn {
			otherDefName := otherDef
			spawnCount := 1
			if util.LooksLikeAFunction(otherDef) {
				oDefName, args := util.GetNameAndArgs(otherDef)
				otherDefName = oDefName
				spawnCount = stats.NewIntervalFromString(args.Get(0)).LinearRandom()
			}
			for i := 0; i < spawnCount; i++ {
				if otherEnemyDef, exists := enemyDefinitions[otherDefName]; exists {
					groupDefs = append(groupDefs, otherEnemyDef)
				}
			}
		}
	}
	// create enemy
	return g.NewEnemyFromDefinition(enemyDef, dungeonLevel), groupDefs
}

func (g *Factory) NewUniqueItemFromName(name string) *Item {
	// new way
	uniqueDefs := g.definitions.UniqueItemDefinitions

	if itemDef, ok := uniqueDefs[name]; ok {
		item := g.NewItemFromDefinition(itemDef, 1)
		item.SetRarity(ItemRarityUnique)

		// 7drl hack
		if name == "sword_of_vengeance" {
			item.SetDynamicDescription(func(player *Actor) []util.LabelText {
				damageBonus := GetVengeanceDamageBonus(player)
				playerStats := player.GetStats()
				vengeanceDamageDealt := playerStats.GetAttribute(stats.Blood)
				return []util.LabelText{
					{
						Text:               "A sword that thirsts for revenge.",
					},
					{
						Text:               fmt.Sprintf("It has drained %d liters of blood from your enemies.", vengeanceDamageDealt),
					},
					{
						Text:               fmt.Sprintf("It receives +%d damage dice, because of that.", damageBonus),
					},
				}
			})
			weaponPart := item.GetWeaponStats()
			specialWeaponPart := weaponPart.WithSpecialDamageDice(func(player *Actor) util.Dice {
				defaultDamage := weaponPart.Damage
				damageBonus := GetVengeanceDamageBonus(player)
				vengeanceDamage := defaultDamage.WithAddedDice(damageBonus)
				return vengeanceDamage
			})
			item.SetWeaponComponent(&specialWeaponPart)
		}
		return item
	}

	return nil
}

func (g *Factory) NewUniqueEnemyFromName(name string) *Actor {
	// new way
	uniqueDefs := g.definitions.UniqueEnemyDefinitions

	if enemyDef, ok := uniqueDefs[name]; ok {
		enemy := g.NewEnemyFromDefinition(enemyDef, 1)

		if name == "kirioth" {
			enemy.SetSpeechForTrigger(defs.SpeechTriggerHurtInTown, []string{
				"You cannot hurt me up here!",
				"The magic of this town and my people protects me",
				"Leave this place, or face the consequences!",
				"Demon, you are not welcome here!",
				"The Sun protects me!",
				"Never will you harm me in this place!",
				"I only fear the pits of hell!",
				"You will never drag me down to hell!",
			})
		}

		return enemy


	}

	return nil
}

func (g *Factory) NewEnemyFromDefinition(enemyDef defs.EnemyDefinition, dungeonLevel int) *Actor {
	actor := NewActor(enemyDef.Icon)
	actor.SetAsciiIcon(enemyDef.AsciiIcon)
	actor.SetUnicodeIcon(enemyDef.UnicodeIcon)
	actor.SetCP437Icon(enemyDef.CP437Icon)
	actor.SetColor(enemyDef.Color)
	//index := len(emptyMap.Actors())
	actor.SetName(enemyDef.Name)
	actor.SetInternalName(enemyDef.InternalName)
	actor.SetDeathStringFormats(enemyDef.DeathStringFormats)
	actor.SetMonsterFaction()

	if enemyDef.Tags.Has(defs.TagUnique) {
		actor.SetUnique()
	}
	if enemyDef.Tags.Has(defs.TagNoCorpse) {
		actor.SetNoCorpse()
	}
	if enemyDef.Tags.Has(defs.TagFlying) {
		actor.AddStatusEffect(StatusFlying, -1)
	}
	if enemyDef.Tags.Has(defs.TagWinCondition) {
		actor.AddStatusEffect(StatusWinCondition, -1)
	}

	for trigger, lines := range enemyDef.Speech {
		actor.SetSpeechForTrigger(trigger, lines)
	}
	actor.SetDamageDice(enemyDef.DamageDice)
	actor.SetBaseSpeed(enemyDef.BaseSpeed)

	actorStats := actor.GetStats()

	actorStats.SetAttribute(stats.MaxHealth, enemyDef.ResourceHealth.Roll())
	actorStats.GetResource(stats.Health).SetToMax()

	for statName, statValue := range enemyDef.Stats {
		actorStats.SetAttribute(stats.StatFromString(statName), statValue)
	}

	for _, itemName := range enemyDef.Equipment {
		item := g.NewItemFromDefinitionName(itemName, dungeonLevel)
		if item == nil {
			g.logError(fmt.Sprintf("could not find item definition %s", itemName))
			continue
		}
		g.giveAndEquipItemToActor(actor, item)
	}

	var targetedSkills []*TargetedAction
	for _, skillName := range enemyDef.TargetedSkills {
		skillAction := g.NewTargetedActionFromName(skillName)
		if skillAction != nil {
			targetedSkills = append(targetedSkills, skillAction)
		}
	}
	actor.SetTargetedSkills(targetedSkills)

	var untargetedSkills []*UntargetedAction
	for _, skillName := range enemyDef.UntargetedSkills {
		skillAction := g.NewUntargetedActionFromName(skillName)
		if skillAction != nil {
			untargetedSkills = append(untargetedSkills, skillAction)
		}
	}
	actor.SetUntargetedSkills(untargetedSkills)
	actor.SetAggroRange(enemyDef.AggroRange)

	return actor
}

func (g *Factory) GenerateLoot(dungeonLevel int) *Item {
	source := g.randomSource
	diceRoll := source.Intn(100)
	var addedItem *Item
	switch {
	case diceRoll >= 0 && diceRoll < 20:
		addedItem = g.NewRandomArmorPiece(dungeonLevel)
	case diceRoll >= 20 && diceRoll < 40:
		addedItem = g.NewRandomRing(dungeonLevel)
	case diceRoll >= 40 && diceRoll < 60:
		addedItem = g.NewRandomPotion(dungeonLevel)
	case diceRoll >= 60 && diceRoll < 100:
		addedItem = g.NewRandomScroll(dungeonLevel)
	}
	return addedItem
}

func (g *Factory) GenerateMagicItem(dungeonLevel int) *Item {
	source := g.randomSource
	diceRoll := source.Intn(100)
	var addedItem *Item

	switch {
	case diceRoll >= 0 && diceRoll < 40:
		addedItem = g.NewRandomWeapon(dungeonLevel)
	case diceRoll >= 40 && diceRoll < 80:
		addedItem = g.NewRandomArmorPiece(dungeonLevel)
	default:
		addedItem = g.NewRandomRing(dungeonLevel)
	}
	return addedItem
}
func (g *Factory) setRarityAndEnchantments(item *Item, itemLevel int) {
	if !item.IsWeapon() && !item.IsArmor() && !item.IsAccessory() {
		return
	}
	random := g.randomSource
	rarity := ItemRarityCommon
	rarityRoll := random.Intn(100) + 1 - itemLevel // 1-100

	if rarityRoll <= 20 {
		rarity = ItemRarityMagic
	}

	if rarity == ItemRarityCommon {
		return
	}

	enchantmentCount := 0
	if rarity == ItemRarityMagic {
		enchantmentCount = random.Intn(1) + 1
	}

	modifierCount, hookCount := randomSplit(random, enchantmentCount)

	if modifierCount > 0 {
		mods := defs.NewModifiersForItem(random, itemLevel, modifierCount)
		item.SetEquippedModifiers(mods)
	}

	if hookCount > 0 {
		itemHooks := NewHooksForItem(random, itemLevel, hookCount)
		item.SetItemHooks(itemHooks)
	}

	if item.IsArmor() {
		randomVal := min(10, random.Intn(10)+itemLevel)
		switch {
		case randomVal >= 5 && randomVal <= 7:
			item.GetArmorStats().Plus = 1
		case randomVal >= 8 && randomVal <= 9:
			item.GetArmorStats().Plus = 2
		case randomVal == 10:
			item.GetArmorStats().Plus = 3
		}
	}

	if item.IsWeapon() {
		randomVal := min(10, random.Intn(10)+itemLevel)
		switch {
		case randomVal >= 5 && randomVal <= 7:
			item.GetWeaponStats().Plus = 1
		case randomVal >= 8 && randomVal <= 9:
			item.GetWeaponStats().Plus = 2
		case randomVal == 10:
			item.GetWeaponStats().Plus = 3
		}
	}

}

func randomSplit(source *rand.Rand, value int) (int, int) {
	firstPart := source.Intn(value + 1)
	restPart := value - firstPart
	return firstPart, restPart
}

func (g *Factory) NewRandomWeapon(dungeonLevel int) *Item {
	// get defs
	itemDefinitions := g.definitions.ItemDefinitions

	// filter for dungeon level
	var filtered []defs.ItemDefinition
	weaponTypeMap := make(map[defs.WeaponType][]defs.ItemDefinition)
	for _, itemDef := range itemDefinitions {

		if !itemDef.IsWeapon() || !itemDef.CanAppearAt(g.randomSource, dungeonLevel) {
			continue
		}

		filtered = append(filtered, itemDef)
		weaponTypeMap[itemDef.Weapon.WeaponType] = append(weaponTypeMap[itemDef.Weapon.WeaponType], itemDef)
	}

	if len(filtered) == 0 {
		g.logError(fmt.Sprintf("no armor found for dungeon level %d", dungeonLevel))
		return nil
	}
	// group by type
	availableTypes := make([]defs.WeaponType, 0)
	for wType, _ := range weaponTypeMap {
		availableTypes = append(availableTypes, wType)
	}

	// pick random type
	weaponType := availableTypes[g.randomSource.Intn(len(availableTypes))]

	// pick random definition
	weaponDef := weaponTypeMap[weaponType][g.randomSource.Intn(len(weaponTypeMap[weaponType]))]

	// create item

	weapon := g.NewItemFromDefinition(weaponDef, dungeonLevel)
	return weapon
}

func (g *Factory) NewItemFromDefinition(itemDefinition defs.ItemDefinition, dungeonLevel int) *Item {
	state := g.state
	baseItem := NewItem(state.GetPlayer, itemDefinition.Name, itemDefinition.Icon)
	baseItem.SetBaseCost(itemDefinition.Cost)
	baseItem.SetCategory(itemDefinition.InventoryCategory)
	baseItem.SetQualityLevel(itemDefinition.DLvl)
	baseItem.SetInternalName(itemDefinition.InternalName)

	if itemDefinition.Armor.IsValid() {
		armorPart := &ArmorStats{
			BaseArmor:  itemDefinition.Armor.ArmorClass.LinearRandom(),
			Type:       itemDefinition.Armor.ArmorType,
		}
		baseItem.SetArmorComponent(armorPart)
	}

	if itemDefinition.Weapon.IsValid() {
		weaponPart := &WeaponStats{
			Damage:             itemDefinition.Weapon.Damage,
			Type:               itemDefinition.Weapon.WeaponType,
			AttackSpeedFactor:  itemDefinition.Weapon.Speed,
			MeleeRange:         1, // currently unused
			Plus:               0,
			MaxRange:           itemDefinition.Weapon.MaxRange,
			AccuracyAtMaxRange: itemDefinition.Weapon.AccuracyAtMaxRange,
			AccuracyAtMinRange: itemDefinition.Weapon.AccuracyAtMinRange,
			ArmorPenetration:   itemDefinition.Weapon.ArmorPenetration,
			ProjectileIcon:     itemDefinition.Weapon.ProjectileIcon,
			IsTwoHanded:        itemDefinition.Weapon.IsTwoHanded,
		}
		baseItem.SetWeaponComponent(weaponPart)
	}

	if itemDefinition.Throw.IsValid() {
		hitZone := HitZoneFromString(state.GetMap, itemDefinition.Throw.HitZone)
		if hitZone != nil {
			throwPart := &ThrowStats{
				Damage:        itemDefinition.Throw.Damage,
				HitZone:       hitZone,
				ChanceToBreak: itemDefinition.Throw.ChanceToBreak,
				OnHitEffect:   g.NewOnHitEffectFromString(itemDefinition.Throw.OnHitEffect),
			}
			baseItem.SetThrowComponent(throwPart)
		}
	}

	for _, modifierName := range itemDefinition.Modifiers {
		var mod *stats.Modifier
		if util.LooksLikeAFunction(modifierName) {
			modName, args := util.GetNameAndArgs(modifierName)
			dice := util.ParseDice(args.Get(0))
			mod = defs.NewModifierFromNameWithDiceBonus(modName, dice)
		} else {
			mod = defs.NewModifierFromNameWithDungeonLevel(modifierName, dungeonLevel)
		}
		if mod != nil {
			baseItem.AddEquippedModifier(mod)
		}
	}
	charges := itemDefinition.Charges
	destroyAfterUse := false

	if itemDefinition.InventoryCategory.HasAlwaysOneCharge() {
		charges = 1
		destroyAfterUse = true
	}

	var untargetedAction *UntargetedAction
	if itemDefinition.UntargetedAction != "" {
		untargetedAction = g.NewUntargetedActionFromName(itemDefinition.UntargetedAction)
	}

	var targetedAction *TargetedAction
	if itemDefinition.TargetedAction != "" {
		targetedAction = g.NewTargetedActionFromName(itemDefinition.TargetedAction)
	}

	baseItem.SetCharges(charges)
	baseItem.SetDestroyAfterUse(destroyAfterUse)
	baseItem.SetUntargetedAction(untargetedAction)
	baseItem.SetTargetedAction(targetedAction)

	g.setRarityAndEnchantments(baseItem, dungeonLevel)

	return baseItem
}

func GetFlickerPatternFromName(pattern string) []float64 {
	switch pattern {
	case "fire":
		return defs.GetFireFlickerPattern()
	case "smooth":
		return defs.GetSmoothFlickerPattern()
	}
	return nil
}

func (g *Factory) NewItemFromDefinitionName(itemString string, dungeonLevel int) *Item {
	if def, exists := g.definitions.ItemDefinitions[itemString]; exists {
		newItem := g.NewItemFromDefinition(def, dungeonLevel)
		if newItem == nil {
			g.logError(fmt.Sprintf("could not create item from definition %s", itemString))
		}
		return newItem
	}
	g.logError(fmt.Sprintf("Unknown item def: %s", itemString))
	return nil
}

func (g *Factory) NewTargetedActionFromName(name string) *TargetedAction {
	var args util.Arguments

	if util.LooksLikeAFunction(name) {
		name, args = util.GetNameAndArgs(name)
	}

	if def, exists := g.definitions.TargetedDefinitions[name]; exists {
		return g.NewTargetedActionFromDefinition(def, args)
	}
	g.logError(fmt.Sprintf("Unknown targeted action: %s", name))
	return nil
}

func (g *Factory) NewTargetedActionFromDefinition(definition defs.TargetedActionDefinition, args util.Arguments) *TargetedAction {
	hitEffect := g.NewOnHitEffectFromNameAndArgs(definition.OnHitEffect, args)
	if hitEffect == nil {
		g.logError(fmt.Sprintf("Unknown on hit effect: %s", definition.OnHitEffect))
		return nil
	}
	targetedAction := NewTargetedAction(
		definition.Name,
		HitZoneFromString(g.state.GetMap, definition.HitZone),
		hitEffect,
		NewTargetingRulesFromString(definition.Targeting),
	)

	targetedAction.SetInternalName(definition.InternalName)
	targetedAction.SetDungeonLevel(definition.DungeonLevel)
	targetedAction.SetIcon(definition.Icon)
	return targetedAction
}

func (g *Factory) GetAllHitEffects() map[string]func(args util.Arguments) *HitEffect {
	return map[string]func(args util.Arguments) *HitEffect{
		"confuse":             g.confuseTarget,
		"berserk":             g.berserkTarget,
		"poison":              g.poisonTarget,
		"immobilize":          g.immobilizeTarget,
		"fireball":            g.throwFireball,
		"project_fire":        g.projectFire,
		"summon_monster":      g.summonRandomMonster,
		"jumpstart":           g.jumpStart,
		"fire_zone":           g.fireZone,
		"steal_item":          g.stealItem,
		"steal_item_teleport": g.stealItemAndTeleport,
		"mind_push":           g.mindPush,
		"bridge_chasm":        g.bridge,
		"blink":               g.blink,
		"earthshaker":         g.earthShaker,
		"break_floor":         g.breakFloor,
	}
}

func (g *Factory) NewOnHitEffectFromString(effectName string) *HitEffect {
	if effectName == "" {
		return nil
	}
	var args util.Arguments
	if util.LooksLikeAFunction(effectName) {
		effectName, args = util.GetNameAndArgs(effectName)
	}
	return g.NewOnHitEffectFromNameAndArgs(effectName, args)
}

func (g *Factory) NewOnHitEffectFromNameAndArgs(effectName string, args util.Arguments) *HitEffect {
	allHitEffects := g.GetAllHitEffects()
	if _, ok := allHitEffects[effectName]; ok {
		return allHitEffects[effectName](args)
	}
	if effectName != "" {
		g.logError(fmt.Sprintf("Unknown on hit effect: %s", effectName))
	}
	return nil
}

func (g *Factory) NewRandomArmorPiece(dungeonLevel int) *Item {
	// get defs
	itemDefinitions := g.definitions.ItemDefinitions

	// filter for dungeon level
	var filteredArmorDefinitions []defs.ItemDefinition
	armorTypeMap := make(map[defs.ArmorType][]defs.ItemDefinition)
	for _, itemDef := range itemDefinitions {

		if !itemDef.IsArmor() || !itemDef.CanAppearAt(g.randomSource, dungeonLevel) {
			continue
		}

		filteredArmorDefinitions = append(filteredArmorDefinitions, itemDef)
		armorTypeMap[itemDef.Armor.ArmorType] = append(armorTypeMap[itemDef.Armor.ArmorType], itemDef)
	}

	if len(filteredArmorDefinitions) == 0 {
		g.logError(fmt.Sprintf("no armor found for dungeon level %d", dungeonLevel))
		return nil
	}
	// group by armor type
	availableArmorTypes := make([]defs.ArmorType, 0)
	for armorType, _ := range armorTypeMap {
		availableArmorTypes = append(availableArmorTypes, armorType)
	}

	// pick random armor type
	armorType := availableArmorTypes[g.randomSource.Intn(len(availableArmorTypes))]

	// pick random armor definition
	armorDefinition := armorTypeMap[armorType][g.randomSource.Intn(len(armorTypeMap[armorType]))]

	// create item
	newArmor := g.NewItemFromDefinition(armorDefinition, dungeonLevel)
	return newArmor
}

func (g *Factory) NewRandomRing(dungeonLevel int) *Item {
	// get defs
	itemDefinitions := g.definitions.ItemDefinitions

	// filter for dungeon level
	var filtered []defs.ItemDefinition
	for _, itemDef := range itemDefinitions {

		if (!itemDef.IsRing()) || !itemDef.CanAppearAt(g.randomSource, dungeonLevel) {
			continue
		}

		filtered = append(filtered, itemDef)
	}

	if len(filtered) == 0 {
		g.logError(fmt.Sprintf("no accessory found for dungeon level %d", dungeonLevel))
		return nil
	}
	// pick random type
	accessoryDef := filtered[g.randomSource.Intn(len(filtered))]
	newAccessory := g.NewItemFromDefinition(accessoryDef, dungeonLevel)
	return newAccessory
}

func (g *Factory) NewRandomWand(dungeonLevel int) *Item {
	source := g.randomSource
	var newWand *Item
	roll := source.Intn(100)
	if roll < 50 {
		wandType := []func(targetedAction *TargetedAction) *Item{
			g.NewFiniteTargetedWand,
			g.NewEtherealTargetedWand,
			g.NewInfiniteTargetedWand,
		}
		randomTargetedAction := g.GetRandomTargetedAction(source, dungeonLevel)
		wandTypeRoll := source.Intn(100)
		wandTypeIndex := 2
		if wandTypeRoll < 45 {
			wandTypeIndex = 0
		} else if wandTypeRoll < 90 {
			wandTypeIndex = 1
		}
		randomTargetableFunc := wandType[wandTypeIndex]
		newWand = randomTargetableFunc(randomTargetedAction)
		newWand.SetInternalName("wand_of_" + randomTargetedAction.GetInternalName())
	} else {
		wandType := []func(untargetedAction *UntargetedAction) *Item{
			g.NewFiniteUntargetedWand,
			g.NewEtherealUntargetedWand,
			g.NewInfiniteUntargetedWand,
		}
		randomUntargetedAction := g.GetRandomUntargetedAction(dungeonLevel, defs.ForWand)
		wandTypeRoll := source.Intn(100)
		wandTypeIndex := 2
		if wandTypeRoll < 45 {
			wandTypeIndex = 0
		} else if wandTypeRoll < 90 {
			wandTypeIndex = 1
		}
		randomUntargetableFunc := wandType[wandTypeIndex]
		newWand = randomUntargetableFunc(randomUntargetedAction)
		newWand.SetInternalName("wand_of_" + randomUntargetedAction.GetInternalName())
	}
	/* scrapped for now, reasons: need an extra UI step (menu), name too long for UI display
	   else {
	       randomTargetedAction := g.GetRandomTargetedAction(source, dungeonLevel)
	       randomUntargetedAction := g.GetRandomUntargetedAction(source, dungeonLevel)
	       newConsumable = g.NewTwinWand(randomTargetedAction, randomUntargetedAction)
	   }
	*/

	return newWand
}

func (g *Factory) NewInfiniteTargetedWand(targetedAction *TargetedAction) *Item {
	return g.NewTargetedWand(targetedAction, -1, false)
}

func (g *Factory) NewFiniteTargetedWand(targetedAction *TargetedAction) *Item {
	return g.NewTargetedWand(targetedAction, 3, false)
}
func (g *Factory) NewEtherealTargetedWand(targetedAction *TargetedAction) *Item {
	return g.NewTargetedWand(targetedAction, 5, true)
}
func (g *Factory) NewInfiniteUntargetedWand(untargetedAction *UntargetedAction) *Item {
	wand := g.NewUntargetedWand(untargetedAction, -1, false)
	wand.SetInternalName("wand_of_" + untargetedAction.GetInternalName())
	return wand
}

func (g *Factory) NewFiniteUntargetedWand(untargetedAction *UntargetedAction) *Item {
	return g.NewUntargetedWand(untargetedAction, 3, false)
}

func (g *Factory) NewEtherealUntargetedWand(untargetedAction *UntargetedAction) *Item {
	return g.NewUntargetedWand(untargetedAction, 5, true)
}

func (g *Factory) NewTargetedWand(targetedAction *TargetedAction, charges int, destroyAfterUse bool) *Item {
	itemName := fmt.Sprintf("a wand of %s", targetedAction.Name())
	baseItem := NewItem(g.state.GetPlayer, itemName, defs.IconItemWand)
	baseItem.SetCategory(defs.ItemCategoryWand)
	baseItem.SetQualityLevel(1)
	baseItem.SetCharges(charges)
	baseItem.SetDestroyAfterUse(destroyAfterUse)
	baseItem.SetTargetedAction(targetedAction)

	return baseItem
}
func (g *Factory) NewUntargetedWand(untargetedAction *UntargetedAction, charges int, destroyAfterUse bool) *Item {
	itemName := fmt.Sprintf("a wand of %s", untargetedAction.Name())
	baseItem := NewItem(g.state.GetPlayer, itemName, defs.IconItemWand)
	baseItem.SetCategory(defs.ItemCategoryWand)
	baseItem.SetQualityLevel(1)
	baseItem.SetCharges(charges)
	baseItem.SetDestroyAfterUse(destroyAfterUse)
	baseItem.SetUntargetedAction(untargetedAction)

	return baseItem
}
func (g *Factory) GetRandomUntargetedAction(dungeonLevel int, actionUsage defs.ActionUsage) *UntargetedAction {
	possibleActions := make([]*UntargetedAction, 0)
	for _, def := range g.definitions.UntargetedDefinitions {
		if def.DungeonLevel <= dungeonLevel && def.AllowedUsage.Allows(actionUsage) {
			action := g.NewUntargetedActionFromDefinition(def, def.DefaultArgs)
			if action != nil {
				possibleActions = append(possibleActions, action)
			} else {
				g.logError(fmt.Sprintf("no untargeted action found for dungeon level %d", dungeonLevel))
			}
		}
	}

	if len(possibleActions) > 0 {
		return possibleActions[rand.Intn(len(possibleActions))]
	}
	g.logError(fmt.Sprintf("no untargeted action found for dungeon level %d", dungeonLevel))

	return nil
}

func (g *Factory) NewUntargetedActionFromDefinition(def defs.UntargetedActionDefinition, args util.Arguments) *UntargetedAction {
	action := NewUntargetedAction(def.Name, def.Icon)
	action.SetInternalName(def.InternalName)
	action.SetDelay(def.ActivationDelay)
	effectFromName := g.UntargetedEffectFromName(def.InternalName, args)
	if effectFromName == nil {
		g.logError(fmt.Sprintf("no untargeted effect found for name %s", def.InternalName))
		return nil
	}
	action.SetEffect(effectFromName)
	if def.ActivationCondition != "" {
		action.SetCondition(g.NewConditionFromString(def.ActivationCondition))
	}

	return action
}
func (g *Factory) UntargetedEffectFromName(name string, args util.Arguments) *UntargetedEffect {
	allEffects := g.GetAllUntargetedEffects()
	if effect, exists := allEffects[name]; exists {
		untargetedEffect := effect(args)
		return &untargetedEffect
	}
	g.logError(fmt.Sprintf("no untargeted effect found for name %s", name))
	return nil
}

func (g *Factory) GetAllUntargetedEffects() map[string]func(args util.Arguments) UntargetedEffect {
	return map[string]func(args util.Arguments) UntargetedEffect{
		string(ActionNameFirestarter):     g.getFirestarter,
		string(ActionNameBerserk):         g.getBerserkSelfAction,
		string(ActionNameApplyStatusSelf): g.getApplyStatusAction,
		string(ActionNameTeleport):        g.getRandomTeleportAction,
		string(ActionNameBuffSpeed):       g.getSpeedBuff,
		string(ActionNameRecall):          g.getPlayerRecallAction,
		string(ActionNameBless):           g.getBlessing,
		string(ActionNameStoneSkin):       g.getStoneSkinBuff,
		string(ActionNameEnchantWeapon):   g.getEnchantWeapon,
		string(ActionNameEnchantArmor):    g.getEnchantArmor,
		string(ActionNameAcquirement):     g.getAcquirement,
		string(ActionNameMultiShot):         g.getMultiShot,
		string(ActionNameTelefrag):          g.getRandomTelefragAction,
		string(ActionNameReplicate):         g.getReplicate,
		string(ActionNameEarthquake):        g.getEarthquake,
	}
}

func (g *Factory) NewRandomPotion(dungeonLevel int) *Item {
	// get defs
	itemDefinitions := g.definitions.ItemDefinitions

	// filter for dungeon level
	var filtered []defs.ItemDefinition
	for _, itemDef := range itemDefinitions {

		if !itemDef.IsPotion() || !itemDef.CanAppearAt(g.randomSource, dungeonLevel) {
			continue
		}

		filtered = append(filtered, itemDef)
	}

	if len(filtered) == 0 {
		g.logError(fmt.Sprintf("no potion found for dungeon level %d", dungeonLevel))
		return nil
	}

	// pick random definition
	definition := filtered[g.randomSource.Intn(len(filtered))]

	// create item
	return g.NewItemFromDefinition(definition, dungeonLevel)
}

func (g *Factory) NewPotion(untargetedAction *UntargetedAction) *Item {
	name := fmt.Sprintf("a potion of %s", untargetedAction.Name())
	baseItem := NewItem(g.state.GetPlayer, name, defs.IconItemPotionRed)
	baseItem.SetCategory(defs.ItemCategoryPotion)
	baseItem.SetQualityLevel(1)
	baseItem.SetCharges(1)
	baseItem.SetDestroyAfterUse(true)
	baseItem.SetUntargetedAction(untargetedAction)
	return baseItem
}

func (g *Factory) NewRandomScroll(dungeonLevel int) *Item {
	// get defs
	itemDefinitions := g.definitions.ItemDefinitions

	// filter for dungeon level
	var filtered []defs.ItemDefinition
	for _, itemDef := range itemDefinitions {

		if !itemDef.IsScroll() || !itemDef.CanAppearAt(g.randomSource, dungeonLevel) {
			continue
		}

		filtered = append(filtered, itemDef)
	}

	if len(filtered) == 0 {
		g.logError(fmt.Sprintf("no scroll found for dungeon level %d", dungeonLevel))
		return nil
	}

	// pick random definition
	definition := filtered[g.randomSource.Intn(len(filtered))]

	// create item
	return g.NewItemFromDefinition(definition, dungeonLevel)
}

func (g *Factory) NewHiddenNote(itemName string, messagText []util.LabelText, onRead func()) *Item {
	action := NewUntargetedAction("read", defs.IconItemScroll)
	//action.SetDescription(fmt.Sprintf("TODO: Description for %s", def.Name))

	effect := g.getReadMessage(messagText, onRead)
	action.SetEffect(&effect)
	action.SetNoTimeNeeded()

	scroll := g.NewScroll(action)
	scroll.SetName(itemName)
	scroll.SetUnlimitedCharges()
	scroll.ShowNameOnlyInTooltip()
	scroll.Hide()
	return scroll
}
func (g *Factory) NewScroll(untargetedAction *UntargetedAction) *Item {
	itemName := fmt.Sprintf("a scroll of %s", untargetedAction.Name())
	baseItem := NewItem(g.state.GetPlayer, itemName, defs.IconItemScroll)
	baseItem.SetCategory(defs.ItemCategoryScroll)
	baseItem.SetQualityLevel(1)
	baseItem.SetCharges(1)
	baseItem.SetDestroyAfterUse(true)
	baseItem.SetUntargetedAction(untargetedAction)
	return baseItem
}

func (g *Factory) NewSign(inscription []string) *BumpObject {
	return NewBumpObject("a sign", defs.IconSign, func(actor *Actor, startAnimating func() bool) {
		g.gui().OpenTextWindow(util.ToLabelText(inscription))
	})
}

func (g *Factory) SetQueueAfterTurn(afterTurn func(call func())) {
	g.queueAfterTurn = afterTurn
}

func (g *Factory) SetAudioPlayer(player *audio.Player) {
	g.audioPlayer = player
}
func (g *Factory) SetUserInterface(ui func() UserInterface) {
	g.gui = ui
}

func (g *Factory) NewPoisonedEffect(turnsLeft int, damagePerRound int) *ParametrizedStatusEffect {
	return NewParametrizedStatusEffect(StatusDamageOverTime, func(a *Actor) {
		g.actionBroker.Execute(NewChangeResourceAction(AbstractCause("poison"), a, stats.Health, float64(-damagePerRound)))
	}, turnsLeft)
}

func (g *Factory) NewBarrel() Object {
	barrel := NewExplodingBarrel(func(target geometry.Point) {
		g.state.currentMap.RemoveObjectAt(target)
		g.actionBroker.QueueForExecution(NewExplosionAction(AbstractCause("Barrel"), target, 3, 3, 5))
	})
	return barrel
}
