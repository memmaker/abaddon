package common

import (
	"rpg/defs"
	"rpg/geometry"
	"rpg/util"
)

type DamageDealingObject struct {
	*BaseTimedBeing
	position             geometry.Point
	turnsUntilRemoval    int
	damagePerTurn        int
	icon                 int32
	frameCount           int
	dealDamageAtLocation func(location geometry.Point, startAnimating func() bool)
}

func (d *DamageDealingObject) IsConductor() bool {
	return false
}

func (d *DamageDealingObject) DamageReceived(amount int) {

}

func (d *DamageDealingObject) Name() string {
	return "fire"
}

func (d *DamageDealingObject) ActorBumpedInto(actor *Actor, startAnimating func() bool) {

}

func (d *DamageDealingObject) ActorWalkedOn(actor *Actor, startAnimating func() bool) {
	d.DealDamage(startAnimating)
}

func (d *DamageDealingObject) IsAlive() bool {
	return d.turnsUntilRemoval > 0
}

func (d *DamageDealingObject) MapPosition() geometry.Point {
	return d.position
}

func (d *DamageDealingObject) Icon(tick uint64) int32 {
	return d.icon + util.GetLoopingFrameFromTick(tick, 0.2, d.frameCount)
}

func (d *DamageDealingObject) SetMapPosition(point geometry.Point) {
	d.position = point
}

func (d *DamageDealingObject) IsWalkable(person *Actor) bool {
	return true
}

func (d *DamageDealingObject) IsTransparent() bool {
	return true
}

func (d *DamageDealingObject) IsPassableForProjectile() bool {
	return true
}

func (d *DamageDealingObject) Turn() {
	d.turnsUntilRemoval--
	if d.onTurn != nil {
		d.onTurn()
	}
}

func (d *DamageDealingObject) DealDamage(startAnimating func() bool) {
	d.dealDamageAtLocation(d.position, startAnimating)
}

func (d *DamageDealingObject) SetDealDamage(dealDamageAtLocation func(location geometry.Point, startAnimating func() bool)) {
	d.dealDamageAtLocation = dealDamageAtLocation
}

/*
AddEnergy(energy int)
RemoveEnergy(used int)
IsAlive() bool
GetBaseSpeed() int
*/
func NewFire(damagePerTurn int) *DamageDealingObject {
	return &DamageDealingObject{
		BaseTimedBeing:    NewBaseTimedBeing(),
		turnsUntilRemoval: 3,
		damagePerTurn:     damagePerTurn,
		icon:              defs.IconEffectFireSmall,
		frameCount:        2,
	}
}
