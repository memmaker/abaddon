//go:build !js

package game

import "github.com/flopp/go-findfont"

func listFonts() []string { return findfont.List() }
