package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/hajimehoshi/ebiten/v2"
	"image/color"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"path"
	"rpg/common"
	"rpg/console"
	"rpg/defs"
	"rpg/dungen"
	"rpg/files"
	"rpg/game"
	"rpg/gui"
	"rpg/osr"
	"rpg/renderer"
	"rpg/tcontrol"
	"rpg/util"
	"strconv"
	"strings"
	"time"
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
		argCmd := strings.ToLower(os.Args[1])
		switch argCmd {
		case "loot":
			theGame, _ := createGameAndUIContext("")
			genLoot(theGame)
		case "dun_gen":
			maptest()
		case "ascii":
			runGameWithConsole(createGameAndUIContext(""))
		case "list":
			theGame, _ := createGameAndUIContext("")
			dumpGameData(theGame.GetGameData())
		case "create":
			theGame, _ := createGameAndUIContext("")
			startCreator(theGame.GetGameData())
		case "osr":
			osrSim()
		case "development":
			runGameWithEbiten(createGameAndUIContext("Dev"))
		case "validate":
			return
		}
	} else {
		showTitleScreen()
	}
}

func showTitleScreen() {
	print("Who are you? ")
	// keyboard input
	reader := bufio.NewReader(os.Stdin)
	playerName, _ := reader.ReadString('\n')

	logoStartY := 3
	logo := util.ReadAllLines(path.Join("assets","logo_7drl.txt"))
	logoHeight := len(logo)

	tScreen := tcontrol.NewScreen()
	_, h := tScreen.Size()
	tScreen.PrintCentered(logoStartY, tcell.StyleDefault, logo)

	lineAfterLogo := logoStartY + logoHeight + 1
	credits := "a 7 day roguelike by Felix Ruzzoli"
	tScreen.PrintCentered(lineAfterLogo, tcell.StyleDefault, []string{credits})

	// center prompt vertically
	promptStartY := ((h - 1) / 2) + 2
	helpMsg := "Press F1 during play for help"
	anyKey := "[Press any key to start]"
	tScreen.PrintCentered(promptStartY, tcell.StyleDefault, []string{helpMsg})
	tScreen.PrintCentered(promptStartY+2, tcell.StyleDefault, []string{anyKey})

	tScreen.Show()
	graphics := true

	tScreen.WaitForKey()

	tScreen.Clear()
	tScreen.Fini()

	if graphics {
		runGameWithEbiten(createGameAndUIContext(playerName))
	} else {
		runGameWithConsole(createGameAndUIContext(playerName))
	}

	endMessage := util.ReadAllLines(path.Join("assets", "endMessage_7drl.txt"))

	tScreen = tcontrol.NewScreen()
	tScreen.Clear()
	tScreen.PrintAt(0, 0, endMessage)
	tScreen.Show()
	tScreen.WaitForKey()
	tScreen.Clear()
	tScreen.Fini()
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
		FileSystem: files.GetFileSystem(),
		PlayerName: playerName,
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

func osrSim() {
	// Helmet / Shield
	// 0 Unarmored +1, +2, +3
	// 1 Light +1, +2, +3
	// 2 Heavy +1, +2, +3
	// 2*5 = 10
	// Armor
	// 0 Unarmored +1, +2, +3
	// 2 Light +1, +2, +3
	// 4 Heavy +1, +2, +3
	// 4+3 = 7
	// => Leather Armor AC = 2
	// => Full Light AC = 4
	// => Full Heavy AC = 8
	// => FULLY ENCHANTED HEAVY AC = 17
	// => Full Enchantment adds 3x3 = 9 to AC

	// Hit Points (1d8 as hit dice)
	// Level 1: 1-8
	// Level 5: 5-40
	// Level 10: 10-80
	// => For a MAX HP and FULLY PROTECTED Fighter it would take 80+17 = 97 damage to kill him in one hit
	// => Anything dealing less than 17 damage in one blow won't even scratch him
	fighterLevelOne := osr.NewAverageFighter("Young Conan")
	fighterLevelOne.SetArmorClass(0) // leather & shield
	fighterLevelFive := osr.NewAverageFighter("Teenage Conan")
	fighterLevelFive.SetArmorClass(4) // chain & shield
	fighterLevelFive.LevelTo(osr.GetFighterSpecs(), 5)

	fighterLevelTen := osr.NewAverageFighter("Conan")
	fighterLevelTen.SetArmorClass(8) // magical plate & shield
	fighterLevelTen.SetDamageDice(util.ParseDice("1d12+2"))
	fighterLevelTen.LevelTo(osr.GetFighterSpecs(), 10)
	// max fighter always wins vs a kobold
	// min fighter still has 20-40% chance to win the fight
	// random fighter usually has a 50-80% chance to win the fight

	kobold := osr.NewMonster(osr.MonsterSpecs{
		Name:       "Kobold",
		AC:         0, // 0 protection
		AttackDice: util.ParseDice("1d6"),
		HitDice:    util.ParseDice("1d8"),
	})

	ghoul := osr.NewMonster(osr.MonsterSpecs{
		Name:       "Ghoul",
		AC:         4, // 1 protection
		AttackDice: util.ParseDice("1d6"),
		HitDice:    util.ParseDice("5d8"),
	})

	manticore := osr.NewMonster(osr.MonsterSpecs{
		Name:       "Manticore",
		AC:         8, // 2 protection
		AttackDice: util.ParseDice("1d12+2"),
		HitDice:    util.ParseDice("10d8"),
	})

	enemies := []*osr.Character{kobold, ghoul, manticore}

	iterations := 10000
	silent := true

	table := make(map[*osr.Character]map[*osr.Character]float64)
	table[fighterLevelOne] = make(map[*osr.Character]float64)
	table[fighterLevelFive] = make(map[*osr.Character]float64)
	table[fighterLevelTen] = make(map[*osr.Character]float64)
	for _, enemy := range enemies {
		srFX := osr.SurvivalRate(fighterLevelOne, enemy, silent, iterations)
		table[fighterLevelOne][enemy] = srFX
	}

	for _, enemy := range enemies {
		srFX := osr.SurvivalRate(fighterLevelFive, enemy, silent, iterations)
		table[fighterLevelFive][enemy] = srFX
	}

	for _, enemy := range enemies {
		srFX := osr.SurvivalRate(fighterLevelTen, enemy, silent, iterations)
		table[fighterLevelTen][enemy] = srFX
	}

	for _, fighter := range []*osr.Character{fighterLevelOne, fighterLevelFive, fighterLevelTen} {
		fmt.Println(fighter.CombatInfo())
	}
	for _, enemy := range enemies {
		fmt.Println(enemy.CombatInfo())
	}
	// print as csv matrix table
	fmt.Println("/," + namesFrom(enemies))
	for _, fighter := range []*osr.Character{fighterLevelOne, fighterLevelFive, fighterLevelTen} {
		fmt.Print(fighter.GetName() + ",")
		for _, enemy := range enemies {
			fmt.Printf("%.2f,", table[fighter][enemy])
		}
		fmt.Println()
	}
}

func namesFrom(enemies []*osr.Character) string {
	names := make([]string, len(enemies))
	for i, enemy := range enemies {
		names[i] = enemy.GetName()
	}
	return strings.Join(names, ",")
}

func startCreator(data defs.DataContainer) {
	creator := NewCreator(data)
	creator.Run()
}

func genLoot(theGame *game.GameInstance) {
	if len(os.Args) < 3 {
		fmt.Println("usage: loot <dungeon_level>")
		return
	}
	dungeonLevel, _ := strconv.Atoi(os.Args[2])
	items := theGame.GenerateLoot(dungeonLevel, 10)
	for _, item := range items {
		fmt.Println(item.BandStyleName())
	}
}

func dumpGameData(gameData defs.DataContainer) {
	println("Defined Items:")
	for key, _ := range gameData.ItemDefinitions {
		println(fmt.Sprintf("  %s", key))
	}

	println("Defined Enemies:")
	for key, _ := range gameData.EnemyDefinitions {
		println(fmt.Sprintf("  %s", key))
	}
}

func maptest() {
	testSource := rand.New(rand.NewSource(42))
	dunGen := dungen.NewMegaDungeonGenerator(testSource)
	// dunGen.SetDebugSteps(true)

	for {
		generatedLayout := dunGen.Generate(64, 64)
		generatedLayout.Print()
		// wait for key
		var input string
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("-more-")
		input, _ = reader.ReadString('\n')
		if strings.ToLower(input) == "q" {
			break
		}
	}

}

func runGameWithConsole(theGame *game.GameInstance, context common.UIContext) {
	theGame.SetConsoleMode(true)
	con := console.NewConsole(context)
	theGame.SetAndInitUserInterface(con)

	terminationCode := 0
	loopConfig := util.LoopConfig{
		TargetFPS:        5,
		IdleThreshold:    5,
		CurrentTimeFunc:  func() float64 { return float64(time.Now().UnixNano()) * 1e-9 },
		ProcessInputFunc: con.HandleInput,
		UpdateFunc:       con.Update,
		RenderFunc:       con.Draw,
	}

	loop := util.CreateLoop(loopConfig, &terminationCode)
	loop()

}

// runGameWithEbiten will start the infinitely running game loop
// game will receive the update and draw calls
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

