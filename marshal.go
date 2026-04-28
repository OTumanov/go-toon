package toon

import (
	"io"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// buffer pool for zero-allocation encoding
var bufPool = sync.Pool{
	New: func() interface{} {
		return make([]byte, 0, 256)
	},
}

// Marshal encodes v into TOON v3.0 format
func Marshal(v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, ErrInvalidTarget
	}
	rv = rv.Elem()

	if supportsPrimitiveRoot(rv.Kind()) {
		return marshalRootPrimitive(rv)
	}

	buf := bufPool.Get().([]byte)
	buf = buf[:0]
	defer bufPool.Put(buf)

	e := &encoder{buf: buf}
	if err := e.encode(rv); err != nil {
		return nil, err
	}

	// Copy result (buf will be reused)
	result := make([]byte, len(e.buf))
	copy(result, e.buf)
	return result, nil
}

func marshalRootPrimitive(v reflect.Value) ([]byte, error) {
	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if needsQuotedStringValue(s) {
			return []byte(`"` + escapeString(s) + `"`), nil
		}
		return []byte(s), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []byte(strconv.FormatInt(v.Int(), 10)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []byte(strconv.FormatUint(v.Uint(), 10)), nil
	case reflect.Float32, reflect.Float64:
		return []byte(strconv.FormatFloat(v.Float(), 'f', -1, 64)), nil
	case reflect.Bool:
		if v.Bool() {
			return []byte("true"), nil
		}
		return []byte("false"), nil
	default:
		return nil, ErrInvalidTarget
	}
}

// marshalObjectLinesForSpec encodes struct fields into line-based object format:
// key: value
// This helper is used for incremental spec-fixture conformance work.
func marshalObjectLinesForSpec(v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if !rv.IsValid() {
		return nil, ErrInvalidTarget
	}
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, ErrInvalidTarget
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, ErrInvalidTarget
	}

	buf := make([]byte, 0, 128)
	if err := encodeSpecObjectInto(&buf, rv, 0); err != nil {
		return nil, err
	}
	return buf, nil
}

func encodeSpecObjectInto(buf *[]byte, v reflect.Value, indent int) error {
	info := getStructInfo(v.Type())
	lineCount := 0

	for _, f := range info.fields {
		field := v.Field(f.index)
		if lineCount > 0 {
			*buf = append(*buf, '\n')
		}
		lineCount++

		for i := 0; i < indent; i++ {
			*buf = append(*buf, ' ')
		}

		key := f.name
		if needsQuotedKey(key) {
			key = `"` + escapeString(key) + `"`
		}
		*buf = append(*buf, key...)
		*buf = append(*buf, ':')

		encoded, nested, err := encodeSpecValue(field)
		if err != nil {
			return err
		}
		if nested {
			if len(encoded) > 0 {
				*buf = append(*buf, '\n')
				*buf = append(*buf, encoded...)
			}
			continue
		}
		*buf = append(*buf, ' ')
		*buf = append(*buf, encoded...)
	}

	return nil
}

func encodeSpecValue(v reflect.Value) ([]byte, bool, error) {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			return []byte("null"), false, nil
		}
		return encodeSpecValue(v.Elem())
	case reflect.Struct:
		nested := make([]byte, 0, 64)
		if err := encodeSpecObjectInto(&nested, v, 1); err != nil {
			return nil, false, err
		}
		return nested, true, nil
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if needsQuotedStringValue(s) {
			return []byte(`"` + escapeString(s) + `"`), false, nil
		}
		return []byte(s), false, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return []byte(strconv.FormatInt(v.Int(), 10)), false, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return []byte(strconv.FormatUint(v.Uint(), 10)), false, nil
	case reflect.Float32, reflect.Float64:
		return []byte(strconv.FormatFloat(v.Float(), 'f', -1, 64)), false, nil
	case reflect.Bool:
		if v.Bool() {
			return []byte("true"), false, nil
		}
		return []byte("false"), false, nil
	default:
		return nil, false, ErrInvalidTarget
	}
}

