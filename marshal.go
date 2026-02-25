package toon

import (
	"io"
	"reflect"
	"strconv"
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
