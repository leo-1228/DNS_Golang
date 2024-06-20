package reader

import "errors"

var (
	ErrRangeReached = errors.New("workspace range has been reached")
	ErrDomainsEOF   = errors.New("all domains has been processed")
)
