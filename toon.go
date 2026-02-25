package toon

import "errors"

var (
	ErrMalformedTOON = errors.New("toon: malformed syntax")
	ErrInvalidTarget = errors.New("toon: target must be a pointer to struct or slice")
)

// Separators per v3.0 spec
const (
	BlockStart = '{'
	BlockEnd   = '}'
	SizeStart  = '['
	SizeEnd    = ']'
	HeaderEnd  = ':'
	Separator  = ','
)
