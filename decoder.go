package toon

import (
	"bytes"
	"reflect"
	"strconv"
)

// Unmarshal parses TOON data into v (must be pointer to struct or slice)
func Unmarshal(data []byte, v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return ErrInvalidTarget
	}

	// Root primitive fallback for spec-compatible decode cases like:
	// "42", "true", "\"hello\"".
	if supportsPrimitiveRoot(rv.Elem().Kind()) {
		if err := unmarshalRootPrimitive(data, rv.Elem()); err == nil {
			return nil
		}
	}

	d := newDecoder(data)
	h, err := d.parseHeader()
	if err != nil {
		// Fallback: allow line-based object form ("key: value") for struct targets.
		// This keeps existing header-based fast path intact while enabling
		// incremental spec fixture compatibility work.
		if rv.Elem().Kind() == reflect.Struct {
			if ferr := unmarshalObjectLines(data, rv.Elem()); ferr == nil {
				return nil
			}
		}
		return err
	}

	return d.decodeValue(h, rv.Elem())
}

func supportsPrimitiveRoot(k reflect.Kind) bool {
	switch k {
	case reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Bool:
		return true
	default:
		return false
	}
}

func unmarshalRootPrimitive(data []byte, v reflect.Value) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 {
		return ErrMalformedTOON
	}
	if bytes.Equal(trimmed, []byte("null")) {
		v.Set(reflect.Zero(v.Type()))
		return nil
	}
	parsed, err := unquoteIfNeeded(trimmed)
	if err != nil {
		return ErrMalformedTOON
	}
	return setFieldBytes(v, parsed)
}

func unmarshalObjectLines(data []byte, v reflect.Value) error {
	info := getStructInfo(v.Type())
	lines := bytes.Split(data, []byte{'\n'})

	for _, rawLine := range lines {
		line := bytes.TrimSpace(rawLine)
		if len(line) == 0 {
			continue
		}

		keyRaw, valRaw, ok := splitObjectLine(line)
		if !ok {
			return ErrMalformedTOON
		}
		key, err := unquoteIfNeeded(bytes.TrimSpace(keyRaw))
		if err != nil {
			return ErrMalformedTOON
		}
		val := bytes.TrimSpace(valRaw)
		if len(key) == 0 {
			return ErrMalformedTOON
		}

		// Empty nested object header like "user:" is currently treated as a
		// structural marker and ignored for struct decoding in this fallback path.
		if len(val) == 0 {
			continue
		}

		idx := info.findFieldIndex(key)
		if idx < 0 {
			// Unknown field: ignore for forward compatibility.
			continue
		}

		field := v.Field(idx)
		if !field.CanSet() {
			continue
		}

		if bytes.Equal(val, []byte("null")) {
			field.Set(reflect.Zero(field.Type()))
			continue
		}

		parsedVal, err := unquoteIfNeeded(val)
		if err != nil {
			return ErrMalformedTOON
		}

		if err := setFieldBytes(field, parsedVal); err != nil {
			return err
		}
	}

	return nil
}

func splitObjectLine(line []byte) (key []byte, val []byte, ok bool) {
	inQuotes := false
	escape := false
	for i, b := range line {
		if escape {
			escape = false
			continue
		}
		if b == '\\' && inQuotes {
			escape = true
			continue
		}
		if b == '"' {
			inQuotes = !inQuotes
			continue
		}
		if b == ':' && !inQuotes {
			return line[:i], line[i+1:], true
		}
	}
	return nil, nil, false
}

