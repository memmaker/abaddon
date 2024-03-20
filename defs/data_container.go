package defs

import (
	"cmp"
	"fmt"
	"math/rand"
	"path"
	"rpg/files"
	"rpg/stats"
	"rpg/util"
	"slices"
)

type DataContainer struct {
	ItemDefinitions       map[string]ItemDefinition
	UniqueItemDefinitions map[string]ItemDefinition

	EnemyDefinitions       map[string]EnemyDefinition
	UniqueEnemyDefinitions map[string]EnemyDefinition

	Modifiers map[string]func(source *rand.Rand, level int) *stats.Modifier

	TargetedDefinitions   map[string]TargetedActionDefinition
	UntargetedDefinitions map[string]UntargetedActionDefinition

	ClassDefinitions  []LoadoutDefinition
	VendorDefinitions map[string]VendorDefinition
}

func (c DataContainer) GetItemCategories() []ItemCategory {
	result := make(map[ItemCategory]bool, 0)
	for _, def := range c.ItemDefinitions {
		result[def.InventoryCategory] = true
	}
	var categories []ItemCategory
	for cat := range result {
		categories = append(categories, cat)
	}
	slices.SortStableFunc(categories, func(i, j ItemCategory) int {
		return cmp.Compare(i, j)
	})

	return categories
}
func LoadGameData(logErr func(err string)) DataContainer {
	printDefLoad := false
	itemDefs := make(map[string]ItemDefinition)
	itemDir := path.Join("assets", "data", "items")
	enemyDir := path.Join("assets", "data", "enemies")
	effectsDir := path.Join("assets", "data", "effects")
	otherDir := path.Join("assets", "data")

	join := path.Join(itemDir, "armor.rec")
	file, _ := files.Open(join)
	armorDefs := ReadItemDefinitions(file, logErr)
	if printDefLoad {
		println(fmt.Sprintf("Loaded %d armor definitions", len(armorDefs)))
	}
	itemDefs = merge(itemDefs, armorDefs)
	file.Close()

	file, _ = files.Open(path.Join(itemDir, "weapons.rec"))
	weaponDefs := ReadItemDefinitions(file, logErr)
	if printDefLoad {
		println(fmt.Sprintf("Loaded %d weapon definitions", len(weaponDefs)))
	}
	itemDefs = merge(itemDefs, weaponDefs)
	file.Close()

	file, _ = files.Open(path.Join(itemDir, "misc.rec"))
	throwableDefs := ReadItemDefinitions(file, logErr)
	if printDefLoad {
		println(fmt.Sprintf("Loaded %d misc. definitions", len(throwableDefs)))
	}
	itemDefs = merge(itemDefs, throwableDefs)
	file.Close()

	file, _ = files.Open(path.Join(itemDir, "accessory.rec"))
	accessoryDefs := ReadItemDefinitions(file, logErr)
	if printDefLoad {
		println(fmt.Sprintf("Loaded %d accessory definitions", len(accessoryDefs)))
	}
	itemDefs = merge(itemDefs, accessoryDefs)
	file.Close()

	file, _ = files.Open(path.Join(itemDir, "light.rec"))
	lightDefs := ReadItemDefinitions(file, logErr)
	if printDefLoad {
		println(fmt.Sprintf("Loaded %d light definitions", len(lightDefs)))
	}
	itemDefs = merge(itemDefs, lightDefs)
	file.Close()

	file, _ = files.Open(path.Join(itemDir, "potions.rec"))
	potionDefs := ReadItemDefinitions(file, logErr)
	if printDefLoad {
		println(fmt.Sprintf("Loaded %d potion definitions", len(potionDefs)))
	}
	itemDefs = merge(itemDefs, potionDefs)
	file.Close()

	file, _ = files.Open(path.Join(itemDir, "scrolls.rec"))
	scrollDefs := ReadItemDefinitions(file, logErr)
	if printDefLoad {
		println(fmt.Sprintf("Loaded %d scroll definitions", len(scrollDefs)))
	}
	itemDefs = merge(itemDefs, scrollDefs)
	file.Close()

	file, _ = files.Open(path.Join(enemyDir, "enemies.rec"))
	enemyDefs := ReadEnemyDefinitions(file, logErr)
	if printDefLoad {
		println(fmt.Sprintf("Loaded %d enemy definitions", len(enemyDefs)))
	}
	file.Close()

	file, _ = files.Open(path.Join(enemyDir, "unique_enemies.rec"))
	uniquEnemyDefs := ReadEnemyDefinitions(file, logErr)
	if printDefLoad {
		println(fmt.Sprintf("Loaded %d unique enemy definitions", len(uniquEnemyDefs)))
	}
	file.Close()

	file, _ = files.Open(path.Join(effectsDir, "untargeted.rec"))
	untargetedDefs := ReadUntargetedDefinitions(file, logErr)
	if printDefLoad {
		println(fmt.Sprintf("Loaded %d untargeted effect definitions", len(untargetedDefs)))
	}
	file.Close()

	file, _ = files.Open(path.Join(effectsDir, "targeted.rec"))
	targetedDefs := ReadTargetedDefinitions(file, logErr)
	if printDefLoad {
		println(fmt.Sprintf("Loaded %d targeted effect definitions", len(targetedDefs)))
	}
	file.Close()

	file, _ = files.Open(path.Join(otherDir, "stores.rec"))
	vendorDefs := ReadVendorDefinitions(file)
	if printDefLoad {
		println(fmt.Sprintf("Loaded %d store definitions", len(vendorDefs)))
	}
	file.Close()

	file, _ = files.Open(path.Join(otherDir, "classes.rec"))
	classDefs := ReadClassDefinitions(file)
	if printDefLoad {
		println(fmt.Sprintf("Loaded %d class definitions", len(classDefs)))
	}
	file.Close()

	file, _ = files.Open(path.Join(itemDir, "unique_items.rec"))
	uniqueItemDefs := ReadItemDefinitions(file, logErr)
	if printDefLoad {
		println(fmt.Sprintf("Loaded %d unique item definitions", len(uniqueItemDefs)))
	}
	file.Close()

	return DataContainer{
		ItemDefinitions:        itemDefs,
		UniqueItemDefinitions:  uniqueItemDefs,
		EnemyDefinitions:       enemyDefs,
		UniqueEnemyDefinitions: uniquEnemyDefs,
		TargetedDefinitions:    targetedDefs,
		UntargetedDefinitions:  untargetedDefs,
		VendorDefinitions:      vendorDefs,
		ClassDefinitions:       classDefs,
		Modifiers:              GetAllModifiers(),
	}
}

