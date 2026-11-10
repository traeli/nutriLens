#!/bin/sh
set -eu
SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
TARGET=${1:-"/Users/Zhuanz1/develop/src/nutriLens/miniapp/src/pages/route-detail/route-detail.vue"}
cp "$SCRIPT_DIR/route-detail.vue.original" "$TARGET"
printf 'restored %s\n' "$TARGET"
