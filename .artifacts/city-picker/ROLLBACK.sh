#!/bin/sh
set -eu
TARGET_ROOT="${1:-/Users/Zhuanz1/develop/src/nutriLens}"
ARTIFACT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
cp "$ARTIFACT_DIR/home.vue.original" "$TARGET_ROOT/miniapp/src/pages/home/home.vue"
cp "$ARTIFACT_DIR/package.json.original" "$TARGET_ROOT/miniapp/package.json"
cp "$ARTIFACT_DIR/package-lock.json.original" "$TARGET_ROOT/miniapp/package-lock.json"
rm -f "$TARGET_ROOT/miniapp/src/utils/city-search.js" "$TARGET_ROOT/miniapp/tests/city-search.test.mjs"
printf '%s\n' 'restored home city selection to showActionSheet source state'
