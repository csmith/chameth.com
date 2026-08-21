.PHONY: db db-down dev build verify

db:
	docker compose up -d --wait

db-down:
	docker compose down

dev:	build db
ifndef AGENT
	trap 'docker compose down' EXIT; \
	trap 'exit 130' INT TERM; \
	bash -c "export $$(grep -v '^#' .env | xargs -d '\n'); /tmp/chamethdotcom"
else
	$(error The dev target should not be run by agents)
endif

build:
	go generate ./...
	bash -c "go build -v -ldflags=\"-X 'chameth.com/chameth.com/cmd/serve/metrics.buildVersion=$$(git rev-parse HEAD)'\" -o /tmp/chamethdotcom ./cmd/serve"

verify: build
	go vet ./...
	go fix ./...
	staticcheck ./...
	go fmt ./...
