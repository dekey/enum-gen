.PHONY: help format lint format-lint

all:
	@make help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

init:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.1
	make add-vendor

##########
### GO ###
##########
format: ## Run Go formatter locally
	golines --base-formatter="goimports" -w -m 120 .
	gofumpt -w .

lint: ## Run Go linter locally (used in pre-commit hook)
	golangci-lint -c ".golangci.yml" run --allow-parallel-runners ./...

format-lint: ## Run Go linter locally & autofix issues
	make format
	make lint

add-vendor: ## Add vendor folder
	go mod tidy
	go mod verify
	go mod vendor

update-dependencies: ## update dependencies
	go get -u ./...
	make add-vendor

test: ## Running test on local machine
	go test -parallel=1 -count=1 ./...

coverage: ## show coverage
	go tool cover -html=cover.out

build:
	PATH=$(PWD)/build:$$PATH go build -o ~/go/bin/enumgen  -ldflags="-X 'main.version=v0.0.2'" ./cmd/enumgen/main.go
