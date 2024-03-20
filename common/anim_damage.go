package common

import (
	"image/color"
	"math"
	"rpg/defs"
	"rpg/geometry"
	"rpg/util"
	"strconv"
	"strings"
)

type DamageNumberAnimation struct {
	BaseAnimation
	targetPos        geometry.Point
	damageAmount     int
	damageColor      color.Color
	runTimeInSeconds float64
	ticksAlive       uint64
	ticksNeeded      uint64

	victimName    string
	isVictimAlive bool
	audioEnabled  bool
}

func (m *DamageNumberAnimation) AudioCue() string {
	if !m.audioEnabled {
		return ""
	}
	hitCue := "human_hit"
	if strings.HasPrefix(m.victimName, "skeleton") {
		if m.isVictimAlive {
			hitCue = "skeleton"
		} else {
			hitCue = "skeleton_death"
		}
	} else if strings.HasPrefix(m.victimName, "rat") {
		if m.isVictimAlive {
			hitCue = "rat"
		} else {
			hitCue = "rat_death"
		}
	} else if strings.HasPrefix(m.victimName, "bat") {
		if m.isVictimAlive {
			hitCue = "bat"
		} else {
			hitCue = "bat_death"
		}
	} else if strings.HasPrefix(m.victimName, "caveling") {
		if m.isVictimAlive {
			hitCue = "bear"
		} else {
			hitCue = "bear_death"
		}
	} else if strings.Contains(m.victimName, "kestrel") {
		if m.isVictimAlive {
			hitCue = "bird"
		} else {
			hitCue = "bird_death"
		}
	} else if strings.Contains(m.victimName, "snake") {
		if m.isVictimAlive {
			hitCue = "snake"
		} else {
			hitCue = "snake_death"
		}
	} else if strings.HasPrefix(m.victimName, "fallen") || strings.HasPrefix(m.victimName, "fiend") {
		if m.isVictimAlive {
			hitCue = "imp"
		} else {
			hitCue = "imp_death"
		}
	} else if m.victimName == "cerberus" {
		if !m.isVictimAlive {
			hitCue = "cerberus_death"
		} else {
			hitCue = "cerberus"
		}
	} else if m.damageAmount < -9 || !m.isVictimAlive {
		hitCue = "gore_hit"
	}
	return hitCue
}

func (m *DamageNumberAnimation) GetActor() *Actor {
	return nil
}

func (m *DamageNumberAnimation) Cancel() {

}
func NewDamageNumberAnimation(victimInternalName string, isVictimAlive bool, mapPosition geometry.Point, amount int) *DamageNumberAnimation {
	timeInSeconds := 1.0
	ticksNeeded := util.SecondsToTicks(timeInSeconds)
	var damageColor color.Color
	if amount > 0 {
		damageColor = defs.PositiveHighlightColor
	} else {
		damageColor = defs.PureRed
	}

	return &DamageNumberAnimation{
		targetPos:        mapPosition,
		damageAmount:     int(math.Abs(float64(amount))),
		runTimeInSeconds: timeInSeconds,
		ticksNeeded:      ticksNeeded,
		damageColor:      damageColor,
		victimName: victimInternalName,
		isVictimAlive: isVictimAlive,
	}
}
func (m *DamageNumberAnimation) IsDone() bool {
	return m.ticksAlive >= m.ticksNeeded
}

func (m *DamageNumberAnimation) Tick() {
	if !m.shouldRunNow() || m.IsDone() {
		return
	}
	m.ticksAlive++
}
func (m *DamageNumberAnimation) Draw(renderer TileRenderer) {
	if !m.shouldRunNow() || m.IsDone() {
		return
	}
	damageString := strconv.Itoa(m.damageAmount)
	originPos := m.targetPos

	percentDone := float64(m.ticksAlive) / float64(m.ticksNeeded)
	percentDone = util.EaseInExpo(percentDone)

	tWidth, _ := renderer.MeasureString(damageString)
	// we want to center that string on a 16 pixel wide tile
	xOffset := float64(renderer.GetScaledTileSize().X-float64(tWidth)) / 2.0
	screenOffset := geometry.Point{X: int(xOffset), Y: -(4 + int(float64(20)*percentDone))}

	renderer.DrawStringOnMapWithOffset(originPos, screenOffset.Add(geometry.Point{Y: 2}), damageString, color.Black)
	renderer.DrawStringOnMapWithOffset(originPos, screenOffset, damageString, m.damageColor)
}

func (m *DamageNumberAnimation) SetAudio(value bool) {
	m.audioEnabled = value
}
