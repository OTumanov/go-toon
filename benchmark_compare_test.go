package toon

import (
	"encoding/json"
	"testing"

	jsoniter "github.com/json-iterator/go"

	examplepkg "github.com/Otumanov/go-toon/example"
)

var jsoniterAPI = jsoniter.ConfigCompatibleWithStandardLibrary

func benchmarkUser() examplepkg.User {
	return examplepkg.User{
		ID:   42,
		Name: "Alice",
		Age:  30,
	}
}

func BenchmarkCompareMarshalEncodingJSON(b *testing.B) {
	u := benchmarkUser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := json.Marshal(u)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompareMarshalJSONIter(b *testing.B) {
	u := benchmarkUser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := jsoniterAPI.Marshal(u)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompareMarshalTOONReflect(b *testing.B) {
	u := benchmarkUser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(&u)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompareMarshalTOONGenerated(b *testing.B) {
	u := benchmarkUser()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := u.MarshalTOON()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompareUnmarshalEncodingJSON(b *testing.B) {
	input := []byte(`{"ID":42,"Name":"Alice","Age":30}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var u examplepkg.User
		if err := json.Unmarshal(input, &u); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompareUnmarshalJSONIter(b *testing.B) {
	input := []byte(`{"ID":42,"Name":"Alice","Age":30}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var u examplepkg.User
		if err := jsoniterAPI.Unmarshal(input, &u); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompareUnmarshalTOONReflect(b *testing.B) {
	input := []byte("user{id,name,age}:42,Alice,30")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var u examplepkg.User
		if err := Unmarshal(input, &u); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCompareUnmarshalTOONGenerated(b *testing.B) {
	input := []byte("user{id,name,age}:42,Alice,30")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var u examplepkg.User
		if err := u.UnmarshalTOON(input); err != nil {
			b.Fatal(err)
		}
	}
}
