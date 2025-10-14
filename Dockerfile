# Step 1 - add content and build with 11ty
FROM oven/bun:1.3.0 AS bun
ADD . /tmp/site
ENV LANG=C.UTF-8
RUN set -eux; \
    cd /tmp/site; \
    bun install; \
    bun run build; \
    rm -rf node_modules;

# Step 2 - build server
FROM golang:1.25.3 AS go
WORKDIR /usr/src/app
ADD go.mod go.sum /usr/src/app/
ADD cmd /usr/src/app/cmd
RUN CGO_ENABLED=0 go build -v -o /serve ./cmd/serve

# Step 3 - combine
FROM ghcr.io/greboid/dockerbase/nonroot:1.20250803.0
COPY --from=bun /tmp/site/_site /site
COPY --from=go /serve /serve
ENV PORT=8080 \
    FILES=/site
ENTRYPOINT ["/serve"]
