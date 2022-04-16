GOCMD=go

.PHONY: generate-api
generate-api:
	oapi-codegen -generate 'types,server,spec' pkg/api/go-deploy.yml > pkg/api/go-deploy.gen.go

.PHONY: lint
lint:
	golangci-lint run
