#!/bin/sh
set -eu
TARGET=${1:-"/Users/Zhuanz1/develop/src/nutriLens/server/internal/model/pgsql/place_theme.go"}
rm -f "$TARGET"
printf 'removed %s\n' "$TARGET"
