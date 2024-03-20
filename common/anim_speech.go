package common

import (
	"image/color"

	"rpg/geometry"
	"rpg/util"
	"strings"
)

type SpeechAnimation struct {
	textToSpeak      []string
	textColor        color.Color
	runTimeInSeconds float64
	ticksAlive       uint64
	ticksNeeded      uint64
	actor            *Actor
}

func (m *SpeechAnimation) AudioCue() string {
	return ""
}

func (m *SpeechAnimation) GetActor() *Actor {
	return nil
}

func (m *SpeechAnimation) Cancel() {
	m.ticksAlive = m.ticksNeeded
}
func NewSpeechAnimation(actor *Actor, textToSpeak []string) *SpeechAnimation {
	timeInSeconds := 4.0
	ticksNeeded := util.SecondsToTicks(timeInSeconds)

	return &SpeechAnimation{
		actor:            actor,
		textToSpeak:      textToSpeak,
		runTimeInSeconds: timeInSeconds,
		ticksNeeded:      ticksNeeded,
		textColor:        color.White,
	}
}
func (m *SpeechAnimation) IsDone() bool {
	return m.ticksAlive >= m.ticksNeeded
}

func (m *SpeechAnimation) Tick() {
	if m.IsDone() {
		return
	}
	m.ticksAlive++
}
func (m *SpeechAnimation) Draw(renderer TileRenderer) {
	if m.IsDone() {
		return
	}
	originPos := m.actor.MapPosition()
	textString := strings.Join(m.textToSpeak, "\n")
	percentDone := float64(m.ticksAlive) / float64(m.ticksNeeded)
	percentDone = util.EaseInExpo(percentDone)

	tWidth, _ := renderer.MeasureString(textString)
	// we want to center that string on a 16 pixel wide tile
	xOffset := (renderer.GetScaledTileSize().X - float64(tWidth)) / 2.0
	screenOffset := geometry.Point{X: int(xOffset), Y: -(4 + int(float64(20)))}

	renderer.DrawStringOnMapWithOffset(originPos, screenOffset.Add(geometry.Point{Y: 2}), textString, color.Black)
	renderer.DrawStringOnMapWithOffset(originPos, screenOffset, textString, m.textColor)
}
