.PHONY: help up down format lint gen-rpc format-lint load-debug-env load-dev-env

## this variable expose environment variables into sub processes. So they will be available inside sub commands.
.EXPORT_ALL_VARIABLES:

define setup_env
	$(eval ENV_FILE := $(1).env)
	@echo " - setup env $(ENV_FILE)"
	$(eval include $(1).env)
	$(eval export sed 's/=.*//' $(1).env)
endef

all:
	@make help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

init:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0
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
	PATH=$(PWD)/bin:$$PATH go build -o /Users/dekey/go/bin/enumgen ./cmd/enumgen/main.go
