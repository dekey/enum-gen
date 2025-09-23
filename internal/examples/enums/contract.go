package enums

//go:generate enumgen --name=Env --debug
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

//go:generate enumgen --name=Role --debug
const (
	admin  = "admin"
	editor = "editor"
	viewer = "viewer"
)
