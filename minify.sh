#!/bin/bash

set -uxeo pipefail

# Run tidy over all HTML 
for file in $(find /tmp/site/ -name '*.html'); do
	# Tidy exits if there are warnings, which there probably will be...
	tidy -q -i -w 120 -m --vertical-space yes --drop-empty-elements no "$file" || true
done

# Convert all images to WebP
for file in $(find /tmp/site -name '*.jpg' -o -name '*.png' -o -name '*.jpeg'); do
	cwebp -m 6 -mt -o "$file.webp" -- "$file"
done
