package toon

import "fmt"

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
