# Step 1 - add content and build with 11ty
FROM oven/bun AS node
ADD . /tmp/site
ENV LANG=C.UTF-8
RUN set -eux; \
    cd /tmp/site; \
    bun install; \
    bun run build; \
    rm -rf node_modules;

# Step 2 - host with SWS
FROM ghcr.io/static-web-server/static-web-server:2.38.1 AS sws
COPY --from=node /tmp/site/_site /site
ENV SERVER_CONFIG_FILE=/sws.toml
ADD sws.toml /sws.toml
