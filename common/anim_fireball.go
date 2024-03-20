package common

import (
	"rpg/defs"
	"rpg/geometry"
)

type FireballAnimation struct {
	projectileAnimation  *ProjectileDirectionAnimation
	hitAnimations        []Animation
	caster               *Actor
	flightIcon           int32
	flightIconFrameCount int
}

func (m *FireballAnimation) GetActor() *Actor {
	return m.caster
}
func (m *FireballAnimation) Cancel() {

}

func NewFireballAnimation(caster *Actor, destination geometry.Point, flightIcon int32, flightIconFrameCount int, hitPositions []geometry.Point) *FireballAnimation {
	hitIcon := int32(defs.IconEffectFireBig)
	durationOfHit := 0.4
	secondsPerTile := 0.02
	targetPos := destination

	projectileAnimation := NewProjectileDirectionAnimation(caster, caster.MapPosition(), targetPos, secondsPerTile, flightIcon, 2)

	var hitAnimations []Animation
	for _, pos := range hitPositions {
		hitAnimations = append(hitAnimations, NewHitAnimation(pos, durationOfHit, hitIcon, 2))
	}

	return &FireballAnimation{
		projectileAnimation: projectileAnimation,
		hitAnimations:       hitAnimations,
		caster:              caster,
	}
}

func (m *FireballAnimation) IsDone() bool {
	for _, hitAnim := range m.hitAnimations {
		if !hitAnim.IsDone() {
			return false
		}
	}
	if m.projectileAnimation != nil {
		return m.projectileAnimation.IsDone()
	}
	return true
}
func (m *FireballAnimation) Tick() {
	if m.IsDone() {
		return
	}
	if m.projectileAnimation != nil {
		m.projectileAnimation.Tick()
	}
	for _, hitAnim := range m.hitAnimations {
		hitAnim.Tick()
	}
}
func (m *FireballAnimation) Draw(renderer TileRenderer) {
	if m.IsDone() {
		return
	}
	if m.projectileAnimation != nil {
		m.projectileAnimation.Draw(renderer)
	}

	for _, hitAnim := range m.hitAnimations {
		hitAnim.Draw(renderer)
	}
}

func (m *FireballAnimation) AudioCue() string {
	return "fireball"
}