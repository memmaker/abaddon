package common

import (
	"rpg/geometry"
)

type MeleeAttackAnimation struct {
	actor       *Actor
	targetPos   geometry.Point
	origin      geometry.Point
	ticksNeeded uint64
	ticksLeft   uint64
	audioCue    string
}

func (m *MeleeAttackAnimation) AudioCue() string {
	if m.audioCue == "" {
		return "melee_attack"
	}
	return m.audioCue
}

func (m *MeleeAttackAnimation) GetActor() *Actor {
	return m.actor
}

func (m *MeleeAttackAnimation) Cancel() {
	m.actor.SetDrawPosition(m.origin.ToPointF())
}

func (m *MeleeAttackAnimation) Draw(renderer TileRenderer) {

}

func (m *MeleeAttackAnimation) IsDone() bool {
	return m.ticksLeft == 0
}

func (m *MeleeAttackAnimation) Tick() {
	if m.ticksLeft == 0 {
		return
	}
	m.ticksLeft--
	m.updateActorPos()
}

func (m *MeleeAttackAnimation) updateActorPos() {
	percentageOfTimeLeft := float64(m.ticksLeft) / float64(m.ticksNeeded)

	percentageOfCompletion := 1.0 - percentageOfTimeLeft
	if percentageOfCompletion > 0.5 {
		percentageOfCompletion = percentageOfTimeLeft
	}

	delta := m.targetPos.Sub(m.origin)
	newPos := m.origin.ToPointF().Add(delta.ToPointF().Mul(percentageOfCompletion))
	m.actor.SetDrawPosition(newPos)
}
