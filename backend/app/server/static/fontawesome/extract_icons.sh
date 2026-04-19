#!/bin/bash

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CSS_FILE="$SCRIPT_DIR/css/all.css"
OUT_FILE="$SCRIPT_DIR/icon_map.json"

# Extract .fa-{name} classes that have --fa: on the next line (real icon defs, not utility classes)
grep -B1 -- '--fa:' "$CSS_FILE" \
  | grep -oP '(?<=^\.fa-)[a-zA-Z0-9-]+' \
  | sort -u \
  | jq -R . \
  | jq -s . > "$OUT_FILE"

echo "Extracted $(jq length "$OUT_FILE") icons to $OUT_FILE"
