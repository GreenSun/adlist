#!/bin/bash

URLS=(
  "https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts"
  "https://www.github.developerdan.com/hosts/lists/ads-and-tracking-extended.txt"
  "https://v.firebog.net/hosts/AdguardDNS.txt"
  "https://s3.amazonaws.com/lists.disconnect.me/simple_tracking.txt"
  "https://s3.amazonaws.com/lists.disconnect.me/simple_ad.txt"
  "https://raw.githubusercontent.com/crazy-max/WindowsSpyBlocker/master/data/hosts/spy.txt"
  "https://winhelp2002.mvps.org/hosts.txt"
  "https://sysctl.org/cameleon/hosts"
)

INCLUDE_DIR="include"
EXCLUDE_DIR="exclude"
RESULT_FILE="result.txt"
TIMEOUT=60
RETRY=3

TEMP_DIR=$(mktemp -d)
trap "rm -rf $TEMP_DIR" EXIT

for url in "${URLS[@]}"; do
  filename="$(echo "$url" | sed 's/^https:\/\///' | sed 's/\.txt$//' | sed 's/[^a-zA-Z0-9.-]/-/g' | sed 's/^www\.//' ).txt"
  temp_filepath="$TEMP_DIR/$filename"
  if curl -sSf --connect-timeout $TIMEOUT --retry $RETRY -o "$temp_filepath" "$url"; then
    if [[ -s "$temp_filepath" ]]; then
      mv "$temp_filepath" "$INCLUDE_DIR/$filename"
    else
      echo "File $temp_filepath for $url is empty"
    fi
  else
    echo "Downloading error $url"
  fi
done

echo "Files downloaded"

find "$INCLUDE_DIR" -type f -name "*.txt" -exec sh -c 'cat "$1" && echo' _ {} \; | sort -u > "$TEMP_DIR/include"
find "$EXCLUDE_DIR" -type f -name "*.txt" -exec awk '1' {} + | sort -u > "$TEMP_DIR/exclude"

cat "$TEMP_DIR/include" |
  sed 's/\s*#.*$//' |
  sed 's/^[[:space:]]*//; s/[[:space:]]*$//' |
  awk '!/^\s*#/ && !/^\s*$/ && !/::/' |
  awk '!/^[0-9]+\.[0-9]+\.[0-9]+\.[0-9]+ / { $0 = "0.0.0.0 " $0 } 1' |
  sort -u > "$TEMP_DIR/result"

grep -vxFf "$TEMP_DIR/exclude" "$TEMP_DIR/result" > "$TEMP_DIR/result_1"

mv "$TEMP_DIR/result_1" "$RESULT_FILE"
