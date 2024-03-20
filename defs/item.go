package defs

import (
	"math/rand"
	"rpg/recfile"
	"rpg/stats"
	"rpg/util"
)

// item line for typical weapon
// name, icon, dlvl, weapon_type, weapon_damage, weapon_speed
type ItemDefinition struct {
	Name         string
	InternalName string
	Icon         int32
	DLvl         int // min dungeon level at which this item can appear (=native depth)
	MaxLevel     int // max level at which this item can appear

	OutOfDepthChance float64 // chance to appear out of depth
	MinLevelForOoD   int

	UntargetedAction string
	TargetedAction   string

	InventoryCategory ItemCategory
	Armor             ArmorDefinition
	Weapon            WeaponDefinition
	Throw             ThrowDefinition
	Cost              int
	Modifiers         []string

	Charges int
}

func NewItemDefinitionFromRecord(record recfile.Record) ItemDefinition {
	itemDefinition := ItemDefinition{
		Weapon: WeaponDefinition{
			ProjectileIcon: -1,
		},
	}
	for _, field := range record {
		switch field.Name {
		case "name":
			itemDefinition.Name = field.Value
		case "internal_name":
			itemDefinition.InternalName = field.Value
		case "icon":
			itemDefinition.Icon = field.AsInt32()
		case "dlvl":
			itemDefinition.DLvl = field.AsInt()
		case "cost":
			itemDefinition.Cost = field.AsInt()
		case "charges":
			itemDefinition.Charges = field.AsInt()
		case "max_level":
			itemDefinition.MaxLevel = field.AsInt()
		case "out_of_depth_chance":
			itemDefinition.OutOfDepthChance = field.AsFloat()
		case "min_level_for_ood":
			itemDefinition.MinLevelForOoD = field.AsInt()
		case "category":
			itemDefinition.InventoryCategory = ItemCategoryFromString(field.Value)
		case "armor_type":
			itemDefinition.Armor.ArmorType = ArmorTypeFromString(field.Value)
		case "armor_class":
			itemDefinition.Armor.ArmorClass = stats.NewIntervalFromString(field.Value)
		case "weapon_type":
			itemDefinition.Weapon.WeaponType = WeaponTypeFromString(field.Value)
		case "weapon_two_handed":
			itemDefinition.Weapon.IsTwoHanded = field.AsBool()
		case "weapon_damage":
			itemDefinition.Weapon.Damage = util.ParseDice(field.Value)
		case "weapon_speed":
			itemDefinition.Weapon.Speed = field.AsFloat()
		case "weapon_armor_penetration":
			itemDefinition.Weapon.ArmorPenetration = field.AsInt()
		case "ranged_max_range":
			itemDefinition.Weapon.MaxRange = field.AsInt()
		case "ranged_accuracy_at_max_range":
			itemDefinition.Weapon.AccuracyAtMaxRange = field.AsFloat()
		case "ranged_accuracy_at_min_range":
			itemDefinition.Weapon.AccuracyAtMinRange = field.AsFloat()
		case "ranged_projectile_icon": itemDefinition.Weapon.ProjectileIcon = field.AsInt32()
		case "throw_hit_zone":
			itemDefinition.Throw.HitZone = field.Value
		case "throw_damage":
			itemDefinition.Throw.Damage = stats.NewIntervalFromString(field.Value)
		case "throw_chance_to_break":
			itemDefinition.Throw.ChanceToBreak = field.AsFloat()
		case "throw_on_hit_effect":
			itemDefinition.Throw.OnHitEffect = field.Value
		case "equip_modifier":
			itemDefinition.Modifiers = append(itemDefinition.Modifiers, field.Value)
		case "untargeted_action":
			itemDefinition.UntargetedAction = field.Value
		case "targeted_action":
			itemDefinition.TargetedAction = field.Value
		}
	}

	return itemDefinition
}
func (d ItemDefinition) CanAppearAt(randomSource *rand.Rand, dungeonLevel int) bool {
	// reject early
	if d.MaxLevel > 0 && dungeonLevel > d.MaxLevel {
		return false
	}
	regularAppearance := dungeonLevel >= d.DLvl

	// apply out of depth chance
	oodAppearance := dungeonLevel >= d.MinLevelForOoD && randomSource.Float64() < d.OutOfDepthChance

	return regularAppearance || oodAppearance
}

func (d ItemDefinition) IsArmor() bool {
	return d.Armor.IsValid()
}

func (d ItemDefinition) IsWeapon() bool {
	return d.Weapon.IsValid()
}

func (d ItemDefinition) IsThrow() bool {
	return d.Throw.IsValid()
}

func (d ItemDefinition) IsPotion() bool {
	return d.InventoryCategory == ItemCategoryPotion
}

func (d ItemDefinition) IsScroll() bool {
	return d.InventoryCategory == ItemCategoryScroll
}

func (d ItemDefinition) IsRing() bool {
	return d.InventoryCategory == ItemCategoryRing
}

func (d ItemDefinition) IsAmulet() bool {
	return d.InventoryCategory == ItemCategoryAmulet
}
