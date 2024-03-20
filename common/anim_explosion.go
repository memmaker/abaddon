package common

import (
	"math/rand"
	"rpg/defs"
	"rpg/geometry"
	"rpg/util"
)

type ExplosionAnimation struct {
	hitPositions map[geometry.Point]int
	buildUpSteps int
	currentStep  int
	ticksPerStep int
	tickCounter int
}

func (m *ExplosionAnimation) AudioCue() string {
	return "exploding_barrel"
}

func (m *ExplosionAnimation) Cancel() {

}
func (m *ExplosionAnimation) reachedDiffusionLimit() bool {
	return m.currentStep >= m.buildUpSteps
}

func NewExplosionAnimation(hitPositions map[geometry.Point]int, radius int) *ExplosionAnimation {

	// idea:
	// 1. spread the small flames first
	// 2. follow with a dist of 1-2 with big flames
	// 3. randomize the fizzling out of the flames

	return &ExplosionAnimation{
		hitPositions: hitPositions,
		buildUpSteps: radius,
		ticksPerStep: int(util.SecondsToTicks(0.1)),
	}
}

func (m *ExplosionAnimation) IsDone() bool {
	return len(m.hitPositions) == 0
}
func (m *ExplosionAnimation) Tick() {
	if m.IsDone() {
		return
	}
	if m.reachedDiffusionLimit() && rand.Intn(3) == 0 {
		m.removeRandomFlame()
		return
	}
	m.tickCounter++
	if m.tickCounter < m.ticksPerStep {
		return
	}
	m.tickCounter = 0
	m.currentStep++
}
func (m *ExplosionAnimation) Draw(renderer TileRenderer) {
	if m.IsDone() {
		return
	}

	maxSteps := m.buildUpSteps * 10
	drawIcon := int32(defs.IconEffectFireBig)
	if !m.reachedDiffusionLimit() {
		maxSteps = m.currentStep * 10
		drawIcon = int32(defs.IconEffectFireSmall)
	}
	for pos, dist := range m.hitPositions {
		if dist > maxSteps {
			continue
		}
		renderer.DrawOnMap(pos, drawIcon, defs.White)
	}
}

func (m *ExplosionAnimation) removeRandomFlame() {
	randomIndex := rand.Intn(len(m.hitPositions))
	index := 0
	for pos := range m.hitPositions {
		if index == randomIndex {
			delete(m.hitPositions, pos)
			return
		}
		index++
	}
}
