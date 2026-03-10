# enumgen

`enumgen` is a type-safe enum generator for Go. It generates robust, type-safe wrappers around string constants, providing validation, string conversion, and helper methods.

## Features

- **Type Safety**: Generates custom types for enums to prevent accidental misuse.
- **Validation**: Includes an `IsValid()` method to check if a value is a valid enum member.
- **String Support**: Implements the `fmt.Stringer` interface.
- **Parsing**: Provides a `FromString()` method to safely convert strings to enum types.
- **Auto-testing**: Automatically generates unit tests for the generated enums.
- **Easy Integration**: Designed to work seamlessly with `go generate`.

## Installation

```bash
go install github.com/dekey/enums/cmd/enumgen@latest
```

## Usage

1. Define your constants in a Go file.
2. Add the `//go:generate enumgen --name=<Name>` directive above the `const` block.

### Example

In `internal/examples/enums/contract.go`:

```go
package enums

//go:generate enumgen --name=Role
const (
	admin  = "admin"
	editor = "editor"
	viewer = "viewer"
)
```

Run the generator:

```bash
go generate ./...
```

### Generated Code API

The generator will create `enum_role_gen.go` with the following API:

```go
type RoleType string

var RoleTypes struct {
    Admin() RoleType
    Editor() RoleType
    Viewer() RoleType
    FromString(string) (RoleType, error)
}

func (e RoleType) IsValid() bool
func (e RoleType) String() string
```

### Usage in Code

```go
role, err := enums.RoleTypes.FromString("admin")
if err != nil {
    // handle error
}

if role == enums.RoleTypes.Admin() {
    fmt.Println("User is an admin")
}
```

## Development

The project includes a `Makefile` for common tasks:

- `make build`: Build the `enumgen` binary.
- `make test`: Run tests.
- `make lint`: Run the linter.
- `make format`: Format the code.

## License

MIT (See [LICENSE](LICENSE))