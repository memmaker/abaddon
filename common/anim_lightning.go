package common

import (
	"rpg/defs"
	"rpg/geometry"
)

type LightningAnimation struct {
	hitAnimations        []Animation
	caster               *Actor
}

func (m *LightningAnimation) AudioCue() string {
	return "chain_lightning"
}

func (m *LightningAnimation) GetActor() *Actor {
	return m.caster
}
func (m *LightningAnimation) Cancel() {

}

func NewLightningAnimation(caster *Actor, hitPositions []geometry.Point) *LightningAnimation {
	hitIcon := int32(defs.IconEffectLightning)
	durationOfHit := 0.4

	var hitAnimations []Animation
	for _, pos := range hitPositions {
		hitAnimations = append(hitAnimations, NewHitAnimation(pos, durationOfHit, hitIcon, 2))
	}

	return &LightningAnimation{
		hitAnimations:       hitAnimations,
		caster:              caster,
	}
}

func (m *LightningAnimation) IsDone() bool {
	for _, hitAnim := range m.hitAnimations {
		if !hitAnim.IsDone() {
			return false
		}
	}
	return true
}
func (m *LightningAnimation) Tick() {
	if m.IsDone() {
		return
	}
	for _, hitAnim := range m.hitAnimations {
		hitAnim.Tick()
	}
}
func (m *LightningAnimation) Draw(renderer TileRenderer) {
	if m.IsDone() {
		return
	}

	for _, hitAnim := range m.hitAnimations {
		hitAnim.Draw(renderer)
	}
}