func unquoteIfNeeded(b []byte) ([]byte, error) {
	if len(b) >= 2 && b[0] == '"' && b[len(b)-1] == '"' {
		s, err := strconv.Unquote(string(b))
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	}
	return b, nil
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

// header represents parsed TOON header using []byte (zero-copy)
type header struct {
	name   []byte
	size   int
	fields [][]byte
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
				h.name = d.data[start : d.pos-1]
			}
			size, err := d.parseSize()
			if err != nil {
				return nil, err
			}
			h.size = size

		case BlockStart:
			if len(h.name) == 0 && d.pos-1 > start {
				h.name = d.data[start : d.pos-1]
			}
			fields, err := d.parseFields()
			if err != nil {
				return nil, err
			}
			h.fields = fields

		case HeaderEnd:
			if len(h.name) == 0 && d.pos-1 > start {
				h.name = d.data[start : d.pos-1]
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
func (d *decoder) parseFields() ([][]byte, error) {
	var fields [][]byte
	start := d.pos

	for {
		b, ok := d.next()
		if !ok {
			return nil, ErrMalformedTOON
		}

		switch b {
		case Separator:
			if d.pos-1 > start {
				fields = append(fields, d.data[start:d.pos-1])
			}
			start = d.pos

		case BlockEnd:
			if d.pos-1 > start {
				fields = append(fields, d.data[start:d.pos-1])
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

	// Use stack-allocated array for field indices (avoids heap allocation)
	var fieldIdxArr [64]int
	fieldIdx := fieldIdxArr[:len(h.fields)]
	for i, name := range h.fields {
		fieldIdx[i] = info.findFieldIndex(name)
	}

	// Parse CSV values after header
	for _, idx := range fieldIdx {
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
		if idx < 0 {
			continue
		}

		// Set field value
		field := v.Field(idx)
		if err := setFieldBytes(field, value); err != nil {
			return err
		}
	}

	return nil
}

// decodeSlice decodes multiple CSV rows into slice
func (d *decoder) decodeSlice(h *header, v reflect.Value) error {
	elemType := v.Type().Elem()

	// Pre-allocate slice if size is known from header
	var newSlice reflect.Value
	if h.size > 0 {
		newSlice = reflect.MakeSlice(v.Type(), h.size, h.size)
	}

	rowIdx := 0
	for {
		d.skipWhitespace()

		if _, ok := d.peek(); !ok {
			break
		}

		// Get or create element
		var elem reflect.Value
		if h.size > 0 && rowIdx < h.size {
			elem = newSlice.Index(rowIdx)
			rowIdx++
		} else {
			elem = reflect.New(elemType).Elem()
		}

		if err := d.decodeStruct(h, elem); err != nil {
			return err
		}

		if h.size <= 0 {
			newSlice = reflect.Append(newSlice, elem)
		}

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

	v.Set(newSlice)
	return nil
}

// setFieldBytes converts []byte value and sets it to reflect.Value
func setFieldBytes(v reflect.Value, b []byte) error {
	// Check for custom Unmarshaler interface
	if v.CanAddr() {
		if m, ok := v.Addr().Interface().(Unmarshaler); ok {
			return m.UnmarshalTOON(b)
		}
	}

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
		n, err := parseFloatBytes(b)
		if err != nil {
			return ErrMalformedTOON
		}
		v.SetFloat(n)
	case reflect.Bool:
		bval, err := parseBoolBytes(b)
		if err != nil {
			return ErrMalformedTOON
		}
		v.SetBool(bval)
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

// parseFloatBytes parses float directly from []byte
func parseFloatBytes(b []byte) (float64, error) {
	if len(b) == 0 {
		return 0, ErrMalformedTOON
	}
	var n float64
	var div float64 = 1
	var frac bool
	var neg bool
	i := 0
	if b[0] == '-' {
		neg = true
		i = 1
	}
	for ; i < len(b); i++ {
		c := b[i]
		if c == '.' {
			frac = true
			continue
		}
		if c < '0' || c > '9' {
			return 0, ErrMalformedTOON
		}
		if frac {
			div *= 10
			n = n + float64(c-'0')/div
		} else {
			n = n*10 + float64(c-'0')
		}
	}
	if neg {
		n = -n
	}
	return n, nil
}

// parseBoolBytes parses bool from []byte (fast path for +/-/true/false/1/0)
func parseBoolBytes(b []byte) (bool, error) {
	switch len(b) {
	case 1:
		switch b[0] {
		case '+', '1', 't', 'T':
			return true, nil
		case '-', '0', 'f', 'F':
			return false, nil
		}
	case 4:
		if (b[0] == 't' || b[0] == 'T') &&
			(b[1] == 'r' || b[1] == 'R') &&
			(b[2] == 'u' || b[2] == 'U') &&
			(b[3] == 'e' || b[3] == 'E') {
			return true, nil
		}
	case 5:
		if (b[0] == 'f' || b[0] == 'F') &&
			(b[1] == 'a' || b[1] == 'A') &&
			(b[2] == 'l' || b[2] == 'L') &&
			(b[3] == 's' || b[3] == 'S') &&
			(b[4] == 'e' || b[4] == 'E') {
			return false, nil
		}
	}
	return false, ErrMalformedTOON
}
