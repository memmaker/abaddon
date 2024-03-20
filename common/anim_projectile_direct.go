package common

import (
	"image/color"

	"rpg/geometry"
	"rpg/util"
)

type ProjectileDirectionAnimation struct {
	firstTile  int32
	frameCount int

	ticksAlive  uint64
	ticksNeeded uint64

	start    geometry.Point
	end      geometry.Point
	attacker *Actor
	onCancel func()
	onFinish func()
}

func (m *ProjectileDirectionAnimation) AudioCue() string {
	return "ranged_attack"
}

func (m *ProjectileDirectionAnimation) Cancel() {
	if m.onCancel != nil {
		m.onCancel()
	}
	if m.onFinish != nil {
		m.onFinish()
	}
}

func NewProjectileDirectionAnimation(attacker *Actor, start, end geometry.Point, secondsPerTile float64, startFrame int32, frameCount int) *ProjectileDirectionAnimation {
	ticksPerTile := util.SecondsToTicks(secondsPerTile)
	dist := geometry.Distance(start, end)
	return &ProjectileDirectionAnimation{
		start:       start,
		end:         end,
		firstTile:   startFrame,
		frameCount:  frameCount,
		ticksNeeded: uint64(float64(ticksPerTile) * dist),
		attacker:    attacker,
	}
}

func (m *ProjectileDirectionAnimation) GetActor() *Actor {
	return m.attacker
}
func (m *ProjectileDirectionAnimation) IsDone() bool {
	return m.ticksAlive >= m.ticksNeeded
}
func (m *ProjectileDirectionAnimation) Tick() {
	if m.IsDone() {
		return
	}
	m.ticksAlive++
	if m.IsDone() {
		if m.onFinish != nil {
			m.onFinish()
		}
	}
}
func (m *ProjectileDirectionAnimation) Draw(renderer TileRenderer) {
	if m.IsDone() {
		return
	}
	icon := m.firstTile
	if m.frameCount > 1 {
		icon += util.GetLoopingFrameFromTick(m.ticksAlive, 0.1, m.frameCount)
	}
	percentage := float64(m.ticksAlive) / float64(m.ticksNeeded)
	lerpedPos := util.LerpPointF(m.start.ToPointF(), m.end.ToPointF(), percentage)
	renderer.DrawOnMapF(lerpedPos, icon, color.White)
}

func (m *ProjectileDirectionAnimation) SetIcon(icon int32) {
	m.firstTile = icon
}

func (m *ProjectileDirectionAnimation) SetOnCancel(onCancel func()) {
	m.onCancel = onCancel
}

func (m *ProjectileDirectionAnimation) SetOnFinish(onFinish func()) {
	m.onFinish = onFinish
}
