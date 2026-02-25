package toon

import (
	"testing"
)

type User struct {
	ID   int
	Name string
}

type Product struct {
	SKU   string
	Price float64
	Stock int
}

func TestMarshalStruct(t *testing.T) {
	u := &User{ID: 42, Name: "John"}
	
	data, err := Marshal(u)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	
	expected := "user{id,name}:42,John"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestMarshalSlice(t *testing.T) {
	users := &[]User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}
	
	data, err := Marshal(users)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	
	expected := "user[2]{id,name}:1,Alice\n2,Bob"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestMarshalTypes(t *testing.T) {
	type AllTypes struct {
		Str   string
		Num   int
		Float float64
		Bool  bool
	}
	
	v := &AllTypes{
		Str:   "hello",
		Num:   42,
		Float: 3.14,
		Bool:  true,
	}
	
	data, err := Marshal(v)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	
	expected := "alltypes{str,num,float,bool}:hello,42,3.14,+"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
}

func TestMarshalInvalidTarget(t *testing.T) {
	_, err := Marshal("not a pointer")
	if err != ErrInvalidTarget {
		t.Errorf("expected ErrInvalidTarget, got %v", err)
	}
	
	var i int
	_, err = Marshal(&i)
	if err != ErrInvalidTarget {
		t.Errorf("expected ErrInvalidTarget for non-struct/slice, got %v", err)
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	original := &[]User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
	}
	
	// Marshal
	data, err := Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	
	// Unmarshal
	var decoded []User
	err = Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	
	// Compare
	if len(decoded) != len(*original) {
		t.Fatalf("length mismatch: %d vs %d", len(decoded), len(*original))
	}
	
	for i, u := range decoded {
		if u.ID != (*original)[i].ID || u.Name != (*original)[i].Name {
			t.Errorf("user[%d]: expected %+v, got %+v", i, (*original)[i], u)
		}
	}
}

func BenchmarkMarshal(b *testing.B) {
	users := &[]User{
		{ID: 1, Name: "Alice"},
		{ID: 2, Name: "Bob"},
		{ID: 3, Name: "Charlie"},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(users)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUnmarshal(b *testing.B) {
	data := []byte("user[3]{id,name}:1,Alice\n2,Bob\n3,Charlie")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var users []User
		err := Unmarshal(data, &users)
		if err != nil {
			b.Fatal(err)
		}
	}
}
