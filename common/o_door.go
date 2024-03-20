package common

import (
	"rpg/geometry"
)

type Door struct {
	position geometry.Point

	closedIcon int32
	openIcon   int32

	isOpen bool
}

func (d *Door) Turn() {

}
func (d *Door) Name() string {
	return "a door"
}
func (d *Door) ActorWalkedOn(actor *Actor, startAnimating func() bool) {
	d.isOpen = true
}
func (d *Door) ActorBumpedInto(actor *Actor, startAnimating func() bool) {

}

func (d *Door) IsConductor() bool {
	return false
}

func (d *Door) MapPosition() geometry.Point {
	return d.position
}

func (d *Door) Icon(tick uint64) int32 {
	if d.isOpen {
		return d.openIcon
	}
	return d.closedIcon
}
func (d *Door) DamageReceived(damage int) {
	d.isOpen = true
}
func (d *Door) SetMapPosition(point geometry.Point) {
	d.position = point
}

func (d *Door) IsWalkable(person *Actor) bool {
	return true
}

func (d *Door) IsTransparent() bool {
	return d.isOpen
}

func (d *Door) IsPassableForProjectile() bool {
	return d.isOpen
}

func (d *Door) IsClosed() bool {
	return !d.isOpen
}

func (d *Door) Open() {
	d.isOpen = true
}
func NewDoor(open, closed int32) *Door {
	return &Door{
		openIcon:   open,
		closedIcon: closed,
	}
}
