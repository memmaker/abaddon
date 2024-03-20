package defs

import "rpg/gridmap"

var FloorTile = gridmap.Tile{
	DefinedIcon:        IconCathedralFloor,
	DefinedDescription: "floor",
	IsWalkable:         true,
	IsTransparent:      true,
}
var ChasmTile = gridmap.Tile{
	DefinedIcon:        IconChasm,
	DefinedDescription: "chasm",
	IsWalkable:         true,
	IsTransparent:      true,
	Special:            gridmap.SpecialTileChasm,
}

var LavaTile = gridmap.Tile{
	DefinedIcon:   IconLava,
	IsWalkable:    true,
	IsTransparent: true,
	Special:       gridmap.SpecialTileLava,
}

var WaterTile = gridmap.Tile{
	DefinedIcon:   IconWater,
	IsWalkable:    true,
	IsTransparent: true,
	Special:       gridmap.SpecialTileWater,
}

var SpikeTile = gridmap.Tile{
	DefinedIcon:   IconSpikes,
	IsWalkable:    true,
	IsTransparent: true,
	Special:       gridmap.SpecialTileSpikes,
}
