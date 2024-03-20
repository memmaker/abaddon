package gui

import (
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	"rpg/common"
	"rpg/defs"
	"rpg/geometry"
	"rpg/gridmap"
	"rpg/renderer"
	"rpg/stats"
	"rpg/util"
)

func (u *UIController) Draw(screen *ebiten.Image) {
	u.tileRenderer.SetRenderTarget(screen)

	u.mapRenderer.Draw(u.worldTicks)

	u.drawActors()

	u.context.Animator().Draw(u.mapRenderer) // should be some "animator"

	if !u.IsMouseCapturedByUI() {
		if u.targeter != nil {
			u.targeter.Draw(u.mapRenderer)
		} else {
			// default map selection rectangle
			u.mapRenderer.DrawOnMap(u.mousePosOnMap, defs.IconTargeting, color.White)
		}
	}

	u.drawUI()

	if u.message.ticksLeft > 0 {

		//u.tileRenderer.DrawTTFOnScreen(1, height+2, common.RemoveColorCodes(u.message.GetText()), color.Black)
		u.tileRenderer.DrawTTFOnScreenWithColorCodes(1, u.message.GetHeight(), u.message.GetColorcodedText())
		u.message.ticksLeft--
	}

	if u.currentTooltip != nil && !u.currentTooltip.IsNull() && u.ticksUntilTooltipAppears == 0 {
		u.currentTooltip.Draw()
	}
}

func (u *UIController) drawActors() {
	visibleMapOnScreen := u.mapRenderer.GetVisibleMap()
	state := u.context.GameState()
	player := state.GetPlayer()
	playerFieldOfView := player.GetFOV()

	enemyColor := defs.White
	actorAtlas := u.GetCurrentAtlas()

	for _, downedActor := range state.GetMap().DownedActors() {
		if !visibleMapOnScreen.Contains(downedActor.MapPosition()) || !playerFieldOfView.Visible(downedActor.MapPosition()) {
			continue
		}
		if !u.configuration.DrawTilesBehindActors && state.GetMap().IsActorAt(downedActor.MapPosition()) {
			continue
		}
		enemyColor = defs.OffWhite
		screenPos := u.mapRenderer.MapFloatToScreen(downedActor.DrawPosition())
		u.tileRenderer.DrawDefaultScaleTile(float64(screenPos.X), float64(screenPos.Y), actorAtlas, downedActor.Icon(u.worldTicks), u.context.GameState().ApplyLighting(downedActor.MapPosition(), enemyColor, u.worldTicks))
	}

	for _, actor := range state.GetMap().Actors() {
		if !visibleMapOnScreen.Contains(actor.MapPosition()) || !playerFieldOfView.Visible(actor.MapPosition()) {
			continue
		}
		if !actor.IsVisible() && !player.HasStatusEffect(common.StatusSeeInvisible) {
			continue
		}

		if u.renderingMode != GraphicRendering {
			enemyColor = actor.GetColor()
		}
		var drawColor color.Color
		drawColor = state.ApplyLighting(actor.MapPosition(), enemyColor, u.worldTicks)

		if actor.IsHurt() {
			percentage := actor.GetHurtTicksAsPercentage() // goes from 1.0 to 0.0
			drawColor = util.LerpColor(enemyColor, defs.Red, percentage)
		}
		if u.configuration.ActorBobbingEnabled {
			u.drawActorWithBobbingAnimation(actor, actorAtlas, drawColor)
		} else {
			screenPos := u.mapRenderer.MapFloatToScreen(actor.DrawPosition())
			flippedX := actor.FlipX() && u.configuration.FlipActorsHorizontally
			icon := u.getActorIcon(actor, u.worldTicks)
			u.tileRenderer.DrawDefaultScaleTileWithFlip(float64(screenPos.X), float64(screenPos.Y), actorAtlas, icon, drawColor, flippedX)
		}
	}

	for _, actor := range state.GetMap().Actors() {
		if !visibleMapOnScreen.Contains(actor.MapPosition()) || !playerFieldOfView.Visible(actor.MapPosition()) {
			continue
		}
		if !actor.GetStats().GetResource(stats.Health).IsAtMax() {
			u.drawHealthBar(actor)
		}
	}
}

