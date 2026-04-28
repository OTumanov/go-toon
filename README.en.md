# go-toon

[Русская версия](README.ru.md)

**TOON (Token-Oriented Object Notation)** is a high-performance TOON implementation for Go. It is designed for LLM-heavy workloads (ChatGPT, Claude, etc.), where payload size directly affects cost.

## Why TOON over JSON?

- **Fewer tokens**: header/body layout can reduce LLM context size significantly.
- **High throughput**: generated code provides very fast encode/decode paths.
- **Schema validation**: generated unmarshaling validates header schema via hash.

## Installation

```bash
go get github.com/OTumanov/go-toon
go install github.com/OTumanov/go-toon/cmd/toongen@latest
```

## TOON format at a glance

Struct example:

```text
user{id,name,age}:42,Alice,30
```

Slice example:

```text
user[2]{id,name,age}:1,Alice,30
2,Bob,25
```

## Quick start (runtime reflection API)

```go
package main

import (
	"fmt"

	"github.com/OTumanov/go-toon"
)

type User struct {
	ID   int    `toon:"id"`
	Name string `toon:"name"`
	Age  int    `toon:"age"`
}

func main() {
	src := &User{ID: 42, Name: "Alice", Age: 30}
	data, err := toon.Marshal(src)
	if err != nil {
		panic(err)
	}

	var dst User
	if err := toon.Unmarshal(data, &dst); err != nil {
		panic(err)
	}

	fmt.Println(string(data), dst.Name)
}
```

## Quick start (code generation)

1. Mark structs with `//toon:generate`:

```go
//toon:generate
type User struct {
	ID   int    `toon:"id"`
	Name string `toon:"name"`
	Age  int    `toon:"age"`
}
```

2. Generate code:

```bash
toongen -i ./example -o ./example/toon_gen.go
```

3. Use generated methods:

```go
u := User{ID: 1, Name: "Alice", Age: 30}

data, _ := u.MarshalTOON()
tokens := u.ToonTokenCount()

var decoded User
_ = decoded.UnmarshalTOON(data)
```

## Streaming API

```go
enc := toon.NewEncoder(writer)
_ = enc.Encode(users)

dec := toon.NewDecoder(reader)
var out []User
_ = dec.Decode(&out)
```

## Custom field encoding

If a field type implements `Marshaler` / `Unmarshaler`, it is encoded/decoded through those interfaces automatically.

## Benchmarks (reflect vs generated)

Numbers from the current project README:

| Operation | Time | Allocations |
| --- | --- | --- |
| JSON Marshal | 276 ns/op | 2 allocs |
| TOON Reflect | 163 ns/op | 2 allocs |
| TOON Generated | **42 ns/op** | **0 allocs** |

## License

MIT
