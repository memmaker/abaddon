#!/bin/sh
# Builds the browser version (assets embedded) into $1 (default: the fx-games site).
set -e
OUT=${1:-../fx-games/site/abaddon/play}
mkdir -p "$OUT"
GOOS=js GOARCH=wasm go build -trimpath -ldflags '-s -w' -tags embedded,ebitensinglethread -o "$OUT/abaddon.wasm" .
cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" "$OUT/" 2>/dev/null || cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" "$OUT/"
SIZE=$(wc -c < "$OUT/abaddon.wasm" | tr -d ' ')
sed "s/@SIZE@/$SIZE/" web/index.html > "$OUT/index.html"
rm -f "$OUT/abaddon.wasm.gz"; gzip -9 -k "$OUT/abaddon.wasm"
echo "$OUT/abaddon.wasm: $SIZE bytes"
