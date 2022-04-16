GOCMD=go

.PHONY: generate-api
generate-api:
	oapi-codegen -generate 'types,server,spec' -package api -o pkg/api/go-deploy.gen.go pkg/api/go-deploy.yml

.PHONY: lint
lint:
	golangci-lint run
