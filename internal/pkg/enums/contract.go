package enums

// StringEnum is a minimal interface for enum types used in tests
// It is implemented by all generated enum types.
type StringEnum interface {
	IsValid() bool
	String() string
}

//go:generate enumgen --name=Env
const (
	prod    = "prod"
	gcpdev  = "gcpdev"
	staging = "staging"
	demo    = "demo"
	test    = "test"
	dev     = "dev"
	lint    = "lint"
	debug   = "debug"
	ci      = "ci"
)

// role.go
//
//go:generate enumgen --name=Role
const (
	admin  = "admin"
	editor = "editor"
	viewer = "viewer"
)
