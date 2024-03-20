package game

import (
	"rpg/common"
	"rpg/defs"
	"rpg/stats"
	"rpg/util"
)


/*
1. Shekel poisoned the water supply and framed the workers union
  - Report from the workers found at the poisoned water supply

2. Abel backstabbed the skeleton king - Last will of the dead good king (now skeleton king)
  - Last will of the dead good king (now skeleton king), at his tomb

3. Kirioth persuaded both of them to follow his plan and promised the wealth and status
Kirioth sacrificed some poor people when he summoned Lucifer, trying to control him
  - Shekel's notes on Kirioth's plan, security to be used for blackmailing, in Shekel's house

4. Shekel & Kirioth are responsible for the creation of meat jack
  - List of subject for the super soldier experiment, there is a lab in the caves.-- TODO*
*/

func NewMasterSword(state *common.State) *common.Item {
	// spell on attack
	baseItem := common.NewItem(state.GetPlayer, "a master sword", defs.IconItemUniqueSword)
	baseItem.SetInternalName("master_sword")
	baseItem.SetCategory(defs.ItemCategoryWeapon)
	baseItem.SetQualityLevel(1)
	baseItem.SetRarity(common.ItemRarityUnique)
	weaponPart := &common.WeaponStats{
		Damage:            util.ParseDice("2d8+3"),
		Type:              defs.WeaponTypeSword,
		MeleeRange:        1,
		AttackSpeedFactor: 1,
	}

	baseItem.SetWeaponComponent(weaponPart)

	// stat modifiers
	modifiers := []*stats.Modifier{

		{
			Description:    "speedy",
			Attribute:      stats.AttackSpeedFactor,
			Value:          1.2,
			Mode:           stats.ModifyMultiplicative,
			IsPercentBased: true,
		},
	}

	baseItem.SetEquippedModifiers(modifiers)

	// TODO: Diablo Style Loot Branding
	// 1. TRIGGERED ACTION
	// -> X % chance to execute Action Y on Hit/Attack/Kill/BeingHit/BeingAttacked
	//    - Invokable Action X (with optional maximum charges)
	//    - X Damage of Type Y to Attacker on Attack/Hit
	//    - Steal X Mana/HP per hit
	//    - X% to Apply Status Effect X
	// 2. STAT MODIFIER
	// - +X to Skill Y
	// - +U..V Physical/Fire Damage, +W Poison damage over X seconds
	// - +X to Maximum Damage
	// - +X% HP/Mana Regeneration per AUT
	// - Requirements -60%
	// - Resistance / Immunity to Status Effect X
	// - Resistance / Immunity / Absorption to Damage Type X
	return baseItem
}

func flatFunc(i int) func() int {
	return func() int {
		return i
	}
}