func (u *UIController) drawActorWithBobbingAnimation(actor *common.Actor, atlas renderer.TextureAtlas, drawColor color.Color) {
	screenPos := u.mapRenderer.MapFloatToScreen(actor.DrawPosition())
	actorTickOffset := actor.MapPosition().X + actor.MapPosition().Y + len(actor.Name())
	tileSize := u.mapRenderer.GetScaledTileSize()
	ticksForAnimation := u.worldTicks + uint64(actorTickOffset)
	percentage := util.GetPercentageFromTick(ticksForAnimation, 1) // -> [0.0..1.0)  but we want [0.0..1.0..0.0]
	percentage *= 2
	if percentage > 1 {
		percentage = 1 - percentage
	}
	//u.tileRenderer.DrawDefaultScaleTile(float64(screenPos.X), float64(screenPos.Y), u.worldTexture, actor.Icon(u.worldTicks), color.White)
	scaleUp := geometry.PointF{X: 1, Y: 1}
	scaleDown := geometry.PointF{X: 1.05, Y: 0.95}
	currentScale := util.LerpPointF(scaleUp, scaleDown, percentage)
	yOffset := float64(tileSize.Y) * (1 - currentScale.Y)
	xOffset := (float64(tileSize.X) * (1 - currentScale.X)) / 2

	icon := u.getActorIcon(actor, ticksForAnimation)

	flippedX := actor.FlipX() && u.configuration.FlipActorsHorizontally
	u.tileRenderer.DrawScaledTileWithFlip(float64(screenPos.X)+xOffset, float64(screenPos.Y)+yOffset, atlas, icon, currentScale, drawColor, flippedX)
}

func (u *UIController) getActorIcon(actor *common.Actor, ticksForAnimation uint64) int32 {
	var icon int32
	switch u.renderingMode {
	case AsciiRendering:
		icon = actor.AsciiIcon(ticksForAnimation)
	case CodePage437Rendering:
		icon = actor.CodePage437Icon(ticksForAnimation)
	case UnicodeRendering:
		icon = actor.UnicodeIcon(ticksForAnimation)
	default:
		icon = actor.Icon(ticksForAnimation)
	}
	return icon
}

func (u *UIController) getTileIcon(tileAt gridmap.Tile) (int32, color.RGBA) {
	tileColor := defs.White
	var icon int32
	switch u.renderingMode {
	case UnicodeRendering:
		fallthrough
	case AsciiRendering:
		icon, tileColor = AsciiIconAndColor(tileAt)
	case CodePage437Rendering:
		icon, tileColor = CodePage437IconAndColor(tileAt)
	default:
		icon = tileAt.Icon()
	}

	return icon, tileColor
}

type RenderingMode int

const (
	GraphicRendering RenderingMode = iota
	AsciiRendering
	CodePage437Rendering
	UnicodeRendering
)

func AsciiIconAndColor(t gridmap.Tile) (int32, color.RGBA) {
	switch t.Special {
	case gridmap.SpecialTileWater:
		return int32('~'), defs.Blue
	case gridmap.SpecialTileForest:
		return int32(';'), defs.Green
	case gridmap.SpecialTileMountain:
		return int32('^'), defs.Brown
	case gridmap.SpecialTileVoid:
		return int32(' '), defs.Black
	case gridmap.SpecialTileStairsUp:
		return int32('<'), defs.White
	case gridmap.SpecialTileStairsDown:
		return int32('>'), defs.White
	case gridmap.SpecialTileLava:
		return int32('~'), defs.Red
	case gridmap.SpecialTileSpikes:
		return int32('^'), defs.LightGray
	}

	if t.IsWalkable {
		return '.', defs.DarkGray
	} else {
		return '#', defs.OffWhite
	}

	//return '?', defs.White
}

func CodePage437IconAndColor(t gridmap.Tile) (int32, color.RGBA) {
	//*
	//if tileAt.IsWalkable {
	//			icon = int32(util.UnicodeToCP437Byte('∙'))
	//		} else {
	//			icon = int32(util.UnicodeToCP437Byte('█'))
	//		}*/
	switch t.Special {
	case gridmap.SpecialTileWater:
		return int32(util.UnicodeToCP437Byte('≈')), defs.Blue
	case gridmap.SpecialTileForest:
		return int32(util.UnicodeToCP437Byte('♣')), defs.Green
	case gridmap.SpecialTileMountain:
		return int32(util.UnicodeToCP437Byte('∩')), defs.Brown
	case gridmap.SpecialTileVoid:
		return int32(util.UnicodeToCP437Byte(' ')), defs.Black
	case gridmap.SpecialTileStairsUp:
		return int32(util.UnicodeToCP437Byte('≡')), defs.White
	case gridmap.SpecialTileStairsDown:
		return int32(util.UnicodeToCP437Byte('≡')), defs.Cyan
	case gridmap.SpecialTileLava:
		return int32(util.UnicodeToCP437Byte('≈')), defs.Red
	case gridmap.SpecialTileSpikes:
		return int32(util.UnicodeToCP437Byte('^')), defs.LightGray
	}

	if t.IsWalkable {
		return int32(util.UnicodeToCP437Byte('∙')), defs.DarkGray
	} else {
		return int32(util.UnicodeToCP437Byte('█')), defs.OffWhite
	}
}
