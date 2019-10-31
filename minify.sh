#!/bin/bash

set -uxeo pipefail

# Run tidy over all HTML
find /tmp/site/ -name '*.html' -print -execdir tidy -q -i -w 120 -m --vertical-space yes --drop-empty-elements no "{}" \;

# Convert all images to WebP
find /tmp/site \( -name '*.jpg' -o -name '*.png' -o -name '*.jpeg' \) -execdir cwebp -m 6 -mt -o "{}.webp" -- "{}" \;
