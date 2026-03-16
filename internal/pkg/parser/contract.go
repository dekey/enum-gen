package parser

import "errors"

var (
	ErrParseLineNumber = errors.New("parse line number")
	ErrInvalidGoLine   = errors.New("goLine must be a positive integer")
)
