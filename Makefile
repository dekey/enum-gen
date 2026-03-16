.PHONY: help format lint format-lint build init add-vendor

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
	go fix ./...

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

# Fetch the latest tag from git and increment it only when building
build: ## Build and install enumgen with dynamic versioning
	@LATEST_TAG=$$(git tag --list --sort=-v:refname | head -n 1); \
	if [ -z "$$LATEST_TAG" ]; then \
		echo "Error: No git tag found. Please create a tag (e.g., v0.0.1) before building."; \
		exit 1; \
	fi; \
	NEW_VERSION=$$(echo $$LATEST_TAG | awk -F. '{$$NF = $$NF + 1; print $$0}' OFS=.); \
	echo "Current version: $$LATEST_TAG"; \
	echo "New version:     $$NEW_VERSION"; \
	PATH=$(PWD)/build:$$PATH go build -o ~/go/bin/enumgen -ldflags="-X 'main.version=$$NEW_VERSION'" ./cmd/enumgen/main.go
