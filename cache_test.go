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

func TestGetStructInfo(t *testing.T) {
	typ := reflect.TypeOf(TestUser{})
	info := getStructInfo(typ)

	// Should have exported fields only
	if len(info.fields) != 3 {
		t.Errorf("expected 3 fields, got %d", len(info.fields))
	}

	// Check struct name (lowercase)
	if info.name != "testuser" {
		t.Errorf("expected name 'testuser', got %q", info.name)
	}

	// Build map for easier checking
	fieldMap := make(map[string]int)
	for _, f := range info.fields {
		fieldMap[f.name] = f.index
	}

	// Check lowercase fields
	if idx, ok := fieldMap["id"]; !ok || idx != 0 {
		t.Errorf("id: expected index 0, got %d", idx)
	}

	if idx, ok := fieldMap["name"]; !ok || idx != 1 {
		t.Errorf("name: expected index 1, got %d", idx)
	}

	// Check tagged field
	if idx, ok := fieldMap["email"]; !ok || idx != 2 {
		t.Errorf("email: expected index 2, got %d", idx)
	}

	// Unexported field should not be present
	if _, ok := fieldMap["age"]; ok {
		t.Error("age should not be in field map (unexported)")
	}
}

func TestCacheConcurrency(t *testing.T) {
	typ := reflect.TypeOf(TestUser{})

	// First call - should build
	info1 := getStructInfo(typ)
	if info1 == nil {
		t.Fatal("expected struct info, got nil")
	}

	// Second call - should return cached
	info2 := getStructInfo(typ)

	// Should be same pointer
	if info1 != info2 {
		t.Error("expected same cached instance")
	}
}

func TestDefaultCacheCompatibility(t *testing.T) {
	typ := reflect.TypeOf(TestUser{})

	// Test legacy API still works
	fm := defaultCache.get(typ)
	if fm == nil {
		t.Fatal("expected field map, got nil")
	}

	if len(fm) != 3 {
		t.Errorf("expected 3 fields, got %d", len(fm))
	}
}
