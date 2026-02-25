package toon

import (
	"reflect"
	"testing"
)

type TestUser struct {
	ID    int
	Name  string
	Email string `toon:"email"`
	age   int    // unexported
}

func TestBuildFieldMap(t *testing.T) {
	typ := reflect.TypeOf(TestUser{})
	fm := buildFieldMap(typ)
	
	// Should have exported fields only
	if len(fm) != 3 {
		t.Errorf("expected 3 fields, got %d", len(fm))
	}
	
	// Check standard fields
	if idx, ok := fm["ID"]; !ok || idx != 0 {
		t.Errorf("ID: expected index 0, got %d", idx)
	}
	
	if idx, ok := fm["Name"]; !ok || idx != 1 {
		t.Errorf("Name: expected index 1, got %d", idx)
	}
	
	// Check tagged field
	if idx, ok := fm["email"]; !ok || idx != 2 {
		t.Errorf("email: expected index 2, got %d", idx)
	}
	
	// Unexported field should not be present
	if _, ok := fm["age"]; ok {
		t.Error("age should not be in field map (unexported)")
	}
}

func TestCacheGet(t *testing.T) {
	typ := reflect.TypeOf(TestUser{})
	
	// First call - should build
	fm1 := defaultCache.get(typ)
	if fm1 == nil {
		t.Fatal("expected field map, got nil")
	}
	
	// Second call - should return cached
	fm2 := defaultCache.get(typ)
	// Compare by checking they have same content (maps can't be compared directly)
	if len(fm1) != len(fm2) {
		t.Error("expected same cached instance size")
	}
}
