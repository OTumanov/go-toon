package toon

import (
	"reflect"
	"strconv"
)

// Unmarshal parses TOON data into v (must be pointer to struct or slice)
func Unmarshal(data []byte, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrInvalidTarget
	}

	d := newDecoder(data)
	h, err := d.parseHeader()
	if err != nil {
		return err
	}

	return d.decodeValue(h, rv.Elem())
}

// decoder scans TOON v3.0 byte stream
type decoder struct {
	data []byte
	pos  int
}

func newDecoder(data []byte) *decoder {
	return &decoder{data: data}
}

// next returns next byte and advances position
func (d *decoder) next() (byte, bool) {
	if d.pos >= len(d.data) {
		return 0, false
	}
	b := d.data[d.pos]
	d.pos++
	return b, true
}

// peek returns next byte without advancing
func (d *decoder) peek() (byte, bool) {
	if d.pos >= len(d.data) {
		return 0, false
	}
	return d.data[d.pos], true
}

// skipWhitespace advances past spaces, tabs, newlines
func (d *decoder) skipWhitespace() {
	for {
		b, ok := d.peek()
		if !ok || (b != ' ' && b != '\t' && b != '\n' && b != '\r') {
			return
		}
		d.pos++
	}
}

// header represents parsed TOON header: name[size]{field1,field2}:
type header struct {
	name   string
	size   int      // -1 if no size specified
	fields []string // nil if no fields specified
}

// parseHeader extracts header info from current position
func (d *decoder) parseHeader() (*header, error) {
	h := &header{size: -1}
	start := d.pos

	for {
		b, ok := d.next()
		if !ok {
			return nil, ErrMalformedTOON
		}

		switch b {
		case SizeStart:
			if d.pos-1 > start {
				h.name = string(d.data[start : d.pos-1])
			}
			size, err := d.parseSize()
			if err != nil {
				return nil, err
			}
			h.size = size

		case BlockStart:
			if h.name == "" && d.pos-1 > start {
				h.name = string(d.data[start : d.pos-1])
			}
			fields, err := d.parseFields()
			if err != nil {
				return nil, err
			}
			h.fields = fields

		case HeaderEnd:
			if h.name == "" && d.pos-1 > start {
				h.name = string(d.data[start : d.pos-1])
			}
			return h, nil

		default:
			// Continue scanning name
		}
	}
}

// parseSize reads number inside [ ]
func (d *decoder) parseSize() (int, error) {
	start := d.pos
	for {
		b, ok := d.next()
		if !ok {
			return 0, ErrMalformedTOON
		}
		if b == SizeEnd {
			break
		}
		if b < '0' || b > '9' {
			return 0, ErrMalformedTOON
		}
	}

	size := 0
	for i := start; i < d.pos-1; i++ {
		size = size*10 + int(d.data[i]-'0')
	}
	return size, nil
}

// parseFields reads field names inside { }
func (d *decoder) parseFields() ([]string, error) {
	var fields []string
	start := d.pos

	for {
		b, ok := d.next()
		if !ok {
			return nil, ErrMalformedTOON
		}

		switch b {
		case Separator:
			if d.pos-1 > start {
				field := string(d.data[start : d.pos-1])
				fields = append(fields, field)
			}
			start = d.pos

		case BlockEnd:
			if d.pos-1 > start {
				field := string(d.data[start : d.pos-1])
				fields = append(fields, field)
			}
			return fields, nil

		default:
			// Continue field name
		}
	}
}

// decodeValue decodes TOON data into reflect.Value based on header
func (d *decoder) decodeValue(h *header, v reflect.Value) error {
	d.skipWhitespace()

	switch v.Kind() {
	case reflect.Struct:
		return d.decodeStruct(h, v)
	case reflect.Slice:
		return d.decodeSlice(h, v)
	default:
		return ErrInvalidTarget
	}
}

