package toon

import (
	"bytes"
	"strings"
	"testing"
)

func TestStreamEncoder(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	type User struct {
		ID   int
		Name string
	}

	users := &[]User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}

	err := enc.Encode(users)
	if err != nil {
		t.Fatalf("encode error: %v", err)
	}

	expected := "user[2]{id,name}:1,Alice\n2,Bob\n"
	if buf.String() != expected {
		t.Errorf("expected %q, got %q", expected, buf.String())
	}
}

func TestStreamDecoder(t *testing.T) {
	input := "users[2]{id,name}:1,Alice\n2,Bob"
	dec := NewDecoder(strings.NewReader(input))

	type User struct {
		ID   int
		Name string
	}

	var users []User
	err := dec.Decode(&users)
	if err != nil {
		t.Fatalf("decode error: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].ID != 1 || users[0].Name != "Alice" {
		t.Errorf("user[0]: expected {1 Alice}, got {%d %s}", users[0].ID, users[0].Name)
	}
	if users[1].ID != 2 || users[1].Name != "Bob" {
		t.Errorf("user[1]: expected {2 Bob}, got {%d %s}", users[1].ID, users[1].Name)
	}
}

func TestStreamRoundTrip(t *testing.T) {
	type Product struct {
		SKU   string
		Price float64
	}

	original := &[]Product{
		{SKU: "ABC123", Price: 99.99},
		{SKU: "XYZ789", Price: 49.50},
		{SKU: "DEF456", Price: 149.00},
	}

	// Encode
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(original); err != nil {
		t.Fatalf("encode error: %v", err)
	}

	// Decode - need to use plural form for slice type name
	dec := NewDecoder(&buf)
	var decoded []Product
	if err := dec.Decode(&decoded); err != nil {
		t.Fatalf("decode error: %v", err)
	}

	// Compare
	if len(decoded) != len(*original) {
		t.Fatalf("length mismatch: %d vs %d", len(decoded), len(*original))
	}
	for i, p := range decoded {
		if p.SKU != (*original)[i].SKU {
			t.Errorf("product[%d] SKU: expected %s, got %s", i, (*original)[i].SKU, p.SKU)
		}
		// Float comparison with epsilon
		if diff := p.Price - (*original)[i].Price; diff < -0.001 || diff > 0.001 {
			t.Errorf("product[%d] Price: expected %f, got %f", i, (*original)[i].Price, p.Price)
		}
	}
}

func BenchmarkStreamEncoder(b *testing.B) {
	type User struct {
		ID   int
		Name string
	}
	users := &[]User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}

	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		enc.Encode(users)
	}
}

func BenchmarkStreamDecoder(b *testing.B) {
	input := "user[3]{id,name}:1,Alice\n2,Bob\n3,Charlie\n"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dec := NewDecoder(strings.NewReader(input))
		type User struct {
			ID   int
			Name string
		}
		var users []User
		dec.Decode(&users)
	}
}
