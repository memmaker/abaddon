# Abaddon

Felix' entry for the 7DRL Challenge 2024 (https://memmakerx.itch.io/abaddon, ranked #36 overall).
Built on the Heaven & Hell engine (Go, Ebitengine, tcell console mode).

## Source

This is the real jam source: commit 2a8ef2e ("state for 7drl release") from the `7drl` branch of
the heavenAndHell repository, recovered after it was thought lost. Only build/platform glue for the
browser was added on top; gameplay is unchanged.

## Build

Desktop: `go build -tags embedded .`
Browser (wasm): `./build_web.sh [outdir]` (default `../fx-games/site/abaddon/play`), `./publish_web.sh` to deploy.

In the browser the terminal title screen is skipped and the game starts directly.
Press F1 in game for help, F2 for the horse maneuvers.
