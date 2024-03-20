package common

import (
	"fmt"
	"math"
	"math/rand"
	"rpg/defs"
	"rpg/stats"
	"rpg/util"
)

type MeleeAttackAction struct {
	BaseAction
	damageFactor     float64
	attacker         *Actor
	totalDamageDealt int
	defender         *Actor
	disableAnimation bool
	audioCue         string
}

func NewMeleeAttackAction(source Cause, attacker *Actor, defender *Actor) *MeleeAttackAction {
	return &MeleeAttackAction{
		BaseAction:   BaseAction{cause: source, name: ActionNameMeleeAttack},
		attacker:     attacker,
		defender:    defender,
		damageFactor: 1.0,
	}
}

func (d *MeleeAttackAction) Execute(context ActionContext) []Action {
	var actions []Action

	evasionChance := 0.0
	if d.defender.HasStatusEffect(StatusFlying) {
		evasionChance = 0.25
	}

	if rand.Float64() < evasionChance {
		// dodged the attack
		logDodged(context, d.attacker, d.defender)
		// play miss audio cue?
		return actions
	}

	isPlayer := d.attacker == context.GameState.GetPlayer()


	damage := RollMeleeDamage(d.attacker, d.defender, d.damageFactor)


	d.totalDamageDealt = damage

	healthOfDefenderBeforeAttack := d.defender.GetStats().GetResource(stats.Health).Value()

	levelName := context.GameState.GetMap().GetName()
	isHellTown := levelName == string(defs.LevelNameTownHell)

	// 7drl hack
	if d.defender.IsWinCondition() && !isHellTown {
		context.Msg(fmt.Sprintf("%s is not affected by any damage!", d.defender.Name()))
		if rand.Intn(3) == 0 {
			context.Speak(d.defender, defs.SpeechTriggerHurtInTown)
		}
		return nil
	}

	logAttack(context, d.attacker, d.defender, damage)

	actions = append(actions, NewChangeResourceAction(NewActorCause("Melee attack", d.attacker), d.defender, stats.Health, float64(-damage)))

	d.defender.SetLookDirFromPoI(d.attacker.MapPosition())
	d.attacker.SetLookDirFromPoI(d.defender.MapPosition())

	if isPlayer {
		bloodAbsorbed := min(healthOfDefenderBeforeAttack, damage)
		// 7drl hack
		isWieldingVengeance := isPlayer && d.attacker.GetEquipment().HasMeleeWeaponEquipped() && d.attacker.GetEquipment().GetWeapon(MeleeAttack).internalName == "sword_of_vengeance"
		if isWieldingVengeance && bloodAbsorbed > 0 {
			if d.defender.IsWinCondition() {
				d.audioCue = "special_attack"
			}
			attackerStats := d.attacker.GetStats()
			attackerStats.IncrementBy(stats.Blood, bloodAbsorbed)
		}
	}

	return actions
}

func GetVengeanceDamageBonus(wielder *Actor) int {
	attackerStats := wielder.GetStats()
	vengeanceDamageDealt := attackerStats.GetAttribute(stats.Blood)
	bonusPerDamage := 0.1
	vengeanceBonusDamage := int(float64(vengeanceDamageDealt) * bonusPerDamage)
	return vengeanceBonusDamage
}

func RollMeleeDamage(attacker, defender *Actor, damageFactor float64) int {
	damageRoll := GetMeleeDamageRoll(attacker, defender.GetInternalName())
	return CalculateMeleeDamage(damageRoll, damageFactor, attacker, defender)
}

func CalculateMeleeDamage(damageRoll int, damageFactor float64, attacker, defender *Actor) int {
	//attackerStr := attacker.GetStats().GetAttribute(stats.Strength)
	//dndStyleModifier := defaultModifierPattern(attackerStr)
	//rawDamage := max(1, damageRoll + dndStyleModifier)

	//damageFactor *= GetWeaponQualityFactor(attacker)
	damage := int(float64(damageRoll) * damageFactor)

	armorPenetration := GetArmorPenetrationOfEquippedMeleeWeapon(attacker)
	defenderAC := max(0, defender.GetArmorClass() - armorPenetration)
	defenderADR := ADRFromAC(defenderAC)
	blockedDamage := int(math.Ceil(float64(damage) * defenderADR))

	// 50% chance to use the shield, if it's more effective
	if defender.GetEquipment().HasShieldEquipped() {
		shield := defender.GetEquipment().GetShield()
		shieldAC := max(1, shield.GetArmorStats().GetTotalArmorClass())
		if shieldAC > blockedDamage && rand.Float64() > 0.5 {
			blockedDamage = shieldAC
		}
	}

	damage = max(1, damage-blockedDamage)
	return damage
}

func GetArmorPenetrationOfEquippedMeleeWeapon(attacker *Actor) int {
	if !attacker.GetEquipment().HasMeleeWeaponEquipped() {
		return 0
	}
	weaponUsed := attacker.GetEquipment().GetWeapon(MeleeAttack).GetWeaponStats()
	return weaponUsed.ArmorPenetration
}

func GetMeleeDamageRoll(attacker *Actor, enemyName string) int {
	damageDice := attacker.GetDamageDice(MeleeAttack, enemyName)
	damageRoll := damageDice.Roll()
	return damageRoll
}

func (d *MeleeAttackAction) GetAnimations() []Animation {
	if d.disableAnimation {
		return nil
	}
	ticksNeeded := util.SecondsToTicks(0.2)
	animation := &MeleeAttackAnimation{
		actor:       d.attacker,
		targetPos:   d.defender.MapPosition(),
		origin:      d.attacker.MapPosition(),
		ticksNeeded: ticksNeeded,
		ticksLeft:   ticksNeeded,
		audioCue:    d.audioCue,
	}
	return OneAnimation(animation)
}
func (d *MeleeAttackAction) GetDamageFactor() float64 {
	return d.damageFactor
}

func (d *MeleeAttackAction) SetDamageFactor(factor float64) {
	d.damageFactor = factor
}
func (d *MeleeAttackAction) GetAttacker() *Actor {
	return d.attacker
}

func (d *MeleeAttackAction) GetDelayInAut() int {
	return 100 / d.attacker.GetAttackSpeed(MeleeAttack)
}

func (d *MeleeAttackAction) GetDefender() *Actor {
	return d.defender
}

func (d *MeleeAttackAction) GetTotalDamageDealt() int {
	return d.totalDamageDealt
}

func (d *MeleeAttackAction) DisableAnimation() {
	d.disableAnimation = true
}
