package defs

import (
	"fmt"
	"math/rand"
	"rpg/stats"
	"rpg/util"
)

/* We got

AddToBaseHealth,
AddToBaseMana,
AddToBaseStrength,
AddToBaseDexterity,
AddToBaseIntelligence,
AddToDamage,
StealHealth,
StealMana,
AddDamage,
AddArmor,

*/

/*
	We need

+ Strength (-10 to +30) - DONE
+ Dexterity (-10 to +30) - DONE
+ Intelligence (-10 to +30) - DONE
+ All Attributes - WON'T DO
+ HP (-25 to +100) - DONE
+ Mana  (-25 to +100) - DONE
% Steal Life (3% to 5%) not cumulative - DONE
% Steal Mana (3% to 5%) not cumulative - DONE
+% Armor Class (-100% to +200%) - DONE
+% Damage Done (-100% to +280%) - DONE
+ Damage Done (1 to 20) - DONE
- Damage Taken (1 to 6)
+ Fire Damage (1 to 20)
+ Lightning Damage (1 to 20)
+% Vision Radius (-40 to +40) - DONE
+% Resist Magic
+% Resist Fire
+% Resist Lightning
+% Resist All
Fragility (Durability = 1)
+% Durability (-75% to +200%)
Indestructible
Ethereal (More power but cannot be repaired)
+% Attack Speed (-50% to +100%)
+ Spell Level (1 to 2)
x Charges (2 to 3)
Armor Penetration (2 to 24) - DONE
Knockback
Fast Block
Thorns (damage to attacker) - DONE
5% chance of doing 3x damage - DONE
Each attack does x0-6 damage. Usually 2x. (Not against unique monsters)
2x damage to monsters, 1x damage to self
*/
func GetAllModifiers() map[string]func(source *rand.Rand, level int) *stats.Modifier {
	modifierFunctions := map[string]func(source *rand.Rand, level int) *stats.Modifier{
		"health":            AddToBaseHealth,
		"vision_range":      AddToBaseVisionRange,
		"armor_class":       AddToArmorClass,
		"armor_multiplier":  AddArmor,
		"attack_speed":      AddToAttackSpeed,
		"move_speed":        AddToMoveSpeed,
		"speed":             AddToBaseSpeed,
	}
	return modifierFunctions
}
func NewModifiersForItem(source *rand.Rand, itemLevel int, count int) []*stats.Modifier {
	modifierFunctions := valuesOf(GetAllModifiers())
	if count > len(modifierFunctions) {
		count = len(modifierFunctions)
	}
	var modifiers []*stats.Modifier
	randomIndices := source.Perm(len(modifierFunctions))
	for i := 0; i < count; i++ {
		modifiers = append(modifiers, modifierFunctions[randomIndices[i]](source, itemLevel))
	}
	return modifiers
}

func NewModifierFromNameWithDungeonLevel(modName string, dungeonLevel int) *stats.Modifier {
	source := rand.New(rand.NewSource(0))
	modifierFunctions := GetAllModifiers()
	modifierFunction := modifierFunctions[modName]
	return modifierFunction(source, dungeonLevel)
}

func NewModifierFromNameWithDiceBonus(modName string, value util.Dice) *stats.Modifier {
	source := rand.New(rand.NewSource(0))
	modifierFunctions := GetAllModifiers()
	modifierFunction := modifierFunctions[modName]
	modifier := modifierFunction(source, 1)
	modifier.Value = float64(value.Roll())
	return modifier
}


func AddToArmorClass(source *rand.Rand, level int) *stats.Modifier {
	levelTable := []ModRow[int, int]{
		{1, util.NewTuple(1, 2)},
		{3, util.NewTuple(2, 4)},
		{5, util.NewTuple(4, 8)},
		{7, util.NewTuple(8, 10)},
	}
	interval := chooseIntervalFromTable(source, level, levelTable)
	acAdded := randIntervalValue(source, interval)
	mod := &stats.Modifier{
		Description:    "of protection",
		Attribute:      stats.ArmorClass,
		Value:          float64(acAdded),
		Mode:           stats.ModifyAdditive,
		IsPercentBased: false,
	}

	return mod
}
func AddToBaseHealth(source *rand.Rand, level int) *stats.Modifier {
	levelTable := []ModRow[int, int]{
		{4, util.NewTuple(-25, -11)},
		{1, util.NewTuple(-10, -1)},
		{1, util.NewTuple(10, 15)},
		{5, util.NewTuple(16, 20)},
		{9, util.NewTuple(21, 30)},
		{15, util.NewTuple(31, 40)},
		{21, util.NewTuple(41, 50)},
		{27, util.NewTuple(51, 60)},
		{35, util.NewTuple(61, 80)},
		{60, util.NewTuple(81, 100)},
	}
	interval := chooseIntervalFromTable(source, level, levelTable)
	return AddToBaseResource(source, stats.MaxHealth, interval)
}

func AddToBaseSpeed(source *rand.Rand, level int) *stats.Modifier {
	interval := util.NewTuple(1, level)
	return AddToBaseAttributeRandomly(source, stats.BaseSpeed, interval)
}

func AddToAttackSpeed(source *rand.Rand, level int) *stats.Modifier {
	// 1.0 would mean 100% increase in attack speed
	maxBonus := float64(level) * 0.1
	modifier := &stats.Modifier{
		Attribute:      stats.AttackSpeedFactor,
		Description:    "of fury",
		Value:          source.Float64() * maxBonus,
		Mode:           stats.ModifyAdditive,
		IsPercentBased: true,
	}

	return modifier
}

