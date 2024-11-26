#!/bin/sh

set -eux

# To avoid interfering with the host's node_modules, copy everything
# to a temporary dir and then blat the node stuff.
SITE_COPY=$(mktemp -d)
cp -a . "$SITE_COPY"
rm -rf "$SITE_COPY/node_modules"

# First build the site as the current user, and fiddle with the
# permissions to allow global write access (yuck). There's no reason
# for this to use puppeteer, but it saves pulling a different image.
docker run --rm \
  -v "$SITE_COPY:/site" \
  --entrypoint /bin/sh \
  --user "$(id -u)" \
  ghcr.io/puppeteer/puppeteer:latest \
  -c '
    set -eux
    cd /site
    npm ci
    npm run build
    chmod -R o+w /site
    '

# Run puppeteer to take the screenshot. This runs as the puppeteer user
# as that's where the puppeteer stuff pre-installed.
docker run --rm \
  -v "$SITE_COPY:/site" \
  -i \
  --cap-add=SYS_ADMIN \
  --init \
  ghcr.io/puppeteer/puppeteer:latest \
  node -e "$(cat screenshot.js)" file:///site/_site/index.html /site/static/screenshot.png

# Resize the screenshot down and move it into the real location.
magick "$SITE_COPY/static/screenshot.png" -scale 640x360 "./static/screenshot.png"

# Clean up.
rm -rf "$SITE_COPY"
