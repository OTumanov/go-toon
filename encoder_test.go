package toon

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncodeNull(t *testing.T) {
	var buf bytes.Buffer
	err := Encode(&buf, NullValue())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if buf.String() != "~" {
		t.Errorf("expected ~, got %q", buf.String())
	}
}

func TestEncodeBoolean(t *testing.T) {
	tests := []struct {
		name     string
		value    bool
		expected string
	}{
		{"true", true, "+"},
		{"false", false, "-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Encode(&buf, BoolValue(tt.value))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if buf.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, buf.String())
			}
		})
	}
}

func TestEncodeNumber(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{"integer", 42, "42"},
		{"negative", -17, "-17"},
		{"float", 3.14, "3.14"},
		{"negative float", -0.5, "-0.5"},
		{"zero", 0, "0"},
		{"large", 1e10, "10000000000"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Encode(&buf, NumberValue(tt.value))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if buf.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, buf.String())
			}
		})
	}
}

func TestEncodeString(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"simple", "hello", `"hello"`},
		{"empty", "", `""`},
		{"with space", "hello world", `"hello world"`},
		{"with newline", "hello\nworld", `"hello\nworld"`},
		{"with quote", `say "hi"`, `"say \"hi\""`},
		{"with backslash", `path\to\file`, `"path\\to\\file"`},
		{"unicode", "привет", `"привет"`},
		{"emoji", "🎉", `"🎉"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Encode(&buf, StringValue(tt.value))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if buf.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, buf.String())
			}
		})
	}
}

func TestEncodeArray(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected string
	}{
		{
			name:     "empty",
			value:    ArrayValue(),
			expected: "[]",
		},
		{
			name:     "single",
			value:    ArrayValue(NumberValue(1)),
			expected: "[1]",
		},
		{
			name:     "multiple",
			value:    ArrayValue(NumberValue(1), NumberValue(2), NumberValue(3)),
			expected: "[1 2 3]",
		},
		{
			name:     "mixed",
			value:    ArrayValue(NumberValue(1), StringValue("two"), BoolValue(true), NullValue()),
			expected: `[1 "two" + ~]`,
		},
		{
			name:     "nested",
			value:    ArrayValue(ArrayValue(NumberValue(1), NumberValue(2)), ArrayValue(NumberValue(3), NumberValue(4))),
			expected: "[[1 2] [3 4]]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Encode(&buf, tt.value)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if buf.String() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, buf.String())
			}
		})
	}
}

func TestEncodeObject(t *testing.T) {
	tests := []struct {
		name         string
		value        Value
		expectedKeys map[string]Type
	}{
		{
			name:         "empty",
			value:        ObjectValue(map[string]Value{}),
			expectedKeys: map[string]Type{},
		},
		{
			name: "single key",
			value: ObjectValue(map[string]Value{
				"name": StringValue("John"),
			}),
			expectedKeys: map[string]Type{
				"name": String,
			},
		},
		{
			name: "multiple keys",
			value: ObjectValue(map[string]Value{
				"name":   StringValue("John"),
				"age":    NumberValue(30),
				"active": BoolValue(true),
			}),
			expectedKeys: map[string]Type{
				"name":   String,
				"age":    Number,
				"active": Boolean,
			},
		},
		{
			name: "nested object",
			value: ObjectValue(map[string]Value{
				"user": ObjectValue(map[string]Value{
					"name": StringValue("John"),
				}),
				"count": NumberValue(1),
			}),
			expectedKeys: map[string]Type{
				"user":  Object,
				"count": Number,
			},
		},
		{
			name: "array value",
			value: ObjectValue(map[string]Value{
				"items": ArrayValue(NumberValue(1), NumberValue(2), NumberValue(3)),
			}),
			expectedKeys: map[string]Type{
				"items": Array,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			err := Encode(&buf, tt.value)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			// Parse back and verify structure
			parsed, err := ParseString(buf.String())
			if err != nil {
				t.Errorf("failed to parse encoded: %v", err)
				return
			}

			if !parsed.IsObject() {
				t.Errorf("expected object, got %s", parsed.Type())
				return
			}

			if parsed.Len() != len(tt.expectedKeys) {
				t.Errorf("expected %d keys, got %d", len(tt.expectedKeys), parsed.Len())
			}

			for key, expectedType := range tt.expectedKeys {
				val := parsed.Get(key)
				if val.IsNull() && expectedType != Null {
					t.Errorf("missing key %q", key)
					continue
				}
				if val.Type() != expectedType {
					t.Errorf("key %q: expected %s, got %s", key, expectedType, val.Type())
				}
			}
		})
	}
}

func TestEncodeComplex(t *testing.T) {
	val := ObjectValue(map[string]Value{
		"users": ArrayValue(
			ObjectValue(map[string]Value{
				"name": StringValue("Alice"),
				"age":  NumberValue(25),
			}),
			ObjectValue(map[string]Value{
				"name": StringValue("Bob"),
				"age":  NumberValue(30),
			}),
		),
		"count": NumberValue(2),
		"valid": BoolValue(true),
	})

	var buf bytes.Buffer
	err := Encode(&buf, val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Parse back and verify structure
	parsed, err := ParseString(buf.String())
	if err != nil {
		t.Fatalf("failed to parse encoded: %v", err)
	}

	if !parsed.IsObject() {
		t.Fatalf("expected object, got %s", parsed.Type())
	}

	// Verify count
	if parsed.Get("count").Number() != 2 {
		t.Errorf("expected count=2, got %v", parsed.Get("count").Number())
	}

	// Verify valid
	if !parsed.Get("valid").Bool() {
		t.Error("expected valid=true")
	}

	// Verify users array
	users := parsed.Get("users")
	if !users.IsArray() {
		t.Fatalf("expected users to be array, got %s", users.Type())
	}
	if users.Len() != 2 {
		t.Errorf("expected 2 users, got %d", users.Len())
	}

	// Verify first user
	firstUser := users.Index(0)
	if firstUser.Get("name").String() != "Alice" {
		t.Errorf("expected first user name=Alice, got %v", firstUser.Get("name").String())
	}
	if firstUser.Get("age").Number() != 25 {
		t.Errorf("expected first user age=25, got %v", firstUser.Get("age").Number())
	}
}

func TestEncodeToString(t *testing.T) {
	val := ObjectValue(map[string]Value{
		"name": StringValue("test"),
		"value": NumberValue(42),
	})

	s, err := EncodeToString(val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := `{name "test" value 42}`
	if s != expected {
		t.Errorf("expected %q, got %q", expected, s)
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []string{
		`~`,
		`+`,
		`-`,
		`42`,
		`-17`,
		`3.14`,
		`"hello"`,
		`"hello world"`,
		`[]`,
		`[1 2 3]`,
		`[1 "two" + ~]`,
		`[[1 2] [3 4]]`,
		`{}`,
		`{name "John"}`,
		`{name "John" age 30}`,
		`{items [1 2 3]}`,
		`{user {name "Alice"} count 1}`,
	}

	for _, input := range tests {
		t.Run(strings.ReplaceAll(input, " ", "_"), func(t *testing.T) {
			// Parse
			val, err := ParseString(input)
			if err != nil {
				t.Fatalf("parse error: %v", err)
			}

			// Encode
			output, err := EncodeToString(val)
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}

			// Parse again
			val2, err := ParseString(output)
			if err != nil {
				t.Fatalf("second parse error: %v", err)
			}

			// Encode again
			output2, err := EncodeToString(val2)
			if err != nil {
				t.Fatalf("second encode error: %v", err)
			}

			// Should be the same
			if output != output2 {
				t.Errorf("round-trip mismatch: %q != %q", output, output2)
			}
		})
	}
}
