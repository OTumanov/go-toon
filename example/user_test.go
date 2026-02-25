package example

import (
	"testing"
)

func TestUserTokenCount(t *testing.T) {
	u := User{ID: 123, Name: "Alice", Age: 30}
	count := u.ToonTokenCount()
	// Expected: 3 separators + 4 overhead + 3 (ID) + 6/4 (Name) + 2 (Age) = ~13
	if count < 10 || count > 20 {
		t.Errorf("unexpected token count: %d", count)
	}
}

func TestUserMarshalUnmarshal(t *testing.T) {
	original := User{ID: 42, Name: "Bob", Age: 25}
	
	data, err := original.MarshalTOON()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	
	expected := "user{id,name,age}:42,Bob,25"
	if string(data) != expected {
		t.Errorf("expected %q, got %q", expected, string(data))
	}
	
	var decoded User
	err = decoded.UnmarshalTOON(data)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	
	if decoded.ID != 42 || decoded.Name != "Bob" || decoded.Age != 25 {
		t.Errorf("decoded mismatch: %+v", decoded)
	}
}

func TestProductMarshalUnmarshal(t *testing.T) {
	original := Product{SKU: "ABC-123", Price: 99.99}
	
	data, err := original.MarshalTOON()
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	
	var decoded Product
	err = decoded.UnmarshalTOON(data)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	
	if decoded.SKU != "ABC-123" {
		t.Errorf("SKU mismatch: %s", decoded.SKU)
	}
	// Float comparison with epsilon
	if decoded.Price < 99.98 || decoded.Price > 100.0 {
		t.Errorf("Price mismatch: %f", decoded.Price)
	}
}

func BenchmarkUserMarshal(b *testing.B) {
	u := User{ID: 42, Name: "Alice", Age: 30}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := u.MarshalTOON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUserUnmarshal(b *testing.B) {
	data := []byte("user{id,name,age}:42,Alice,30")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var u User
		err := u.UnmarshalTOON(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkUserTokenCount(b *testing.B) {
	u := User{ID: 42, Name: "Alice", Age: 30}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = u.ToonTokenCount()
	}
}
