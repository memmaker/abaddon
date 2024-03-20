package common

import (
	"math/rand"
	"rpg/geometry"
	"rpg/stats"
)

type ExplosionAction struct {
	BaseAction
	damageFactor      float64
	damageAmount      int
	radius            int
	pushRange         int
	origin            geometry.Point
	actionDescription ActionName
	explosionTiles    map[geometry.Point]int
	isGory            bool
}

func NewExplosionAction(cause Cause, origin geometry.Point, radius, pushRange, damage int) *ExplosionAction {
	return &ExplosionAction{
		BaseAction:           BaseAction{cause: cause, name: ActionNameExplosion},
		origin:               origin,
		radius:               radius,
		pushRange:            pushRange,
		damageFactor:         1.0,
		actionDescription:    "Explosion",
		damageAmount:         damage,
	}
}

func (d *ExplosionAction) Execute(context ActionContext) []Action {
	d.damageAmount = int(float64(d.damageAmount) * d.damageFactor)
	state := context.GameState
	currentMap := state.GetMap()

	d.explosionTiles = currentMap.GetDijkstraMap(d.origin, d.radius, func(point geometry.Point) bool {
		return currentMap.Contains(point) && currentMap.IsTileWalkable(point)
	})

	var hitActions []Action

	for pos, _ := range d.explosionTiles {
		if currentMap.IsObjectAt(pos) && pos != d.origin {
			hitObject := currentMap.ObjectAt(pos)
			hitObject.DamageReceived(d.damageAmount)
		}

		if currentMap.IsActorAt(pos) {
			hitActor := currentMap.ActorAt(pos)
			if d.pushRange > 0 {
				// Push the actor
				pushDirection := hitActor.MapPosition().Sub(d.origin).ToCompassDirection()
				knockback := NewKnockbackAction(d.cause, hitActor, pushDirection, d.pushRange)
				hitActions = append(hitActions, knockback)
			}
			hitActions = append(hitActions, NewChangeResourceAction(d.cause, hitActor, stats.Health, float64(-d.damageAmount)))
		}

		if d.isGory && rand.Intn(2) == 0 {
			context.GameState.AddBloodDecalAt(pos)
		}
	}
	return hitActions
}
func (d *ExplosionAction) SetGory() {
	d.isGory = true
}
func (d *ExplosionAction) GetAnimations() []Animation {
	if d.isGory {
		return nil
	}
	animation := NewExplosionAnimation(d.explosionTiles, d.radius)
	return OneAnimation(animation)
}

func (d *ExplosionAction) GetDamageFactor() float64 {
	return d.damageFactor
}

func (d *ExplosionAction) SetDamageFactor(factor float64) {
	d.damageFactor = factor
}

func (d *ExplosionAction) GetDamageDealt() int {
	return d.damageAmount
}
