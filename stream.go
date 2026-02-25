package toon

import (
	"bufio"
	"io"
)

// StreamEncoder writes TOON-encoded values to an output stream
type StreamEncoder struct {
	w   io.Writer
	buf []byte
}

// NewEncoder creates a new StreamEncoder writing to w
func NewEncoder(w io.Writer) *StreamEncoder {
	return &StreamEncoder{w: w}
}

// Encode writes v to the stream in TOON format
func (e *StreamEncoder) Encode(v interface{}) error {
	// Reuse internal buffer if available
	var err error
	if e.buf != nil {
		e.buf = e.buf[:0]
		e.buf, err = marshalTo(v, e.buf)
	} else {
		e.buf, err = Marshal(v)
	}
	if err != nil {
		return err
	}

	// Add newline for stream delimiter
	e.buf = append(e.buf, '\n')
	_, err = e.w.Write(e.buf)
	return err
}

// Decoder reads TOON-encoded values from an input stream
type Decoder struct {
	r   *bufio.Reader
	buf []byte
}

// NewDecoder creates a new Decoder reading from r
func NewDecoder(r io.Reader) *Decoder {
	br, ok := r.(*bufio.Reader)
	if !ok {
		br = bufio.NewReader(r)
	}
	return &Decoder{r: br}
}

// Decode reads the next TOON value from the stream into v
func (d *Decoder) Decode(v interface{}) error {
	// Read header line (ends with \n or EOF)
	line, err := d.r.ReadSlice('\n')
	if err != nil && err != io.EOF {
		return err
	}
	if len(line) == 0 && err == io.EOF {
		return io.EOF
	}

	// Find colon to split header and body
	colonIdx := -1
	for i, b := range line {
		if b == HeaderEnd {
			colonIdx = i
			break
		}
	}
	if colonIdx == -1 {
		return ErrMalformedTOON
	}

	// Parse header to get size
	dec := &decoder{data: line, pos: 0}
	h, err := dec.parseHeader()
	if err != nil {
		return err
	}

	// Start building full data
	data := make([]byte, 0, len(line)*2)
	data = append(data, line...)

	// For slices with known size, read remaining rows
	if h.size > 1 {
		for i := 1; i < h.size; i++ {
			row, err := d.r.ReadSlice('\n')
			if err != nil && err != io.EOF {
				return err
			}
			data = append(data, row...)
			if err == io.EOF {
				break
			}
		}
	}

	return Unmarshal(data, v)
}



// marshalTo marshals v into provided buffer
func marshalTo(v interface{}, buf []byte) ([]byte, error) {
	// For now, use Marshal and copy to buffer
	// TODO: implement zero-copy marshaling into existing buffer
	data, err := Marshal(v)
	if err != nil {
		return buf, err
	}
	return append(buf, data...), nil
}
