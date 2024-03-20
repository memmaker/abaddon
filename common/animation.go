package common

import (
	"image/color"
	"rpg/geometry"
)

type Animation interface {
	IsDone() bool
	Tick()
	Draw(renderer TileRenderer)
	Cancel()
	AudioCue() string
}
type TileRenderer interface {
	DrawOnMap(mapPos geometry.Point, icon int32, color color.Color)
	DrawStringOnMapWithOffset(mapPos, screenOffset geometry.Point, text string, color color.Color)
	MeasureString(text string) (float64, float64)
	GetScaledTileSize() geometry.PointF
	DrawOnMapF(pos geometry.PointF, icon int32, color color.Color)
}
