package apperr

var (
	ErrLinkExists   = "link already exists"
	ErrLinkNotFound = "link not found"
	ErrBodyRead     = "cannot read the body"
	ErrOnlyGET      = "only GET requests are allowed"
	ErrOnlyPOST     = "only POST requests are allowed"
)
