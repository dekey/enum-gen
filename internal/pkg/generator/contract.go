package generator

import "errors"

var (
	ErrParseCodeTemplate = errors.New("parse code template")
	ErrParseTestTemplate = errors.New("parse test template")
	ErrParseBaseTemplate = errors.New("parse base template")
	ErrWriteTestFile     = errors.New("write test file")
)
