package toon

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Parser handles TOON parsing
type Parser struct {
	r   *bufio.Reader
	pos int
}

// Parse parses TOON data from a reader and returns a Value
func Parse(r io.Reader) (Value, error) {
	p := &Parser{r: bufio.NewReader(r)}
	val, err := p.parseValue()
	if err != nil {
		return val, err
	}
	
	// After parsing a value at top level, there should only be whitespace left
	p.skipWhitespace()
	if ch, err := p.peek(); err == nil {
		return Value{}, fmt.Errorf("unexpected token after value: %c", ch)
	}
	
	return val, nil
}

// ParseString parses TOON data from a string and returns a Value
func ParseString(s string) (Value, error) {
	return Parse(strings.NewReader(s))
}

func (p *Parser) parseValue() (Value, error) {
	p.skipWhitespace()

	ch, err := p.peek()
	if err != nil {
		return Value{}, fmt.Errorf("unexpected EOF")
	}

	switch ch {
	case '~':
		return p.parseNull()
	case '+', '-':
		// Check if it's a number or boolean
		next, err := p.peekAhead(1)
		if err != nil || (!isDigit(next) && next != '.') {
			return p.parseBoolean()
		}
		return p.parseNumber()
	case '"':
		return p.parseString()
	case '[':
		return p.parseArray()
	case '{':
		return p.parseObject()
	default:
		if isDigit(ch) || ch == '.' {
			return p.parseNumber()
		}
		return Value{}, fmt.Errorf("unexpected token: %c", ch)
	}
}



func (p *Parser) parseNull() (Value, error) {
	ch, err := p.read()
	if err != nil || ch != '~' {
		return Value{}, fmt.Errorf("expected ~")
	}
	return NullValue(), nil
}

func (p *Parser) parseBoolean() (Value, error) {
	ch, err := p.read()
	if err != nil {
		return Value{}, fmt.Errorf("unexpected EOF")
	}

	switch ch {
	case '+':
		return BoolValue(true), nil
	case '-':
		return BoolValue(false), nil
	default:
		return Value{}, fmt.Errorf("expected + or - for boolean")
	}
}

func (p *Parser) parseNumber() (Value, error) {
	var sb strings.Builder

	// Optional sign
	ch, err := p.peek()
	if err != nil {
		return Value{}, fmt.Errorf("unexpected EOF")
	}
	if ch == '+' || ch == '-' {
		p.read()
		sb.WriteByte(ch)
	}

	// Integer part
	for {
		ch, err := p.peek()
		if err != nil || !isDigit(ch) {
			break
		}
		p.read()
		sb.WriteByte(ch)
	}

	// Fractional part
	ch, err = p.peek()
	if err == nil && ch == '.' {
		p.read()
		sb.WriteByte(ch)

		for {
			ch, err := p.peek()
			if err != nil || !isDigit(ch) {
				break
			}
			p.read()
			sb.WriteByte(ch)
		}
	}

	// Exponent part
	ch, err = p.peek()
	if err == nil && (ch == 'e' || ch == 'E') {
		p.read()
		sb.WriteByte(ch)

		ch, err = p.peek()
		if err == nil && (ch == '+' || ch == '-') {
			p.read()
			sb.WriteByte(ch)
		}

		for {
			ch, err := p.peek()
			if err != nil || !isDigit(ch) {
				break
			}
			p.read()
			sb.WriteByte(ch)
		}
	}

	numStr := sb.String()
	if numStr == "" || numStr == "+" || numStr == "-" {
		return Value{}, fmt.Errorf("invalid number")
	}

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return Value{}, fmt.Errorf("invalid number: %s", numStr)
	}

	return NumberValue(num), nil
}

func (p *Parser) parseString() (Value, error) {
	ch, err := p.read()
	if err != nil || ch != '"' {
		return Value{}, fmt.Errorf("expected \"")
	}

	var sb strings.Builder
	for {
		ch, err := p.read()
		if err != nil {
			return Value{}, fmt.Errorf("unterminated string")
		}

		if ch == '"' {
			break
		}

		if ch == '\\' {
			next, err := p.read()
			if err != nil {
				return Value{}, fmt.Errorf("unterminated escape sequence")
			}

			switch next {
			case '"':
				sb.WriteByte('"')
			case '\\':
				sb.WriteByte('\\')
			case '/':
				sb.WriteByte('/')
			case 'b':
				sb.WriteByte('\b')
			case 'f':
				sb.WriteByte('\f')
			case 'n':
				sb.WriteByte('\n')
			case 'r':
				sb.WriteByte('\r')
			case 't':
				sb.WriteByte('\t')
			case 'u':
				// Unicode escape
				code, err := p.parseUnicodeEscape()
				if err != nil {
					return Value{}, err
				}
				sb.WriteRune(code)
			default:
				return Value{}, fmt.Errorf("invalid escape sequence: \\%c", next)
			}
		} else {
			sb.WriteByte(ch)
		}
	}

	return StringValue(sb.String()), nil
}

