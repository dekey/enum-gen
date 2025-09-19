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

##############
### System ###
##############
load-debug-env: ## load debug.env environment file
	$(call setup_env,debug)

load-dev-env: ## load dev.env environment file
	$(call setup_env,dev)

load-test-env: ## load test.env environment file
	$(call setup_env,test)

yaml-lint: ## Run yaml linter locally
	yamlfmt --conf ./.yamlfmt.yaml -lint .

yaml-format: ## Run yaml formatter locally
	yamlfmt --conf .yamlfmt.yaml .

##########
### GO ###
##########
format: ## Run Go formatter locally
	golines --base-formatter="goimports" -w -m 120 .
	gofumpt -w .

lint: ## Run Go linter locally (used in pre-commit hook)
	golangci-lint -c ".golangci.yml" run --allow-parallel-runners ./...

lint-fix: ## Run Go linter locally & autofix issues
	golangci-lint -c ".golangci.yml" run --fix --allow-parallel-runners ./...

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

run-example:
	GOFILE=internal/pkg/enums/contract.go go run ./cmd/enumgen --name=Env