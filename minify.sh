#!/bin/bash

set -uxeo pipefail

# Compress all the CSS files together
cat /tmp/site/res/css/*.css > "/tmp/css-concatted.css"
yui-compressor "/tmp/css-concatted.css" > "/tmp/css-combined.css"

# Generate a small hash to bust caches if the file changes
HASH=`sha256sum /tmp/css-combined.css | cut -c -10`

# Replace the old CSS with the new
mv "/tmp/css-combined.css" "/tmp/site/res/stylesheet-$HASH.css"
rm -rf "/tmp/site/res/css"

# Replace the references in the HTML, then run tidy over it
for file in $(find /tmp/site/ -name '*.html'); do
	sed -i "s#\"/res/css/style.css\"#\"/res/stylesheet-$HASH.css\"#g" "$file"
	sed -i '\#"/res/css/.*.css"#d' "$file"

	# Tidy exits if there are warnings, which there probably will be...
	tidy -q -i -w 120 -m --vertical-space yes --drop-empty-elements no "$file" || true
done
