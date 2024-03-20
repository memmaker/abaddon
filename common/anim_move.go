package common

import (
	"rpg/geometry"
)

type MovementAnimation struct {
	actor       *Actor
	path        []geometry.Point
	ticksNeeded uint64
	ticksAlive  uint64
	isForPlayer bool
	audioCue    string
}

func (m *MovementAnimation) Cancel() {
	m.actor.SetDrawPosition(m.GetEndPos().ToPointF())
	m.actor.SetMovingAnimation(false)
}
func (m *MovementAnimation) GetEndPos() geometry.Point {
	return m.path[len(m.path)-1]
}
func (m *MovementAnimation) finishAnimation() {
	m.Cancel()
}
func (m *MovementAnimation) AudioCue() string {
	return m.audioCue
}

func (m *MovementAnimation) Draw(renderer TileRenderer) {

}

func (m *MovementAnimation) IsDone() bool {
	return m.ticksAlive == m.ticksNeeded
}

func (m *MovementAnimation) Tick() {
	if m.IsDone() {
		return
	}

	if m.ticksAlive == 0 {
		m.actor.SetMovingAnimation(true)
	}

	m.ticksAlive++
	m.updateActorPos()

	if m.IsDone() {
		m.finishAnimation()
		return
	}
}

func (m *MovementAnimation) updateActorPos() {
	ticksPerWaypoint := float64(m.ticksNeeded) / float64(len(m.path)-1)
	currentWaypointOffset := float64(m.ticksAlive) / ticksPerWaypoint
	currentWayPoint := int(currentWaypointOffset)

	from := m.path[currentWayPoint]

	if currentWayPoint == len(m.path)-1 {
		return
	}

	to := m.path[currentWayPoint+1]

	percentageOfWaypoint := currentWaypointOffset - float64(currentWayPoint)

	delta := to.Sub(from)
	newPos := from.ToPointF().Add(delta.ToPointF().Mul(percentageOfWaypoint))
	m.actor.SetDrawPosition(newPos)
}

func (m *MovementAnimation) GetActor() *Actor {
	return m.actor
}

func (m *MovementAnimation) Merge(animation *MovementAnimation) *MovementAnimation {
	// just combine the paths
	return &MovementAnimation{
		actor:       m.actor,
		path:        append(m.path, animation.path[1:]...),
		ticksNeeded: m.ticksNeeded,
		ticksAlive:  m.ticksAlive,
	}
}

func (m *MovementAnimation) IsForThePlayer() bool {
	return m.isForPlayer
}

func (m *MovementAnimation) SetSimplePath(originPos geometry.Point, dest geometry.Point) {
	m.path = []geometry.Point{originPos, dest}
}
