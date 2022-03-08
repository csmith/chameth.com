##
## Step 1 - add content and build with Hugo
##

FROM reg.c5h.io/hugo as hugo

ADD site /tmp/site
RUN ["hugo", "-v", "-s", "/tmp/site", "-d", "/tmp/hugo"]

##
## Step 2 - compress, minify, etc
##

FROM reg.c5h.io/alpine as minify
RUN apk add --no-cache tidyhtml libwebp-tools brotli

COPY --from=hugo --chown=65532:65532 /tmp/hugo /tmp/site
COPY --from=hugo --chown=65532:65532 /tmp/hugo/index.xml /tmp/site/feed.xml

USER 65532:65532
RUN set -eux; \
    find /tmp/site/ -name '*.html' -print -exec tidy -q -i -w 120 -m --vertical-space yes --drop-empty-elements no "{}" \;; \
    find /tmp/site/ \( -name '*.jpg' -o -name '*.png' -o -name '*.jpeg' \) -not -name 'favicon*' -exec cwebp -m 6 -mt -o "{}.webp" -- "{}" \;; \
    find /tmp/site/ -name 'favicon*.png' -exec cwebp -z 9 -mt -o "{}.webp" -- "{}" \;; \
    find /tmp/site/ \( -name '*.html' -o -name '*.css' -o -name '*.xml' \) -exec brotli -kq 11 "{}" \; -exec gzip -k9 "{}" \;;

##
## Step 3 - host!
##

FROM nginx:mainline-alpine AS nginx
COPY --from=minify /tmp/site /usr/share/nginx/html
ADD nginx.conf /etc/nginx/nginx.conf
VOLUME /logs
