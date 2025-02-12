#!/bin/bash

INCLUDE_DIR="include"
EXCLUDE_DIR="exclude"
RESULT_FILE="result.txt"

TEMP_INCLUDE=$(mktemp)
find "$INCLUDE_DIR" -type f -name "*.txt" -exec sh -c 'cat "$1" && echo' _ {} \; | sort -u > "$TEMP_INCLUDE"

TEMP_EXCLUDE=$(mktemp)
find "$EXCLUDE_DIR" -type f -name "*.txt" -exec awk '1' {} + | sort -u > "$TEMP_EXCLUDE"

grep -vxFf "$TEMP_EXCLUDE" "$TEMP_INCLUDE" | sort -u > "$RESULT_FILE"

rm -f "$TEMP_INCLUDE" "$TEMP_EXCLUDE"