func (p *Parser) parseUnicodeEscape() (rune, error) {
	var code uint32
	for i := 0; i < 4; i++ {
		ch, err := p.read()
		if err != nil {
			return 0, fmt.Errorf("incomplete unicode escape")
		}

		var digit uint32
		switch {
		case ch >= '0' && ch <= '9':
			digit = uint32(ch - '0')
		case ch >= 'a' && ch <= 'f':
			digit = uint32(ch - 'a' + 10)
		case ch >= 'A' && ch <= 'F':
			digit = uint32(ch - 'A' + 10)
		default:
			return 0, fmt.Errorf("invalid hex digit: %c", ch)
		}

		code = code<<4 | digit
	}
	return rune(code), nil
}

func (p *Parser) parseArray() (Value, error) {
	ch, err := p.read()
	if err != nil || ch != '[' {
		return Value{}, fmt.Errorf("expected [")
	}

	var items []Value
	for {
		p.skipWhitespace()

		ch, err := p.peek()
		if err != nil {
			return Value{}, fmt.Errorf("unterminated array")
		}

		if ch == ']' {
			p.read()
			break
		}

		item, err := p.parseValue()
		if err != nil {
			return Value{}, err
		}
		items = append(items, item)
	}

	return ArrayValue(items...), nil
}

func (p *Parser) parseObject() (Value, error) {
	ch, err := p.read()
	if err != nil || ch != '{' {
		return Value{}, fmt.Errorf("expected {")
	}

	obj := make(map[string]Value)
	for {
		p.skipWhitespace()

		ch, err := p.peek()
		if err != nil {
			return Value{}, fmt.Errorf("unterminated object")
		}

		if ch == '}' {
			p.read()
			break
		}

		// Parse key
		key, err := p.parseKey()
		if err != nil {
			return Value{}, err
		}

		p.skipWhitespace()

		// Parse value
		value, err := p.parseValue()
		if err != nil {
			return Value{}, err
		}

		obj[key] = value
	}

	return ObjectValue(obj), nil
}

func (p *Parser) parseKey() (string, error) {
	p.skipWhitespace()

	ch, err := p.peek()
	if err != nil {
		return "", fmt.Errorf("unexpected EOF while parsing key")
	}

	// Keys can be identifiers or strings
	if ch == '"' {
		val, err := p.parseString()
		if err != nil {
			return "", err
		}
		return val.String(), nil
	}

	// Parse identifier
	var sb strings.Builder
	for {
		ch, err := p.peek()
		if err != nil {
			break
		}

		if !isIdentifierChar(ch) {
			break
		}

		p.read()
		sb.WriteByte(ch)
	}

	key := sb.String()
	if key == "" {
		return "", fmt.Errorf("expected object key")
	}

	return key, nil
}

func (p *Parser) skipWhitespace() {
	for {
		ch, err := p.peek()
		if err != nil {
			return
		}
		if !unicode.IsSpace(rune(ch)) {
			return
		}
		p.read()
	}
}

func (p *Parser) peek() (byte, error) {
	ch, err := p.r.Peek(1)
	if err != nil {
		return 0, err
	}
	return ch[0], nil
}

func (p *Parser) peekAhead(n int) (byte, error) {
	ch, err := p.r.Peek(n + 1)
	if err != nil {
		return 0, err
	}
	return ch[n], nil
}

func (p *Parser) read() (byte, error) {
	ch, err := p.r.ReadByte()
	if err != nil {
		return 0, err
	}
	p.pos++
	return ch, nil
}

func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

func isIdentifierChar(ch byte) bool {
	return unicode.IsLetter(rune(ch)) || unicode.IsDigit(rune(ch)) || ch == '_' || ch == '-'
}
