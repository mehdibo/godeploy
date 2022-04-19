GOCMD=go
SERVER_NAME=bin/server
CONSOLE_NAME=bin/console

all: $(SERVER_NAME) $(CONSOLE_NAME)

$(SERVER_NAME): vendor cmd/server/main.go pkg/api pkg/auth pkg/db pkg/env pkg/middleware pkg/server
	$(GOCMD) build -o $(SERVER_NAME) cmd/server/main.go

$(CONSOLE_NAME): vendor cmd/console/main.go pkg/db pkg/env
	$(GOCMD) build -o $(CONSOLE_NAME) cmd/console/main.go

vendor: go.mod go.sum
	go mod tidy
	go mod vendor

pkg/api/go-deploy.gen.go: pkg/api/go-deploy.yml
	oapi-codegen -generate 'types,server,spec' -package api -o pkg/api/go-deploy.gen.go pkg/api/go-deploy.yml

.PHONY: generate-api
generate-api: pkg/api/go-deploy.gen.go

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	$(GOCMD) test ./pkg/auth ./pkg/env ./pkg/server