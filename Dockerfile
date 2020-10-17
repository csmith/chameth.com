##
## Step 1 - add content and build with Hugo
##

FROM csmith/hugo as hugo

ADD site /tmp/site
RUN hugo -v -s /tmp/site -d /tmp/hugo && \
	cp /tmp/hugo/post/index.xml /tmp/hugo/feed.xml && \
	cp /tmp/hugo/post/index.xml /tmp/hugo/index.xml

##
## Step 2 - compress, minify, etc
##

FROM debian:stretch as minify
RUN apt-get update \
    && DEBIAN_FRONTEND=noninteractive apt-get -qq install -y --no-install-recommends yui-compressor tidy webp \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=hugo /tmp/hugo /tmp/site
ADD minify.sh /tmp/minify.sh

RUN chown -R nobody:nogroup /tmp/site && chmod +x /tmp/minify.sh
USER nobody:nogroup
RUN /tmp/minify.sh

##
## Step 3 - host!
##

FROM nginx:mainline-alpine AS nginx
COPY --from=minify /tmp/site /usr/share/nginx/html
ADD nginx.conf /etc/nginx/nginx.conf
VOLUME /logs
