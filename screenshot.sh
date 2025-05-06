#!/bin/sh

set -eux

# To avoid interfering with the host's node_modules, copy everything
# to a temporary dir and then blat the node stuff.
SITE_COPY=$(mktemp -d --tmpdir chameth.com-screenshot.XXXXXXXX)
cp -a . "$SITE_COPY"
rm -rf "$SITE_COPY/node_modules"

# Build the site 
docker run --rm \
  -v "$SITE_COPY:/site:U" \
  --entrypoint /bin/sh \
  ghcr.io/puppeteer/puppeteer:latest \
  -c '
    set -eux
    cd /site
    npm ci
    npm run build
    '

# Run puppeteer to take the screenshot. This runs as the puppeteer user
# as that's where the puppeteer stuff pre-installed.
docker run --rm \
  -v "$SITE_COPY:/site:U" \
  -i \
  --cap-add=SYS_ADMIN \
  --init \
  ghcr.io/puppeteer/puppeteer:latest \
  node -e "$(cat screenshot.js)" file:///site/_site/index.html /site/static/screenshot.png

# Resize the screenshot down and move it into the real location.
magick "$SITE_COPY/static/screenshot.png" -scale 640x360 "./static/screenshot.png"
