SERVER_BIN := "./bin/abf"

run: build
	docker-compose --env-file deployments/.env -f deployments/docker-compose.yaml up -d --build --remove-orphans

down:
	docker-compose --env-file deployments/.env -f deployments/docker-compose.yaml down

test:
	go test -race ./internal/...

integration-test: run
	chmod +x ./integration_test.sh && ./integration_test.sh

start: build
	$(SERVER_BIN)

build:
	go build -v -o $(SERVER_BIN) ./cmd/server

generate:
	go generate ./...

install-lint-deps:
	(which golangci-lint > /dev/null) || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.41.1

lint: install-lint-deps
	golangci-lint run ./...