package toon

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Encoder handles TOON encoding
type Encoder struct {
	w io.Writer
}

// Encode writes a Value to the writer in TOON format
func Encode(w io.Writer, v Value) error {
	e := &Encoder{w: w}
	return e.encode(v)
}

// EncodeToString encodes a Value to a TOON string
func EncodeToString(v Value) (string, error) {
	var buf strings.Builder
	err := Encode(&buf, v)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (e *Encoder) encode(v Value) error {
	switch v.Type() {
	case Null:
		return e.write("~")
	case Boolean:
		if v.Bool() {
			return e.write("+")
		}
		return e.write("-")
	case Number:
		return e.encodeNumber(v.Number())
	case String:
		return e.encodeString(v.String())
	case Array:
		return e.encodeArray(v.Array())
	case Object:
		return e.encodeObject(v.Object())
	default:
		return fmt.Errorf("unknown type: %s", v.Type())
	}
}

func (e *Encoder) encodeNumber(n float64) error {
	// Check if it's an integer
	if n == float64(int64(n)) {
		return e.write(strconv.FormatInt(int64(n), 10))
	}
	return e.write(strconv.FormatFloat(n, 'f', -1, 64))
}

func (e *Encoder) encodeString(s string) error {
	var buf strings.Builder
	buf.WriteByte('"')

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			if c < 0x20 {
				// Control characters - use unicode escape
				buf.WriteString(fmt.Sprintf(`\u%04x`, c))
			} else {
				buf.WriteByte(c)
			}
		}
	}

	buf.WriteByte('"')
	return e.write(buf.String())
}

func (e *Encoder) encodeArray(arr []Value) error {
	if err := e.write("["); err != nil {
		return err
	}

	for i, item := range arr {
		if i > 0 {
			if err := e.write(" "); err != nil {
				return err
			}
		}
		if err := e.encode(item); err != nil {
			return err
		}
	}

	return e.write("]")
}

func (e *Encoder) encodeObject(obj map[string]Value) error {
	if err := e.write("{"); err != nil {
		return err
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}

	// Simple sort for consistent output
	for i := 0; i < len(keys); i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}

	for i, key := range keys {
		if i > 0 {
			if err := e.write(" "); err != nil {
				return err
			}
		}

		// Write key (use identifier if possible, otherwise string)
		if isIdentifier(key) {
			if err := e.write(key); err != nil {
				return err
			}
		} else {
			if err := e.encodeString(key); err != nil {
				return err
			}
		}

		if err := e.write(" "); err != nil {
			return err
		}

		if err := e.encode(obj[key]); err != nil {
			return err
		}
	}

	return e.write("}")
}

func (e *Encoder) write(s string) error {
	_, err := e.w.Write([]byte(s))
	return err
}

func isIdentifier(s string) bool {
	if s == "" {
		return false
	}

	// First character must be a letter or underscore
	c := s[0]
	if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_') {
		return false
	}

	// Rest can be letters, digits, underscores, or hyphens
	for i := 1; i < len(s); i++ {
		c = s[i]
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-') {
			return false
		}
	}

	return true
}
