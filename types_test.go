package toon

import (
	"testing"
)

func TestTypeString(t *testing.T) {
	tests := []struct {
		typ      Type
		expected string
	}{
		{Null, "null"},
		{Boolean, "boolean"},
		{Number, "number"},
		{String, "string"},
		{Array, "array"},
		{Object, "object"},
		{Type(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.typ.String(); got != tt.expected {
				t.Errorf("Type.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestValueConstructors(t *testing.T) {
	t.Run("NullValue", func(t *testing.T) {
		v := NullValue()
		if v.Type() != Null {
			t.Errorf("expected Null, got %s", v.Type())
		}
		if !v.IsNull() {
			t.Error("expected IsNull() to be true")
		}
	})

	t.Run("BoolValue", func(t *testing.T) {
		v := BoolValue(true)
		if v.Type() != Boolean {
			t.Errorf("expected Boolean, got %s", v.Type())
		}
		if !v.IsBool() {
			t.Error("expected IsBool() to be true")
		}
		if !v.Bool() {
			t.Error("expected Bool() to be true")
		}
	})

	t.Run("NumberValue", func(t *testing.T) {
		v := NumberValue(42.5)
		if v.Type() != Number {
			t.Errorf("expected Number, got %s", v.Type())
		}
		if !v.IsNumber() {
			t.Error("expected IsNumber() to be true")
		}
		if v.Number() != 42.5 {
			t.Errorf("expected 42.5, got %f", v.Number())
		}
	})

	t.Run("StringValue", func(t *testing.T) {
		v := StringValue("hello")
		if v.Type() != String {
			t.Errorf("expected String, got %s", v.Type())
		}
		if !v.IsString() {
			t.Error("expected IsString() to be true")
		}
		if v.String() != "hello" {
			t.Errorf("expected \"hello\", got %q", v.String())
		}
	})

	t.Run("ArrayValue", func(t *testing.T) {
		v := ArrayValue(NumberValue(1), NumberValue(2))
		if v.Type() != Array {
			t.Errorf("expected Array, got %s", v.Type())
		}
		if !v.IsArray() {
			t.Error("expected IsArray() to be true")
		}
		if v.Len() != 2 {
			t.Errorf("expected len 2, got %d", v.Len())
		}
	})

	t.Run("ObjectValue", func(t *testing.T) {
		v := ObjectValue(map[string]Value{
			"key": StringValue("value"),
		})
		if v.Type() != Object {
			t.Errorf("expected Object, got %s", v.Type())
		}
		if !v.IsObject() {
			t.Error("expected IsObject() to be true")
		}
		if v.Len() != 1 {
			t.Errorf("expected len 1, got %d", v.Len())
		}
	})
}

func TestValueGet(t *testing.T) {
	obj := ObjectValue(map[string]Value{
		"name": StringValue("John"),
		"age":  NumberValue(30),
	})

	t.Run("existing key", func(t *testing.T) {
		v := obj.Get("name")
		if v.String() != "John" {
			t.Errorf("expected \"John\", got %q", v.String())
		}
	})

	t.Run("non-existing key", func(t *testing.T) {
		v := obj.Get("missing")
		if !v.IsNull() {
			t.Error("expected null for missing key")
		}
	})

	t.Run("on non-object", func(t *testing.T) {
		v := StringValue("test").Get("key")
		if !v.IsNull() {
			t.Error("expected null for non-object")
		}
	})
}

func TestValueIndex(t *testing.T) {
	arr := ArrayValue(NumberValue(10), NumberValue(20), NumberValue(30))

	t.Run("valid index", func(t *testing.T) {
		v := arr.Index(1)
		if v.Number() != 20 {
			t.Errorf("expected 20, got %f", v.Number())
		}
	})

	t.Run("negative index", func(t *testing.T) {
		v := arr.Index(-1)
		if !v.IsNull() {
			t.Error("expected null for negative index")
		}
	})

	t.Run("out of bounds", func(t *testing.T) {
		v := arr.Index(10)
		if !v.IsNull() {
			t.Error("expected null for out of bounds")
		}
	})

	t.Run("on non-array", func(t *testing.T) {
		v := StringValue("test").Index(0)
		if !v.IsNull() {
			t.Error("expected null for non-array")
		}
	})
}

func TestValuePanics(t *testing.T) {
	t.Run("Bool on non-bool", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		StringValue("test").Bool()
	})

	t.Run("Number on non-number", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		StringValue("test").Number()
	})

	t.Run("String on non-string", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expected panic")
			}
		}()
		NumberValue(42).String()
	})
}
