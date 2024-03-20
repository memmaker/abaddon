package common

import (
	"math"
	"math/rand"
	"rpg/geometry"
	"rpg/stats"
	"rpg/util"
)

type RangedAttackAction struct {
	BaseAction
	damageFactor    float64
	attacker        *Actor
	damageAmount    int
	defender        *Actor
	targetPos       geometry.Point
	defenders       []*Actor
	additionalOnHitActions func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) []Action
}
func (d *RangedAttackAction) SetAdditionalOnHitActions(actions func(attacker *Actor, target geometry.Point, hitPositions []geometry.Point) []Action) {
	d.additionalOnHitActions = actions
}
func (d *RangedAttackAction) GetTotalDamageDealt() int {
	return d.damageAmount
}

func NewRangedAttackAction(source Cause, attacker *Actor, targetPos geometry.Point) *RangedAttackAction {
	return &RangedAttackAction{
		BaseAction:   BaseAction{cause: source, name: ActionNameRangedAttack},
		attacker:     attacker,
		targetPos:    targetPos,
		damageFactor: 1.0,
	}
}

func (d *RangedAttackAction) Execute(context ActionContext) []Action {
	currentMap := context.GameState.GetMap()
	var actions []Action

	d.damageAmount = 0

	if !currentMap.IsActorAt(d.targetPos) {
		return actions
	}

	d.defender = currentMap.ActorAt(d.targetPos)

	hitChance := CalculateRangedWeaponHitChance(d.attacker, d.defender)


	if rand.Float64() > hitChance {
		// missed the attack
		logMissed(context, d.attacker, d.defender)
		return actions
	}

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

	damage := RollRangedDamage(d.attacker, d.defender, d.damageFactor)
	d.damageAmount = damage

	logAttack(context, d.attacker, d.defender, damage)

	actions = append(actions, NewChangeResourceAction(NewActorCause(ActionNameRangedAttack, d.attacker), d.defender, stats.Health, float64(-damage)))
	if d.additionalOnHitActions != nil {
		actions = append(actions, d.additionalOnHitActions(d.attacker, d.defender.MapPosition(), []geometry.Point{d.defender.MapPosition()})...)
	}

	return actions
}

func RollRangedDamage(attacker, defender *Actor, damageMultiplier float64) int {
	damageRoll := GetRangedDamageRoll(attacker, defender.GetInternalName())

	return CalculateRangedDamage(damageRoll, damageMultiplier, attacker, defender)
}

func CalculateRangedDamage(damageRoll int, damageMultiplier float64, attacker, defender *Actor) int {
	//damageMultiplier *= GetWeaponQualityFactor(attacker)

	damage := int(float64(damageRoll) * damageMultiplier)
	armorPenetration := GetArmorPenetrationOfEquippedRangedWeapon(attacker)

	defenderAC := max(0, defender.GetArmorClass() - armorPenetration)
	defenderADR := ADRFromAC(defenderAC)
	blockedDamage := int(math.Floor(float64(damage) * defenderADR))

	if defender.GetEquipment().HasShieldEquipped() {
		shield := defender.GetEquipment().GetShield()
		shieldAC := max(1, shield.GetArmorStats().GetTotalArmorClass())
		if shieldAC > blockedDamage && rand.Float64() > 0.5 {
			blockedDamage = shieldAC
		}
	}

	damage = max(0, damage-blockedDamage)

	return damage
}

func GetRangedDamageRoll(attacker *Actor, enemyName string) int {
	damageDice := attacker.GetDamageDice(RangedAttack, enemyName)
	damageRoll := damageDice.Roll()
	return damageRoll
}
func GetArmorPenetrationOfEquippedRangedWeapon(attacker *Actor) int {
	if !attacker.GetEquipment().HasRangedWeaponEquipped() {
		return 0
	}
	weaponUsed := attacker.GetEquipment().GetWeapon(RangedAttack).GetWeaponStats()
	return weaponUsed.ArmorPenetration
}

func CalculateRangedWeaponHitChance(attacker, defender *Actor) float64 {
	distance := geometry.Distance(defender.MapPosition(), attacker.MapPosition())
	rangedWeapon := attacker.GetEquipment().GetWeapon(RangedAttack)
	if rangedWeapon == nil {
		return 0
	}

	dexModifier := 0.0

	// dex modifier only applies in full to skill based ranged weapons
	// => in particular, crossbows don't get the penalty part
	weaponStats := rangedWeapon.GetWeaponStats()
	maxRange := float64(weaponStats.MaxRange)
	accuracyAtMaxRange := weaponStats.AccuracyAtMaxRange
	accuracyAtMinRange := weaponStats.AccuracyAtMinRange

	if distance > maxRange {
		dropOffPerTile := 0.3
		delta := distance - maxRange
		return util.Clamp(0.0, 1.0, accuracyAtMaxRange - (delta * dropOffPerTile) + dexModifier)
	}

	if distance <= 1 {
		return util.Clamp(0.0, 1.0, accuracyAtMinRange + dexModifier)
	}

	// interpolate

	// min range == 2
	normalizedMaxRange := maxRange - 1
	percent := (distance - 1) / normalizedMaxRange
	accuracy := util.Clamp(0.0, 1.0, accuracyAtMinRange + (percent * (accuracyAtMaxRange - accuracyAtMinRange)) + dexModifier)

	return accuracy
}

func defaultModifierPattern(attribute int) int {
	switch {
	case attribute <= 3:
		return -3
	case attribute <= 5:
		return -2
	case attribute <= 8:
		return -1
	case attribute <= 12:
		return 0
	case attribute <= 15:
		return 1
	case attribute <= 17:
		return 2
	case attribute >= 18:
		return 3
	}
	return 0
}

func (d *RangedAttackAction) GetAnimations() []Animation {
	weapon := d.attacker.GetEquipment().GetWeapon(RangedAttack).GetWeaponStats()
	projectileIcon := int32(41)
	if weapon.ProjectileIcon > -1 {
		projectileIcon = weapon.ProjectileIcon
	}
	return OneAnimation(NewProjectileDirectionAnimation(d.attacker, d.attacker.MapPosition(), d.targetPos, 0.02, projectileIcon, 1))
	//return OneAnimation(NewProjectileAnimation(d.attacker, d.flightPath, 0.02, 41, 1))
	//return &DummyAnimation{time: 5, cancelRunningAnimations: true}
}
func (d *RangedAttackAction) GetDamageFactor() float64 {
	return d.damageFactor
}

func (d *RangedAttackAction) SetDamageFactor(factor float64) {
	d.damageFactor = factor
}

func (d *RangedAttackAction) GetDamageDealt() int {
	return d.damageAmount
}

func (d *RangedAttackAction) GetAttacker() *Actor {
	return d.attacker
}

func (d *RangedAttackAction) GetDelayInAut() int {
	return 100 / d.attacker.GetAttackSpeed(RangedAttack)
}

func (d *RangedAttackAction) GetDefender() interface{} {
	return d.defender
}

func (d *RangedAttackAction) GetTargetPos() geometry.Point {
	return d.targetPos
}
