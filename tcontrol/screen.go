package tcontrol

import (
    "github.com/gdamore/tcell/v2"
    "time"
)

type TerminalScreen struct {
    tcell.Screen
}

func NewScreen() *TerminalScreen {
    s, _ := tcell.NewScreen()
    err := s.Init()
    if err != nil {
        println(err)
        return nil
    }
    s.Clear()
    return &TerminalScreen{
        Screen : s,
    }
}

func (s *TerminalScreen) PrintCentered(rowStart int, style tcell.Style, text []string) int {
    width := 0
    for y := 0; y < len(text); y++ {
        line := text[y]
        width = max(width, len(line))
    }
    w, _ := s.Screen.Size()
    xStartOffset := (w - width) / 2

    for y := 0; y < len(text); y++ {
        line := text[y]
        for x := 0; x < len(line); x++ {
            xPos := xStartOffset + x
            s.Screen.SetContent(xPos, rowStart+y, rune(line[x]), nil, style)
        }
    }
    return xStartOffset
}

func (s *TerminalScreen) WaitForKey() *tcell.EventKey {
    for {
        ev := s.Screen.PollEvent()
        if ev == nil {
            return tcell.NewEventKey(tcell.KeyRune, 'Q', tcell.ModNone)
        }
        switch ev := ev.(type) {
        case *tcell.EventKey:
            return ev
        }
        time.Sleep(32 * time.Millisecond)
    }
}

func (s *TerminalScreen) PrintAt(x int, y int, message []string) {
    for i := 0; i < len(message); i++ {
        for j := 0; j < len(message[i]); j++ {
            s.Screen.SetContent(x+j, y+i, rune(message[i][j]), nil, tcell.StyleDefault)
        }
    }
}
