# Step 1 - add content and build with 11ty
FROM node as node
ADD . /tmp/site
ENV LANG C.UTF-8
RUN set -eux; \
    cd /tmp/site; \
    npm install; \
    npm run build; \
    rm -rf node_modules;

# Step 2 - host with SWS
FROM ghcr.io/static-web-server/static-web-server AS sws
COPY --from=node /tmp/site/_site /site
ENV SERVER_CONFIG_FILE=/sws.toml
ADD sws.toml /sws.toml