// decodeStruct decodes CSV data into struct fields
func (d *decoder) decodeStruct(h *header, v reflect.Value) error {
	info := getStructInfo(v.Type())
	fieldMap := info.fields

	// Build index mapping: header field index -> struct field index
	fieldIdx := make([]int, len(h.fields))
	for i, name := range h.fields {
		idx := -1
		for j, f := range fieldMap {
			if f.name == name {
				idx = j
				break
			}
		}
		fieldIdx[i] = idx
	}

	// Parse CSV values after header
	for _, fieldIndex := range fieldIdx {
		d.skipWhitespace()

		// Read value until separator or end
		start := d.pos
		for {
			b, ok := d.peek()
			if !ok || b == Separator || b == '\n' {
				break
			}
			d.pos++
		}

		value := d.data[start:d.pos]

		// Skip separator
		if b, ok := d.peek(); ok && b == Separator {
			d.pos++
		}

		// Skip unknown fields
		if fieldIndex < 0 || fieldIndex >= len(fieldMap) {
			continue
		}

		// Set field value directly without string conversion
		field := v.Field(fieldMap[fieldIndex].index)
		if err := setFieldBytes(field, value); err != nil {
			return err
		}
	}

	return nil
}

// decodeSlice decodes multiple CSV rows into slice
func (d *decoder) decodeSlice(h *header, v reflect.Value) error {
	elemType := v.Type().Elem()

	for {
		d.skipWhitespace()

		if _, ok := d.peek(); !ok {
			break
		}

		elem := reflect.New(elemType).Elem()
		if err := d.decodeStruct(h, elem); err != nil {
			return err
		}

		v.Set(reflect.Append(v, elem))

		// Skip newline between rows
		if b, ok := d.peek(); ok && (b == '\n' || b == '\r') {
			d.pos++
			if b == '\r' {
				if b2, ok := d.peek(); ok && b2 == '\n' {
					d.pos++
				}
			}
		}
	}

	return nil
}

// setFieldBytes converts []byte value and sets it to reflect.Value
// Avoids string allocation by parsing directly from bytes
func setFieldBytes(v reflect.Value, b []byte) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(string(b))
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := parseIntBytes(b)
		if err != nil {
			return ErrMalformedTOON
		}
		v.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := parseUintBytes(b)
		if err != nil {
			return ErrMalformedTOON
		}
		v.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(string(b), 64)
		if err != nil {
			return ErrMalformedTOON
		}
		v.SetFloat(n)
	case reflect.Bool:
		// Fast path for + and - (TOON format)
		if len(b) == 1 {
			switch b[0] {
			case '+':
				v.SetBool(true)
				return nil
			case '-':
				v.SetBool(false)
				return nil
			}
		}
		// Fallback to standard parsing
		val, err := strconv.ParseBool(string(b))
		if err != nil {
			return ErrMalformedTOON
		}
		v.SetBool(val)
	default:
		return ErrInvalidTarget
	}
	return nil
}

// parseIntBytes parses int directly from []byte without allocation
func parseIntBytes(b []byte) (int64, error) {
	if len(b) == 0 {
		return 0, ErrMalformedTOON
	}
	var neg bool
	var n int64
	i := 0
	if b[0] == '-' {
		neg = true
		i = 1
	}
	for ; i < len(b); i++ {
		c := b[i]
		if c < '0' || c > '9' {
			return 0, ErrMalformedTOON
		}
		n = n*10 + int64(c-'0')
	}
	if neg {
		n = -n
	}
	return n, nil
}

// parseUintBytes parses uint directly from []byte without allocation
func parseUintBytes(b []byte) (uint64, error) {
	if len(b) == 0 {
		return 0, ErrMalformedTOON
	}
	var n uint64
	for i := 0; i < len(b); i++ {
		c := b[i]
		if c < '0' || c > '9' {
			return 0, ErrMalformedTOON
		}
		n = n*10 + uint64(c-'0')
	}
	return n, nil
}
