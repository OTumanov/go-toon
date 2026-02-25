// Package toon provides high-performance TOON (Token-Oriented Object Notation)
// implementation for Go. TOON saves up to 40% LLM tokens compared to JSON.
package toon

import (
	"fmt"
)

// Type represents the type of a TOON value
type Type int

const (
	// Null represents a null value
	Null Type = iota
	// Boolean represents a boolean value
	Boolean
	// Number represents a numeric value (integer or float)
	Number
	// String represents a string value
	String
	// Array represents an ordered list of values
	Array
	// Object represents a key-value mapping
	Object
)

// String returns the string representation of the type
func (t Type) String() string {
	switch t {
	case Null:
		return "null"
	case Boolean:
		return "boolean"
	case Number:
		return "number"
	case String:
		return "string"
	case Array:
		return "array"
	case Object:
		return "object"
	default:
		return "unknown"
	}
}

// Value represents a TOON value of any type
type Value struct {
	typ Type

	// Primitive values
	boolVal   bool
	numVal    float64
	strVal    string

	// Complex values
	arrVal []Value
	objVal map[string]Value
}

// Type returns the type of the value
func (v Value) Type() Type {
	return v.typ
}

// IsNull returns true if the value is null
func (v Value) IsNull() bool {
	return v.typ == Null
}

// IsBool returns true if the value is a boolean
func (v Value) IsBool() bool {
	return v.typ == Boolean
}

// IsNumber returns true if the value is a number
func (v Value) IsNumber() bool {
	return v.typ == Number
}

// IsString returns true if the value is a string
func (v Value) IsString() bool {
	return v.typ == String
}

// IsArray returns true if the value is an array
func (v Value) IsArray() bool {
	return v.typ == Array
}

// IsObject returns true if the value is an object
func (v Value) IsObject() bool {
	return v.typ == Object
}

// Bool returns the boolean value.
// Panics if the value is not a boolean.
func (v Value) Bool() bool {
	if v.typ != Boolean {
		panic(fmt.Sprintf("cannot convert %s to bool", v.typ))
	}
	return v.boolVal
}

// Number returns the numeric value.
// Panics if the value is not a number.
func (v Value) Number() float64 {
	if v.typ != Number {
		panic(fmt.Sprintf("cannot convert %s to number", v.typ))
	}
	return v.numVal
}

// String returns the string value.
// Panics if the value is not a string.
func (v Value) String() string {
	if v.typ != String {
		panic(fmt.Sprintf("cannot convert %s to string", v.typ))
	}
	return v.strVal
}

// Array returns the array value.
// Panics if the value is not an array.
func (v Value) Array() []Value {
	if v.typ != Array {
		panic(fmt.Sprintf("cannot convert %s to array", v.typ))
	}
	return v.arrVal
}

// Object returns the object value.
// Panics if the value is not an object.
func (v Value) Object() map[string]Value {
	if v.typ != Object {
		panic(fmt.Sprintf("cannot convert %s to object", v.typ))
	}
	return v.objVal
}

// Get returns the value for the given key in an object.
// Returns a null Value if the key doesn't exist or if the value is not an object.
func (v Value) Get(key string) Value {
	if v.typ != Object {
		return Value{typ: Null}
	}
	if val, ok := v.objVal[key]; ok {
		return val
	}
	return Value{typ: Null}
}

// Index returns the value at the given index in an array.
// Returns a null Value if the index is out of bounds or if the value is not an array.
func (v Value) Index(i int) Value {
	if v.typ != Array {
		return Value{typ: Null}
	}
	if i < 0 || i >= len(v.arrVal) {
		return Value{typ: Null}
	}
	return v.arrVal[i]
}

// Len returns the length of an array or object.
// Returns 0 for other types.
func (v Value) Len() int {
	switch v.typ {
	case Array:
		return len(v.arrVal)
	case Object:
		return len(v.objVal)
	default:
		return 0
	}
}

// Constructors

// NullValue creates a null value
func NullValue() Value {
	return Value{typ: Null}
}

// BoolValue creates a boolean value
func BoolValue(b bool) Value {
	return Value{typ: Boolean, boolVal: b}
}

// NumberValue creates a numeric value
func NumberValue(n float64) Value {
	return Value{typ: Number, numVal: n}
}

// StringValue creates a string value
func StringValue(s string) Value {
	return Value{typ: String, strVal: s}
}

// ArrayValue creates an array value
func ArrayValue(items ...Value) Value {
	return Value{typ: Array, arrVal: items}
}

// ObjectValue creates an object value
func ObjectValue(pairs map[string]Value) Value {
	return Value{typ: Object, objVal: pairs}
}
