##
## Step 1 - add content and build with Hugo
##

FROM debian:stretch as hugo
RUN apt-get -qq update \
	&& DEBIAN_FRONTEND=noninteractive apt-get -qq install -y --no-install-recommends python-pygments git ca-certificates asciidoc \
	&& rm -rf /var/lib/apt/lists/*

ENV HUGO_VERSION 0.53
ENV HUGO_BINARY hugo_${HUGO_VERSION}_Linux-64bit.deb

ADD https://github.com/spf13/hugo/releases/download/v${HUGO_VERSION}/${HUGO_BINARY} /tmp/hugo.deb
RUN dpkg -i /tmp/hugo.deb \
	&& rm /tmp/hugo.deb

ADD site /tmp/site
RUN hugo -b https://www.chameth.com/ -v -s /tmp/site -d /tmp/hugo && \
	cp /tmp/hugo/index.xml /tmp/hugo/feed.xml

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
