package apperr

import "errors"

var (
	ErrLinkExists   = errors.New("link already exists")
	ErrLinkNotFound = errors.New("link not found")
	ErrBodyRead     = errors.New("cannot read the body")
	ErrOnlyGET      = errors.New("only GET requests are allowed")
	ErrOnlyPOST     = errors.New("only POST requests are allowed")
)
