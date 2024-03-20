package common

import (
	"image/color"
	"rpg/geometry"
	"rpg/util"
)

type HitAnimation struct {
	BaseAnimation
	targetPos geometry.Point

	firstTile  int32
	frameCount int

	runTimeInSeconds float64
	ticksAlive       uint64
	ticksNeeded      uint64
	onFinished       func()
}

func (m *HitAnimation) AudioCue() string {
	return ""
}

func NewHitAnimation(targetPos geometry.Point, runtimeInSeconds float64, startFrame int32, frameCount int) *HitAnimation {
	ticksNeeded := util.SecondsToTicks(runtimeInSeconds)
	return &HitAnimation{
		targetPos:        targetPos,
		firstTile:        startFrame,
		frameCount:       frameCount,
		runTimeInSeconds: runtimeInSeconds,
		ticksNeeded:      ticksNeeded,
	}
}
func (m *HitAnimation) IsDone() bool {
	return m.ticksAlive >= m.ticksNeeded
}

func (m *HitAnimation) Draw(renderer TileRenderer) {
	if !m.shouldRunNow() || m.IsDone() {
		return
	}
	icon := m.firstTile + util.GetLoopingFrameFromTick(m.ticksAlive, m.runTimeInSeconds/float64(m.frameCount), m.frameCount)
	renderer.DrawOnMap(m.targetPos, icon, color.White)
}
func (m *HitAnimation) Tick() {
	if !m.shouldRunNow() || m.IsDone() {
		return
	}
	m.ticksAlive++
	if m.IsDone() && m.onFinished != nil {
		m.onFinished()
	}
}

func (m *HitAnimation) GetActor() *Actor {
	return nil
}

func (m *HitAnimation) Cancel() {

}

func (m *HitAnimation) SetOnFinished(onFinished func()) {
	m.onFinished = onFinished
}