type BuiltInDataContainer struct {
	Modifiers         map[string]bool
	HitZones          map[string]bool
	HitEffects        map[string]bool
	TargetingRules    map[string]bool
	UntargetedEffects map[string]bool
}

func validateDefinitions(gameData DataContainer, oobData BuiltInDataContainer, logErr func(err string)) { // needs to be higher in the hierarchy..
	modifiersAvailable := oobData.Modifiers
	hitZonesAvailable := oobData.HitZones
	hitEffectsAvailable := oobData.HitEffects
	targetRulesAvailable := oobData.TargetingRules
	untargetedEffectsAvailable := oobData.UntargetedEffects

	/*
	   modifiersAvailable := g.GetAllModifiers()
	   hitZonesAvailable := g.GetAllHitZones()
	   hitEffectsAvailable := g.GetAllHitEffects()
	   targetRulesAvailable := g.GetAllTargetingRules()
	   untargetedEffectsAvailable := g.GetAllUntargetedEffects()
	*/
	for _, itemDef := range gameData.ItemDefinitions {
		if !itemDef.InventoryCategory.IsValid() {
			logErr(fmt.Sprintf("Item %s has invalid category %d", itemDef.Name, itemDef.InventoryCategory))
		}
		if _, targetedExists := gameData.TargetedDefinitions[itemDef.TargetedAction]; itemDef.TargetedAction != "" && !targetedExists {
			logErr(fmt.Sprintf("Item %s has invalid targeted action %s", itemDef.Name, itemDef.TargetedAction))
		}
		if itemDef.UntargetedAction != "" {
			untName := itemDef.UntargetedAction
			if util.LooksLikeAFunction(itemDef.UntargetedAction) {
				untName, _ = util.GetNameAndArgs(itemDef.UntargetedAction)
			}
			if _, untargetedExists := gameData.UntargetedDefinitions[untName]; !untargetedExists {
				logErr(fmt.Sprintf("Item %s has invalid untargeted action %s", itemDef.Name, itemDef.UntargetedAction))
			}
		}

		for _, modifier := range itemDef.Modifiers {
			if util.LooksLikeAFunction(modifier) {
				modifier, _ = util.GetNameAndArgs(modifier)
			}
			if _, exists := modifiersAvailable[modifier]; !exists {
				logErr(fmt.Sprintf("Item %s has invalid modifier %s", itemDef.Name, modifier))
			}
		}
		if itemDef.IsThrow() && itemDef.Throw.IsValid() {
			hitZone := itemDef.Throw.HitZone
			if util.LooksLikeAFunction(hitZone) {
				hitZone, _ = util.GetNameAndArgs(hitZone)
			}
			if _, exists := hitZonesAvailable[hitZone]; hitZone != "" && !exists {
				logErr(fmt.Sprintf("Item %s has invalid throw hit zone %s", itemDef.Name, itemDef.Throw.HitZone))
			}
			hitEffect := itemDef.Throw.OnHitEffect
			if util.LooksLikeAFunction(hitEffect) {
				hitEffect, _ = util.GetNameAndArgs(hitEffect)
			}
			if _, exists := hitEffectsAvailable[hitEffect]; hitEffect != "" && !exists {
				logErr(fmt.Sprintf("Item %s has invalid throw hit effect %s", itemDef.Name, itemDef.Throw.OnHitEffect))
			}
		}
	}

	for _, targetedDef := range gameData.TargetedDefinitions {
		hitZone := targetedDef.HitZone
		if util.LooksLikeAFunction(hitZone) {
			hitZone, _ = util.GetNameAndArgs(hitZone)
		}
		if _, exists := hitZonesAvailable[hitZone]; !exists {
			logErr(fmt.Sprintf("Targeted effect %s has invalid hit zone %s", targetedDef.Name, hitZone))
		}
		hitEffect := targetedDef.OnHitEffect
		if util.LooksLikeAFunction(hitEffect) {
			hitEffect, _ = util.GetNameAndArgs(hitEffect)
		}
		if _, exists := hitEffectsAvailable[hitEffect]; !exists {
			logErr(fmt.Sprintf("Targeted effect %s has invalid hit effect %s", targetedDef.Name, hitEffect))
		}
		if _, exists := targetRulesAvailable[targetedDef.Targeting]; !exists {
			logErr(fmt.Sprintf("Targeted effect %s has invalid target rule %s", targetedDef.Name, targetedDef.Targeting))
		}
	}

	for _, untargetedDef := range gameData.UntargetedDefinitions {
		if _, exists := untargetedEffectsAvailable[untargetedDef.InternalName]; !exists {
			logErr(fmt.Sprintf("Untargeted effect %s is unknown", untargetedDef.InternalName))
		}
	}

	for _, enemyDef := range gameData.EnemyDefinitions {
		for _, itemName := range enemyDef.Equipment {
			if _, exists := gameData.ItemDefinitions[itemName]; !exists {
				logErr(fmt.Sprintf("Enemy %s has invalid equipment %s", enemyDef.Name, itemName))
			}
		}
		for _, targetedName := range enemyDef.TargetedSkills {
			if _, exists := gameData.TargetedDefinitions[targetedName]; !exists {
				logErr(fmt.Sprintf("Enemy %s has invalid targeted skill %s", enemyDef.Name, targetedName))
			}
		}
		for _, untargetedName := range enemyDef.UntargetedSkills {
			if _, exists := gameData.UntargetedDefinitions[untargetedName]; !exists {
				logErr(fmt.Sprintf("Enemy %s has invalid untargeted skill %s", enemyDef.Name, untargetedName))
			}
		}
	}

	for _, vendorDef := range gameData.VendorDefinitions {
		for _, itemName := range vendorDef.Always {
			if _, exists := gameData.ItemDefinitions[itemName]; !exists {
				logErr(fmt.Sprintf("Store %s has invalid always item %s", vendorDef.InternalName, itemName))
			}
		}
		for _, itemName := range vendorDef.Sometimes {
			if _, exists := gameData.ItemDefinitions[itemName]; !exists {
				logErr(fmt.Sprintf("Store %s has invalid sometimes item %s", vendorDef.InternalName, itemName))
			}
		}
	}
}

func merge(defs map[string]ItemDefinition, other map[string]ItemDefinition) map[string]ItemDefinition {
	for key, value := range other {
		if _, ok := defs[key]; ok {
			println("duplicate key", key)
		}
		defs[key] = value
	}
	return defs
}