func needsQuotedKey(s string) bool {
	if s == "" {
		return true
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ':', ',', '[', ']', '{', '}', '"':
			return true
		}
	}
	if strings.HasPrefix(s, "-") {
		return true
	}
	if strings.TrimSpace(s) != s {
		return true
	}
	if strings.Contains(s, " ") {
		return true
	}
	return false
}

func needsQuotedStringValue(s string) bool {
	if s == "" {
		return true
	}
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case ':', ',', '\n', '\t', '\r', '\\', '"':
			return true
		}
	}
	if strings.TrimSpace(s) != s {
		return true
	}
	if s == "true" || s == "false" || s == "null" {
		return true
	}
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	return false
}

func escapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	return s
}

// MarshalTo writes TOON encoding to w
type MarshalerTo interface {
	MarshalTOON(w io.Writer) error
}

// encoder writes TOON v3.0 format
type encoder struct {
	buf []byte
}

func (e *encoder) encode(v reflect.Value) error {
	switch v.Kind() {
	case reflect.Struct:
		return e.encodeStruct(v)
	case reflect.Slice:
		return e.encodeSlice(v)
	case reflect.Ptr:
		if v.IsNil() {
			e.buf = append(e.buf, '~')
			return nil
		}
		return e.encode(v.Elem())
	default:
		return ErrInvalidTarget
	}
}

func (e *encoder) encodeStruct(v reflect.Value) error {
	t := v.Type()
	info := getStructInfo(t)

	// Header: name[size]{fields}:
	e.buf = append(e.buf, info.name...)
	e.buf = append(e.buf, '{')
	for i, f := range info.fields {
		if i > 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = append(e.buf, f.name...)
	}
	e.buf = append(e.buf, '}', ':')

	// Values
	for i, f := range info.fields {
		if i > 0 {
			e.buf = append(e.buf, ',')
		}
		if err := e.encodeValue(v.Field(f.index)); err != nil {
			return err
		}
	}

	return nil
}

func (e *encoder) encodeSlice(v reflect.Value) error {
	if v.Len() == 0 {
		e.buf = append(e.buf, '~')
		return nil
	}

	// Encode first element to get header
	elem := v.Index(0)
	t := elem.Type()
	info := getStructInfo(t)

	// Header with size: name[size]{fields}:
	e.buf = append(e.buf, info.name...)
	e.buf = append(e.buf, '[')
	e.buf = strconv.AppendInt(e.buf, int64(v.Len()), 10)
	e.buf = append(e.buf, ']')
	e.buf = append(e.buf, '{')
	for i, f := range info.fields {
		if i > 0 {
			e.buf = append(e.buf, ',')
		}
		e.buf = append(e.buf, f.name...)
	}
	e.buf = append(e.buf, '}', ':')

	// Values for each element
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			e.buf = append(e.buf, '\n')
		}
		elem := v.Index(i)
		for j, f := range info.fields {
			if j > 0 {
				e.buf = append(e.buf, ',')
			}
			if err := e.encodeValue(elem.Field(f.index)); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *encoder) encodeValue(v reflect.Value) error {
	// Check for custom Marshaler interface
	if v.CanInterface() {
		if m, ok := v.Interface().(Marshaler); ok {
			data, err := m.MarshalTOON()
			if err != nil {
				return err
			}
			e.buf = append(e.buf, data...)
			return nil
		}
	}

	switch v.Kind() {
	case reflect.String:
		e.buf = append(e.buf, v.String()...)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.buf = strconv.AppendInt(e.buf, v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		e.buf = strconv.AppendUint(e.buf, v.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		e.buf = strconv.AppendFloat(e.buf, v.Float(), 'f', -1, 64)
	case reflect.Bool:
		if v.Bool() {
			e.buf = append(e.buf, '+')
		} else {
			e.buf = append(e.buf, '-')
		}
	case reflect.Slice:
		// Nested slice as [item1,item2]
		e.buf = append(e.buf, '[')
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				e.buf = append(e.buf, ',')
			}
			if err := e.encodeValue(v.Index(i)); err != nil {
				return err
			}
		}
		e.buf = append(e.buf, ']')
	default:
		return ErrInvalidTarget
	}
	return nil
}
