package toon // import "github.com/OTumanov/go-toon"

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

// Marshaler is the interface implemented by types that can marshal themselves into TOON
type Marshaler interface {
	MarshalTOON() ([]byte, error)
}

// Unmarshaler is the interface implemented by types that can unmarshal themselves from TOON
type Unmarshaler interface {
	UnmarshalTOON(data []byte) error
}
