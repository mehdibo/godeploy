GOCMD=go
SERVER_NAME=bin/server
CONSOLE_NAME=bin/console
CONSUMER_NAME=bin/consumer
PKG_NAME=github.com/mehdibo/godeploy
VERSION ?= "dev-version"

all: $(SERVER_NAME) $(CONSOLE_NAME)

$(SERVER_NAME): vendor cmd/server/main.go pkg/api pkg/auth pkg/db pkg/env pkg/middleware pkg/server pkg/api/go-deploy.gen.go
	$(GOCMD) build -ldflags "-X 'main.Version=$(VERSION)'" -o $(SERVER_NAME) cmd/server/main.go

$(CONSOLE_NAME): vendor cmd/console/cmd/new_user.go cmd/console/cmd/root.go pkg/db pkg/env
	$(GOCMD) build -ldflags "-X '$(PKG_NAME)/cmd/console/cmd.Version=$(VERSION)'" -o $(CONSOLE_NAME) cmd/console/main.go

$(CONSUMER_NAME): vendor cmd/consumer/main.go pkg/db pkg/env pkg/messenger pkg/deployer
	$(GOCMD) build -ldflags "-X 'main.Version=$(VERSION)'" -o $(CONSUMER_NAME) cmd/consumer/main.go

vendor: go.mod go.sum
	go mod tidy
	go mod vendor

pkg/api/go-deploy.gen.go: pkg/api/go-deploy.yml
	oapi-codegen -generate 'types,server,spec' -package api -o pkg/api/go-deploy.gen.go pkg/api/go-deploy.yml

.PHONY: generate-api
api: pkg/api/go-deploy.gen.go

.PHONY: lint
lint:
	golangci-lint run

.PHONY: test
test:
	$(GOCMD) test ./pkg/auth ./pkg/env ./pkg/server