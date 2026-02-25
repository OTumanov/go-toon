package toon

import (
	"strings"
	"testing"
)

func TestParseNull(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		errContains string
	}{
		{"simple null", "~", false, ""},
		{"null with whitespace", "  ~  ", false, ""},
		{"null with tabs", "\t~\t", false, ""},
		{"invalid - uppercase", "N", true, "unexpected token"},
		{"invalid - word", "null", true, "unexpected token"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := Parse(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				} else if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !val.IsNull() {
				t.Errorf("expected null, got %s", val.Type())
			}
		})
	}
}

func TestParseBoolean(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
		wantErr  bool
	}{
		{"true", "+", true, false},
		{"false", "-", false, false},
		{"true with space", "  +  ", true, false},
		{"false with space", "  -  ", false, false},
		{"invalid - T", "T", false, true},
		{"invalid - F", "F", false, true},
		{"invalid - true", "true", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := Parse(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !val.IsBool() {
				t.Errorf("expected boolean, got %s", val.Type())
				return
			}
			if val.Bool() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, val.Bool())
			}
		})
	}
}

func TestParseNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
		wantErr  bool
	}{
		{"integer", "42", 42, false},
		{"negative", "-17", -17, false},
		{"float", "3.14", 3.14, false},
		{"negative float", "-0.5", -0.5, false},
		{"zero", "0", 0, false},
		{"with whitespace", "  123  ", 123, false},
		{"scientific", "1e10", 1e10, false},
		{"scientific negative", "1e-5", 1e-5, false},
		{"scientific positive", "2.5e+3", 2500, false},
		{"invalid - letters", "abc", 0, true},
		{"invalid - mixed", "12a34", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := Parse(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !val.IsNumber() {
				t.Errorf("expected number, got %s", val.Type())
				return
			}
			if val.Number() != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, val.Number())
			}
		})
	}
}

func TestParseString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{"simple", `"hello"`, "hello", false},
		{"empty", `""`, "", false},
		{"with spaces", `  "world"  `, "world", false},
		{"with escape", `"hello\nworld"`, "hello\nworld", false},
		{"with quote escape", `"say \"hi\""`, `say "hi"`, false},
		{"with backslash", `"path\\to\\file"`, `path\to\file`, false},
		{"unicode", `"привет"`, "привет", false},
		{"emoji", `"🎉"`, "🎉", false},
		{"unterminated", `"hello`, "", true},
		{"invalid escape", `"\q"`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := Parse(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !val.IsString() {
				t.Errorf("expected string, got %s", val.Type())
				return
			}
			if val.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, val.String())
			}
		})
	}
}

func TestParseArray(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Value
		wantErr  bool
	}{
		{
			name:     "empty",
			input:    "[]",
			expected: []Value{},
			wantErr:  false,
		},
		{
			name:     "single number",
			input:    "[1]",
			expected: []Value{NumberValue(1)},
			wantErr:  false,
		},
		{
			name:     "multiple numbers",
			input:    "[1 2 3]",
			expected: []Value{NumberValue(1), NumberValue(2), NumberValue(3)},
			wantErr:  false,
		},
		{
			name:     "mixed types",
			input:    `[1 "two" + ~]`,
			expected: []Value{NumberValue(1), StringValue("two"), BoolValue(true), NullValue()},
			wantErr:  false,
		},
		{
			name:     "nested array",
			input:    "[[1 2] [3 4]]",
			expected: []Value{ArrayValue(NumberValue(1), NumberValue(2)), ArrayValue(NumberValue(3), NumberValue(4))},
			wantErr:  false,
		},
		{
			name:     "with whitespace",
			input:    "[  1  2  3  ]",
			expected: []Value{NumberValue(1), NumberValue(2), NumberValue(3)},
			wantErr:  false,
		},
		{
			name:     "unterminated",
			input:    "[1 2",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := Parse(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !val.IsArray() {
				t.Errorf("expected array, got %s", val.Type())
				return
			}
			arr := val.Array()
			if len(arr) != len(tt.expected) {
				t.Errorf("expected len %d, got %d", len(tt.expected), len(arr))
				return
			}
		})
	}
}

func TestParseObject(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		checkFn  func(t *testing.T, v Value)
	}{
		{
			name:    "empty",
			input:   "{}",
			wantErr: false,
			checkFn: func(t *testing.T, v Value) {
				if v.Len() != 0 {
					t.Errorf("expected empty object, got len %d", v.Len())
				}
			},
		},
		{
			name:    "single key",
			input:   `{name "John"}`,
			wantErr: false,
			checkFn: func(t *testing.T, v Value) {
				if v.Get("name").String() != "John" {
					t.Errorf("expected name=John, got %v", v.Get("name"))
				}
			},
		},
		{
			name:    "multiple keys",
			input:   `{name "John" age 30 active +}`,
			wantErr: false,
			checkFn: func(t *testing.T, v Value) {
				if v.Get("name").String() != "John" {
					t.Errorf("expected name=John")
				}
				if v.Get("age").Number() != 30 {
					t.Errorf("expected age=30")
				}
				if !v.Get("active").Bool() {
					t.Errorf("expected active=true")
				}
			},
		},
		{
			name:    "nested object",
			input:   `{user {name "John"} count 1}`,
			wantErr: false,
			checkFn: func(t *testing.T, v Value) {
				user := v.Get("user")
				if !user.IsObject() {
					t.Errorf("expected user to be object")
					return
				}
				if user.Get("name").String() != "John" {
					t.Errorf("expected user.name=John")
				}
			},
		},
		{
			name:    "array value",
			input:   `{items [1 2 3]}`,
			wantErr: false,
			checkFn: func(t *testing.T, v Value) {
				items := v.Get("items")
				if !items.IsArray() {
					t.Errorf("expected items to be array")
					return
				}
				if items.Len() != 3 {
					t.Errorf("expected 3 items, got %d", items.Len())
				}
			},
		},
		{
			name:    "unterminated",
			input:   `{name "John"`,
			wantErr: true,
		},
		{
			name:    "missing value",
			input:   `{name}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val, err := Parse(strings.NewReader(tt.input))
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if !val.IsObject() {
				t.Errorf("expected object, got %s", val.Type())
				return
			}
			if tt.checkFn != nil {
				tt.checkFn(t, val)
			}
		})
	}
}

func TestParseComplex(t *testing.T) {
	input := `{
		users [
			{name "Alice" age 25}
			{name "Bob" age 30}
		]
		count 2
		valid +
	}`

	val, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !val.IsObject() {
		t.Fatalf("expected object, got %s", val.Type())
	}

	users := val.Get("users")
	if !users.IsArray() {
		t.Errorf("expected users to be array")
	}
	if users.Len() != 2 {
		t.Errorf("expected 2 users, got %d", users.Len())
	}

	firstUser := users.Index(0)
	if firstUser.Get("name").String() != "Alice" {
		t.Errorf("expected first user name=Alice")
	}
	if firstUser.Get("age").Number() != 25 {
		t.Errorf("expected first user age=25")
	}

	if val.Get("count").Number() != 2 {
		t.Errorf("expected count=2")
	}

	if !val.Get("valid").Bool() {
		t.Errorf("expected valid=true")
	}
}

func TestParseEmptyInput(t *testing.T) {
	_, err := Parse(strings.NewReader(""))
	if err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseOnlyWhitespace(t *testing.T) {
	_, err := Parse(strings.NewReader("   \n\t  "))
	if err == nil {
		t.Error("expected error for whitespace-only input")
	}
}
