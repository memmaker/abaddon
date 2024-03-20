package common

import (
	"rpg/geometry"
)

type Object interface {
	MapPosition() geometry.Point
	Icon(tick uint64) int32
	SetMapPosition(geometry.Point)
	IsWalkable(person *Actor) bool
	IsTransparent() bool
	IsPassableForProjectile() bool
	Turn()
	ActorWalkedOn(actor *Actor, startAnimating func() bool)
	ActorBumpedInto(actor *Actor, startAnimating func() bool)
	Name() string
	DamageReceived(amount int)
	IsConductor() bool
}
