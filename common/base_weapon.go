package common

import (
	"fmt"
	"rpg/defs"
	"rpg/stats"
	"rpg/util"
)

type ThrowStats struct {
	Damage        stats.Interval
	OnHitEffect   *HitEffect
	ChanceToBreak float64
	HitZone       *AoEResolver
}

func (s ThrowStats) Clone() *ThrowStats {
	return &ThrowStats{
		Damage:        s.Damage,
		OnHitEffect:   s.OnHitEffect,
		ChanceToBreak: s.ChanceToBreak,
		HitZone:       s.HitZone,
	}
}

func (s ThrowStats) ToLabelText() []util.LabelText {
	var throwInfo []util.LabelText
	if s.Damage.NotZero() {
		throwInfo = append(throwInfo, util.LabelText{
			Text:      fmt.Sprintf("Damage when thrown: %s", s.Damage.ToString()),
			TextColor: defs.OffWhite,
		})
	}
	if s.ChanceToBreak > 0 {
		throwInfo = append(throwInfo, util.LabelText{
			Text:      fmt.Sprintf("Chance to break when thrown: %.0f%%", s.ChanceToBreak*100),
			TextColor: defs.OffWhite,
		})
	}
	return throwInfo
}

type WeaponStats struct {
	Damage                   util.Dice
	Type                     defs.WeaponType
	AttackSpeedFactor        float64
	MeleeRange               int
	Plus                     int // Plus is enchantment that gets added to damage
	MaxRange                 int
	AccuracyAtMaxRange       float64
	AccuracyAtMinRange       float64
	ArmorPenetration         int
	ProjectileIcon           int32
	IsTwoHanded       bool
	SpecialDamageDice map[string]func(player *Actor) util.Dice
}
func (s WeaponStats) ToString(player *Actor) []util.LabelText {
	damageAsString := s.Damage.String()
	var baseInfo []util.LabelText
	if s.SpecialDamageDice != nil {
		for internalEnemyName, getDice := range s.SpecialDamageDice {
			damageOnly := getDice(player).String()
			damageAsString = fmt.Sprintf("%s against %s", damageOnly, niceNameFromInternalEnemyName(internalEnemyName))
		}
		baseInfo = []util.LabelText{
			util.LabelTextFromStringWithColor(fmt.Sprintf("Weapon Type: %s", s.Type.ToString()), defs.OffWhite),
			util.LabelTextFromStringWithColor(fmt.Sprintf("Damage: %s", s.Damage.String()), defs.OffWhite),
			util.LabelTextFromStringWithColor(fmt.Sprintf("Slay: %s", damageAsString), defs.OffWhite),
			util.LabelTextFromStringWithColor(fmt.Sprintf("Enchantment: +%d", s.Plus), defs.OffWhite),
			util.LabelTextFromStringWithColor(fmt.Sprintf("Attack Speed: %.2f", s.AttackSpeedFactor), defs.OffWhite),
		}
	} else {
		baseInfo = []util.LabelText{
			util.LabelTextFromStringWithColor(fmt.Sprintf("Weapon Type: %s", s.Type.ToString()), defs.OffWhite),
			util.LabelTextFromStringWithColor(fmt.Sprintf("Damage: %s", s.Damage.String()), defs.OffWhite),
			util.LabelTextFromStringWithColor(fmt.Sprintf("Enchantment: +%d", s.Plus), defs.OffWhite),
			util.LabelTextFromStringWithColor(fmt.Sprintf("Attack Speed: %.2f", s.AttackSpeedFactor), defs.OffWhite),
		}
	}

	return baseInfo
}

func niceNameFromInternalEnemyName(name string) string {
	return "King Kirioth" // 7drl only hack
}
func (s WeaponStats) ToTableString() []string {
	baseInfo := []util.TableRow{
		{Label: "Type:", Columns: []string{s.Type.ToString()}},
		{Label: "Damage:", Columns: []string{s.Damage.String()}},
		{Label: "Attack Speed:", Columns: []string{fmt.Sprintf("%.2f", s.AttackSpeedFactor)}},
	}
	var weaponStats []string
	weaponStats = util.TableLayout(baseInfo)

	return weaponStats
}
func (s WeaponStats) IsMeleeWeapon() bool {
	return s.Type == defs.WeaponTypeSword ||
		s.Type == defs.WeaponTypeAxe ||
		s.Type == defs.WeaponTypeMace ||
		s.Type == defs.WeaponTypeScythe ||
		s.Type == defs.WeaponTypeSpear ||
		s.Type == defs.WeaponTypeDagger
}

func (s WeaponStats) IsRangedWeapon() bool {
	return s.Type == defs.WeaponTypeBow ||
		s.Type == defs.WeaponTypeCrossbow ||
		s.Type == defs.WeaponTypeSling
}

func (s WeaponStats) NameParts() string {
	// angband style (3, +3)
	return fmt.Sprintf("(%s, +%d)", s.Damage.ShortString(), s.Plus)
}

func (s WeaponStats) Clone() *WeaponStats {
	return &WeaponStats{
		Damage:                      s.Damage,
		Type:               s.Type,
		AttackSpeedFactor:  s.AttackSpeedFactor,
		MeleeRange:         s.MeleeRange,
		Plus:               s.Plus,
		MaxRange:           s.MaxRange,
		AccuracyAtMaxRange: s.AccuracyAtMaxRange,
		AccuracyAtMinRange: s.AccuracyAtMinRange,
		ArmorPenetration:   s.ArmorPenetration,
		ProjectileIcon:     s.ProjectileIcon,
		IsTwoHanded:        s.IsTwoHanded,
		SpecialDamageDice:  s.SpecialDamageDice,
	}
}

func (s WeaponStats) CanBeFurtherEnchanted() bool {
	return s.Plus < 3
}

func (s WeaponStats) GetInventorySlot() defs.SlotName {
	if s.IsRangedWeapon(){
		return defs.SlotNameMissileLauncher
	}
	if s.Type == defs.WeaponTypeShield {
		return defs.SlotNameOffHand
	}
	if s.Type == defs.WeaponTypeScythe || s.Type == defs.WeaponTypeSpear {
		return defs.SlotNameTwoHandedWeapon
	}
	if s.IsTwoHanded {
		return defs.SlotNameTwoHandedWeapon
	}
	return defs.SlotNameMainHand
}

func (s WeaponStats) WithSpecialDamageDice(getDamageDice func(player *Actor) util.Dice) WeaponStats {
	s.SpecialDamageDice = map[string]func(player *Actor) util.Dice{
		"kirioth": getDamageDice,
	}
	return s
}

func (s WeaponStats) HasSpecialDamage() bool {
	return s.SpecialDamageDice != nil
}

func (s WeaponStats) GetSpecialDamageDice(actor *Actor, enemyName string) util.Dice {
	if s.SpecialDamageDice != nil && s.SpecialDamageDice[enemyName] != nil {
		return s.SpecialDamageDice[enemyName](actor)
	}
	return s.Damage
}
