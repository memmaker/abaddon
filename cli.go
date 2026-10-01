//go:build !js

package main

import (
	"bufio"
	"fmt"
	"github.com/gdamore/tcell/v2"
	"github.com/inancgumus/screen"
	"math/rand"
	"os"
	"path"
	"rpg/common"
	"rpg/console"
	"rpg/defs"
	"rpg/dungen"
	"rpg/game"
	"rpg/osr"
	"rpg/util"
	"strconv"
	"strings"
	"time"
)

// runCommand handles the developer/console command line modes (not available in the browser).
func runCommand(argCmd string) {
	switch argCmd {
	case "loot":
		theGame, _ := createGameAndUIContext()
		genLoot(theGame)
	case "dun_gen":
		maptest()
	case "ascii":
		runGameWithConsole(createGameAndUIContext())
	case "list":
		theGame, _ := createGameAndUIContext()
		dumpGameData(theGame.GetGameData())
	case "create":
		theGame, _ := createGameAndUIContext()
		startCreator(theGame.GetGameData())
	case "osr":
		osrSim()
	case "7drl":
		showTitleScreen()
	case "validate":
		return
	}
}

func showTitleScreen() {
	logoFile := util.MustOpenReader(path.Join("assets", "logo_7drl.txt"))
	defer logoFile.Close()
	//reader := bufio.NewReader(logoFile)
	scanner := bufio.NewScanner(logoFile)
	logoHeight := 0
	logoWidth := 0
	var logoLines []string
	for scanner.Scan() {
		line := scanner.Text()
		logoLines = append(logoLines, line)
		logoHeight++
		logoWidth = max(logoWidth, len(line))
	}

	s, _ := tcell.NewScreen()
	s.Init()
	s.Clear()
	w, h := s.Size()

	// center logo
	logoStartY := 3
	for y := 0; y < logoHeight; y++ {
		line := logoLines[y]
		for x := 0; x < len(line); x++ {
			s.SetContent((w-logoWidth)/2+x, logoStartY+y, rune(line[x]), nil, tcell.StyleDefault)
		}
	}

	lineAfterLogo := logoStartY + logoHeight + 1
	credits := "a game by Felix Ruzzoli"
	centeredCredits := (w - len(credits)) / 2
	for x, c := range credits {
		s.SetContent(centeredCredits+x, lineAfterLogo, c, nil, tcell.StyleDefault)
	}

	// center prompt vertically
	promptStartY := ((h - 1) / 2) + 2
	prompt := "Do you want graphics? (y/n)"
	centeredPrompt := (w - len(prompt)) / 2
	for x, c := range prompt {
		s.SetContent(centeredPrompt+x, promptStartY, c, nil, tcell.StyleDefault)
	}
	s.ShowCursor(centeredPrompt+len(prompt), promptStartY)
	s.SetCursorStyle(tcell.CursorStyleBlinkingUnderline)
	s.Show()
	graphics := false
outerloop:
	for event := s.PollEvent(); event != nil; event = s.PollEvent() {
		if ev, ok := event.(*tcell.EventKey); ok {
			if ev.Key() == tcell.KeyRune {
				switch ev.Rune() {
				case 'y':
					fallthrough
				case 'Y':
					graphics = true
					break outerloop
				case 'n':
					fallthrough
				case 'N':
					graphics = false
					break outerloop
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	s.Clear()
	s.Fini()
	if graphics {
		runGameWithEbiten(createGameAndUIContext())
	} else {
		runGameWithConsole(createGameAndUIContext())
	}

	screen.Clear()
	screen.MoveTopLeft()

	messageFile := util.MustOpenReader(path.Join("assets", "endMessage_7drl.txt"))
	defer messageFile.Close()
	//reader := bufio.NewReader(logoFile)
	msgScanner := bufio.NewScanner(messageFile)
	for msgScanner.Scan() {
		line := msgScanner.Text()
		println(line)
	}
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
		TargetFPS:        30,
		IdleThreshold:    5,
		CurrentTimeFunc:  func() float64 { return float64(time.Now().UnixNano()) * 1e-9 },
		ProcessInputFunc: con.HandleInput,
		UpdateFunc:       con.Update,
		RenderFunc:       con.Draw,
	}

	loop := util.CreateLoop(loopConfig, &terminationCode)
	loop()

}
