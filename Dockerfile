FROM golang:1.25.7-alpine AS go
RUN apk add git
WORKDIR /usr/src/app
ADD . .
RUN CGO_ENABLED=0 go build -v -ldflags="-X 'chameth.com/chameth.com/metrics.buildVersion=$(git rev-parse HEAD)'" -o /serve ./cmd/serve && mkdir /tailscale

FROM ghcr.io/greboid/dockerbase/nonroot:1.20250803.0
COPY --from=go /serve /serve
COPY --from=go --chown=65532:65532 /tailscale /tailscale
VOLUME /tailscale
ENV PORT=8080
ENTRYPOINT ["/serve"]
