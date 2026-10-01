package main

func runCommand(string) {}

// start skips the terminal title screen in the browser (no stdin/tty) and goes straight to the game.
func start() { runGameWithEbiten(createGameAndUIContext("Player")) }