func AddToMoveSpeed(source *rand.Rand, level int) *stats.Modifier {
	// 1.0 would mean 100% increase in attack speed
	maxBonus := float64(level) * 0.1
	modifier := &stats.Modifier{
		Attribute:      stats.MoveSpeedFactor,
		Description:    "of running",
		Value:          source.Float64() * maxBonus,
		Mode:           stats.ModifyAdditive,
		IsPercentBased: true,
	}

	return modifier
}
func AddToBaseVisionRange(source *rand.Rand, level int) *stats.Modifier {
	interval := util.NewTuple(-7, 16)
	return AddToBaseAttributeRandomly(source, stats.VisionRange, interval)
}

func AddArmor(source *rand.Rand, level int) *stats.Modifier {
	table := []ModRow[int, int]{
		{3, util.NewTuple(-1, -0)},
		{1, util.NewTuple(0, 1)},
		{6, util.NewTuple(1, 2)},
	}
	interval := chooseIntervalFromTable(source, level, table)
	return AddDamageOrArmor(source, stats.ArmorClass, interval)
}

func StealBaseResource(source *rand.Rand, randomAttribute stats.Name, minMax util.Tuple[float64, float64]) *stats.Modifier {
	intervalMin := minMax.Item1
	intervalMax := minMax.Item2
	var bonusValue float64
	if intervalMin == intervalMax {
		bonusValue = float64(intervalMin)
	} else {
		bonusValue = float64(source.Float64()*(intervalMax-intervalMin) + intervalMin)
	}
	modifier := &stats.Modifier{
		Attribute:      randomAttribute,
		Description:    fmt.Sprintf("of %s stealing", randomAttribute),
		Value:          bonusValue,
		Mode:           stats.ModifyAdditive,
		IsPercentBased: true,
	}

	return modifier
}

func AddDamageOrArmor(source *rand.Rand, randomAttribute stats.Name, minMax util.Tuple[int, int]) *stats.Modifier {
	intervalMin := minMax.Item1
	intervalMax := minMax.Item2

	var bonusValue int
	if intervalMin == intervalMax {
		bonusValue = int(intervalMin)
	} else {
		bonusValue = source.Intn(intervalMax-intervalMin) + intervalMin
	}
	bonusValue += 1.0

	modifier := &stats.Modifier{
		Attribute:      randomAttribute,
		Description:    fmt.Sprintf("of %s", randomAttribute),
		Value:          float64(bonusValue),
		Mode:           stats.ModifyAdditive,
		IsPercentBased: false,
	}

	return modifier
}

func AddToBaseResource(source *rand.Rand, randomAttribute stats.Name, minMax util.Tuple[int, int]) *stats.Modifier {
	intervalMin := minMax.Item1
	intervalMax := minMax.Item2
	var bonusValue float64
	if intervalMin == intervalMax {
		bonusValue = float64(intervalMin)
	} else {
		bonusValue = float64(source.Intn(intervalMax-intervalMin) + intervalMin)
	}
	modifier := &stats.Modifier{
		Attribute:      randomAttribute,
		Description:    fmt.Sprintf("of %s", randomAttribute),
		Value:          bonusValue,
		Mode:           stats.ModifyAdditive,
		IsPercentBased: false,
	}

	return modifier
}
func AddToBaseAttributeRandomly(source *rand.Rand, randomAttribute stats.Name, minMax util.Tuple[int, int]) *stats.Modifier {
	bonusValue := randIntervalValue(source, minMax)
	return AddToBaseAttribute(source, randomAttribute, bonusValue)
}

func AddToBaseAttributeFromRoll(source *rand.Rand, randomAttribute stats.Name, minMax util.Dice) *stats.Modifier {
	bonusValue := minMax.Roll()
	return AddToBaseAttribute(source, randomAttribute, float64(bonusValue))
}

func AddToBaseAttribute(source *rand.Rand, randomAttribute stats.Name, bonusValue float64) *stats.Modifier {
	modifier := &stats.Modifier{
		Attribute:      randomAttribute,
		Description:    fmt.Sprintf("of %s", randomAttribute),
		Value:          bonusValue,
		Mode:           stats.ModifyAdditive,
		IsPercentBased: false,
	}

	return modifier
}

func randIntervalValue(source *rand.Rand, minMax util.Tuple[int, int]) float64 {
	intervalMin := minMax.Item1
	intervalMax := minMax.Item2

	var bonusValue float64
	if intervalMin == intervalMax {
		bonusValue = float64(intervalMin)
	} else {
		bonusValue = float64(source.Intn(intervalMax-intervalMin+1) + intervalMin)
	}
	return bonusValue
}

type ModRow[T1 any, T2 any] struct {
	QualityLevel int
	MinMax       util.Tuple[T1, T2]
}

func chooseIntervalFromTable[T1 any, T2 any](source *rand.Rand, level int, levelTable []ModRow[T1, T2]) util.Tuple[T1, T2] {
	minQLevel := level / 2
	maxQLevel := level

	allowedIntervals := make([]util.Tuple[T1, T2], 0)
	for _, row := range levelTable {
		if row.QualityLevel >= minQLevel && row.QualityLevel <= maxQLevel {
			allowedIntervals = append(allowedIntervals, row.MinMax)
		}
	}

	if len(allowedIntervals) == 0 {
		println(fmt.Sprintf("ERR?! No allowed value intervals for level %d! Choosing randomly.", level))
		return levelTable[source.Intn(len(levelTable))].MinMax
	}
	interval := allowedIntervals[source.Intn(len(allowedIntervals))]
	return interval
}
func valuesOf(modifiers map[string]func(source *rand.Rand, level int) *stats.Modifier) []func(source *rand.Rand, level int) *stats.Modifier {
	result := make([]func(source *rand.Rand, level int) *stats.Modifier, 0)
	for _, value := range modifiers {
		result = append(result, value)
	}
	return result
}
