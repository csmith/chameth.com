FROM golang:1.25.6-alpine AS go
RUN apk add git # Needed for Go to embed VCS information
WORKDIR /usr/src/app
ADD go.mod go.sum /usr/src/app/
ADD cmd /usr/src/app/cmd
RUN CGO_ENABLED=0 go build -v -o /serve ./cmd/serve && mkdir /tailscale

FROM ghcr.io/greboid/dockerbase/nonroot:1.20250803.0
COPY --from=go /serve /serve
COPY --from=go --chown=65532:65532 /tailscale /tailscale
VOLUME /tailscale
ENV PORT=8080
ENTRYPOINT ["/serve"]
