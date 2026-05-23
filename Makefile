.PHONY: db dev build verify

db:
	docker-compose up -d

dev:	build db
ifndef AGENT
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
