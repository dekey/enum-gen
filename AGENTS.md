# Project Overview

enum generator written in go

## Key Commands

- `make test` Runs all tests in the project.
- `make build` Build for local testing.

## Important Caveats

- Tech Stack: Go 1.26, text/template, testify/require, golangci-lint with nolintlint enabled.
- Run `go:generate enumgen` in `@examples/enums/contract.go` for regenerate `enum_*.go` files this files must be valid.

## Further documentation
