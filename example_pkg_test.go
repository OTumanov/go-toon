package toon

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

type exampleUser struct {
	ID   int    `toon:"id"`
	Name string `toon:"name"`
}

func ExampleMarshal() {
	u := &exampleUser{ID: 42, Name: "Alice"}
	data, err := Marshal(u)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
	// Output: exampleuser{id,name}:42,Alice
}

func ExampleUnmarshal() {
	input := []byte("user{id,name}:42,Alice")
	var u exampleUser
	if err := Unmarshal(input, &u); err != nil {
		panic(err)
	}
	fmt.Printf("%d %s\n", u.ID, u.Name)
	// Output: 42 Alice
}

func ExampleNewEncoder() {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	if err := enc.Encode(&exampleUser{ID: 7, Name: "Bob"}); err != nil {
		panic(err)
	}
	fmt.Println(strings.TrimSpace(buf.String()))
	// Output: exampleuser{id,name}:7,Bob
}

func ExampleNewDecoder() {
	src := strings.NewReader("exampleuser{id,name}:7,Bob\n")
	dec := NewDecoder(src)

	var u exampleUser
	if err := dec.Decode(&u); err != nil {
		panic(err)
	}
	fmt.Printf("%d %s\n", u.ID, u.Name)
	// Output: 7 Bob
}

type exampleHexID int

func (id exampleHexID) MarshalTOON() ([]byte, error) {
	return []byte(strconv.FormatInt(int64(id), 16)), nil
}

func (id *exampleHexID) UnmarshalTOON(data []byte) error {
	n, err := strconv.ParseInt(string(data), 16, 64)
	if err != nil {
		return err
	}
	*id = exampleHexID(n)
	return nil
}

func ExampleMarshaler() {
	type payload struct {
		ID exampleHexID `toon:"id"`
	}

	out, err := Marshal(&payload{ID: 26})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(out))
	// Output: payload{id}:1a
}

func ExampleUnmarshaler() {
	type payload struct {
		ID exampleHexID `toon:"id"`
	}

	var p payload
	if err := Unmarshal([]byte("payload{id}:1a"), &p); err != nil {
		panic(err)
	}
	fmt.Println(int(p.ID))
	// Output: 26
}
