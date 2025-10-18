# Step 1 - build server
FROM golang:1.25.3 AS go
WORKDIR /usr/src/app
ADD go.mod go.sum /usr/src/app/
ADD cmd /usr/src/app/cmd
RUN CGO_ENABLED=0 go build -v -o /serve ./cmd/serve

# Step 2 - serve
FROM ghcr.io/greboid/dockerbase/nonroot:1.20250803.0
COPY --from=go /serve /serve
ENV PORT=8080
ENTRYPOINT ["/serve"]
