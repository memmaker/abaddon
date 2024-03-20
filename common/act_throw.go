package common

import (
	"math/rand"
	"rpg/geometry"
	"rpg/stats"
	"rpg/util"
)

type ThrowAction struct {
	BaseAction
	damageFactor float64
	attacker     *Actor
	damageAmount int
	defender     *Actor
	targetPos    geometry.Point
	thrownItem   *Item
}

func (d *ThrowAction) GetTotalDamageDealt() int {
	return d.damageAmount
}

func NewThrowAction(source Cause, attacker *Actor, thrownItem *Item, targetPos geometry.Point) *ThrowAction {
	return &ThrowAction{
		BaseAction:   BaseAction{cause: source, name: ActionNameThrow},
		attacker:     attacker,
		targetPos:    targetPos,
		thrownItem:   thrownItem,
		damageFactor: 1.0,
	}
}

func (d *ThrowAction) Execute(context ActionContext) []Action {
	//audio := engine.GetAudio()
	//audio.PlayCue("bow_fire")

	d.thrownItem.SetDrawnToScreen(false) // hide for animation

	throwInfo := d.thrownItem.GetThrowStats()

	var actions []Action

	// unequip
	if d.attacker.GetEquipment().IsEquipped(d.thrownItem) {
		unequipAction := NewUnequipItemAction(d.Cause(), d.attacker, d.thrownItem)
		actions = append(actions, unequipAction)
	}
	// remove from inventory & place on map
	moveItem := NewMoveItemAction(d.Cause(), d.attacker, d.targetPos, d.thrownItem)
	moveItem.SetChanceToBreak(throwInfo.ChanceToBreak)
	actions = append(actions, moveItem)

	currentMap := context.GameState.GetMap()
	if currentMap.IsActorAt(d.targetPos) {
		d.defender = currentMap.ActorAt(d.targetPos)

		hitChance := CalculateThrowHitChance(d.attacker, d.defender)

		doesHit := hitChance > 0 && hitChance >= rand.Float64()
		if doesHit && throwInfo.Damage.NotZero() { // deal direct throw damage
			damage := throwInfo.Damage.PyramidRandom()
			logAttack(context, d.attacker, d.defender, damage)
			actions = append(actions, NewChangeResourceAction(NewActorCause(ActionNameThrow, d.attacker), d.defender, stats.Health, float64(-damage)))
			d.damageAmount += damage
		}
	} else if currentMap.IsObjectAt(d.targetPos) && throwInfo.Damage.NotZero() {
		damage := throwInfo.Damage.PyramidRandom()
		objectAt := currentMap.ObjectAt(d.targetPos)
		objectAt.DamageReceived(damage)
	}
	if throwInfo.OnHitEffect != nil { // apply generic effect at target location
		hitLocations := throwInfo.HitZone.GetHitPositions(d.attacker.MapPosition(), d.targetPos)
		throwInfo.OnHitEffect.Activate(d.attacker, d.targetPos, hitLocations)
	}

	return actions
}

func (d *ThrowAction) GetAnimations() []Animation {
	animation := NewProjectileDirectionAnimation(d.attacker, d.attacker.MapPosition(), d.targetPos, 0.04, 41, 1)
	animation.SetIcon(d.thrownItem.Icon(0))
	animation.SetOnFinish(func() {
		d.thrownItem.SetDrawnToScreen(true)
	})
	return OneAnimation(animation)
	//return OneAnimation(NewProjectileAnimation(d.attacker, d.flightPath, 0.02, 41, 1))
	//return &DummyAnimation{time: 5, cancelRunningAnimations: true}
}
func (d *ThrowAction) GetDamageFactor() float64 {
	return d.damageFactor
}

func (d *ThrowAction) SetDamageFactor(factor float64) {
	d.damageFactor = factor
}

func (d *ThrowAction) GetDamageDealt() int {
	return d.damageAmount
}

func (d *ThrowAction) GetAttacker() *Actor {
	return d.attacker
}

func (d *ThrowAction) GetDelayInAut() int {
	return 100 / d.attacker.GetAttackSpeed(RangedAttack)
}

func (d *ThrowAction) GetDefender() interface{} {
	return d.defender
}

func (d *ThrowAction) GetTargetPos() geometry.Point {
	return d.targetPos
}



func CalculateThrowHitChance(attacker, defender *Actor) float64 {
	distance := geometry.Distance(defender.MapPosition(), attacker.MapPosition())

	// dex modifier only applies in full to skill based ranged weapons
	// => in particular, crossbows don't get this penalty


	maxRange := 10.0
	accuracyAtMaxRange := 0.5
	accuracyAtMinRange := 0.9

	if distance > maxRange {
		dropOffPerTile := 0.3
		delta := distance - maxRange
		return util.Clamp(0.0, 1.0, accuracyAtMaxRange - (delta * dropOffPerTile))
	}

	if distance <= 1 {
		return util.Clamp(0.0, 1.0, accuracyAtMinRange)
	}

	// interpolate

	// min range == 2
	normalizedMaxRange := maxRange - 1
	percent := (distance - 1) / normalizedMaxRange
	accuracy := util.Clamp(0.0, 1.0, accuracyAtMinRange + (percent * (accuracyAtMaxRange - accuracyAtMinRange)))

	return accuracy
}