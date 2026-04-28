// Package toon implements TOON (Token-Oriented Object Notation) for Go.
//
// The package provides high-performance Marshal/Unmarshal APIs, stream
// encoding/decoding, and custom Marshaler/Unmarshaler hooks.
//
// Runtime encoding/decoding uses reflection. For the fastest path with minimal
// allocations, use code generation via cmd/toongen.
//
// go-toon is designed for compact payloads in LLM pipelines where token count
// and latency matter.
//
// Conformance progress is tracked in SPEC_COMPATIBILITY.md in this repository.
// The project is distributed under the MIT license.
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
