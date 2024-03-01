package common

import (
	"rpg/defs"
	"rpg/geometry"
	"rpg/util"
)

type ExplodingObject struct {
	*BaseTimedBeing
	position             geometry.Point
	icon                 int32
	frameCount           int
}

func (d *ExplodingObject) Name() string {
	return "fire"
}

func (d *ExplodingObject) ActorBumpedInto(actor *Actor, startAnimating func() bool) {
	d.explode()
}

func (d *ExplodingObject) ActorWalkedOn(actor *Actor, startAnimating func() bool) {

}

func (d *ExplodingObject) DamageReceived(damage int) {
	d.explode()
}
func (d *ExplodingObject) MapPosition() geometry.Point {
	return d.position
}

func (d *ExplodingObject) Icon(tick uint64) int32 {
	return d.icon + util.GetLoopingFrameFromTick(tick, 0.2, d.frameCount)
}

func (d *ExplodingObject) SetMapPosition(point geometry.Point) {
	d.position = point
}

func (d *ExplodingObject) IsWalkable(person *Actor) bool {
	return false
}

func (d *ExplodingObject) IsTransparent() bool {
	return true
}

func (d *ExplodingObject) IsPassableForProjectile() bool {
	return false
}

func (d *ExplodingObject) explode() {

}

/*
AddEnergy(energy int)
RemoveEnergy(used int)
IsAlive() bool
GetBaseSpeed() int
*/
func NewExplodingBarrel(explosionAction *TargetedAction) *ExplodingObject {
	return &ExplodingObject{
		BaseTimedBeing:    NewBaseTimedBeing(),
		icon:              defs.IconEffectFireSmall,
		frameCount:        2,
	}
}
