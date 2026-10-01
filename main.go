package main

import (
	"errors"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	_ "image/png"
	"log"
	"os"
	"path"
	"rpg/common"
	"rpg/defs"
	"rpg/files"
	"rpg/game"
	"rpg/gui"
	"rpg/renderer"
	"rpg/util"
	"strings"
)

func main() {
	/*
		cpuProfileFile, err := os.Create("cpu.prof")
		if err != nil {
			panic(err)
		}
		defer cpuProfileFile.Close()

		// start CPU profiling
		if err := pprof.StartCPUProfile(cpuProfileFile); err != nil {
			panic(err)
		}
		defer pprof.StopCPUProfile()
	*/
	if len(os.Args) > 1 {
		runCommand(strings.ToLower(os.Args[1]))
	} else {
		start()
	}
}

func createGameAndUIContext(playerName string) (*game.GameInstance, common.UIContext) {
	userConfig := common.NewConfigurationFromRecFile("config.rec")
	startInfos := common.StartInfo{
		Title:       "Abaddon",
		ScreenSize:  userConfig.WindowSize,
		TileScale:   4.0,
		LDTKMapFile: "assets/maps.ldtk",
		FontInfos: common.FontInfo{
			TTFFile: userConfig.DefaultFont,
			TTFSize: userConfig.DefaultFontSize,
		},
		UserConfiguration: userConfig,
		FileSystem:        files.GetFileSystem(),
		PlayerName:        playerName,
	}
	theGame := game.NewGameInstance(startInfos, NewDummyUI())
	uiContext := common.UIContext{
		GameState:    theGame.GetWorldState,
		PlayerAction: theGame.GetActions,
		AudioPlayer:  theGame.GetAudio(),
		ResetGame:    theGame.ResetGame,
		Animator: func() common.Animator {
			return theGame.GetActionBroker()
		},
		GetLog:      theGame.GetLog,
		StartInfos:  startInfos,
		Definitions: theGame.GetGameData(),
	}
	return theGame, uiContext
}

func runGameWithEbiten(theGame *game.GameInstance, context common.UIContext) {
	startInfos := context.StartInfos
	tileRenderer := renderer.NewTileRenderer(startInfos.TileScale)
	tileRenderer.SetColorFromName(defs.ColorFromName)
	tileRenderer.SetDefaultAtlas(theGame.GetWorldState().GetWorldTexture())
	tileRenderer.SetDefaultColors(defs.DarkTextColor, color.Black)
	tileRenderer.SetWhiteTile(defs.IconWhite)

	tileMapAnimations := map[int32]gui.TilemapAnimInfo{
		28:                       {FrameCount: 2, DelayInSeconds: 1.5},
		32:                       {FrameCount: 2, DelayInSeconds: 0.2},
		defs.IconEffectFireSmall: {FrameCount: 2, DelayInSeconds: 0.2},
	}
	uiController := gui.NewUIController(tileRenderer, context, startInfos.UserConfiguration)
	uiController.SetTileMapAnimations(tileMapAnimations)
	uiController.SetTTFFont(util.MustOpen(startInfos.FontInfos.TTFFile), startInfos.FontInfos.TTFSize)
	uiController.SetAsciiAtlas(renderer.NewTextureAtlas(path.Join("assets", "Bm437_IBM_CGA.png"), 8, 8))
	theGame.SetAndInitUserInterface(uiController)

	screenSize := uiController.GetDeviceIndependentScreenSize()

	ebiten.SetWindowTitle("Abaddon")
	ebiten.SetWindowSize(screenSize.X, screenSize.Y)
	ebiten.SetScreenClearedEveryFrame(true)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSizeLimits(640, 480, -1, -1)
	ebiten.SetWindowClosingHandled(false)

	err := ebiten.RunGameWithOptions(uiController, &ebiten.RunGameOptions{
		GraphicsLibrary: ebiten.GraphicsLibraryOpenGL,
	})
	if err != nil && !errors.Is(err, ebiten.Termination) {
		log.Fatal(err)
	}
}
