# go-toon

[![CI](https://github.com/OTumanov/go-toon/actions/workflows/ci.yml/badge.svg)](https://github.com/OTumanov/go-toon/actions/workflows/ci.yml)
[![Spec Fixtures](https://github.com/OTumanov/go-toon/actions/workflows/spec-fixtures.yml/badge.svg)](https://github.com/OTumanov/go-toon/actions/workflows/spec-fixtures.yml)

[Русская версия](README.ru.md)

**TOON (Token-Oriented Object Notation)** is a high-performance TOON implementation for Go. It is designed for LLM-heavy workloads (ChatGPT, Claude, etc.), where payload size directly affects cost.

## Official TOON links

- TOON ecosystem: [github.com/toon-format](https://github.com/toon-format)
- Official Go implementation: [github.com/toon-format/toon-go](https://github.com/toon-format/toon-go)
- Official spec tests: [github.com/toon-format/spec/tree/main/tests](https://github.com/toon-format/spec/tree/main/tests)

## Spec compatibility status

`go-toon` tracks TOON spec fixture integration in CI using the official
`toon-format/spec/tests` dataset.

Current status: fixture ingestion is automated; full behavioral conformance
assertions are being implemented incrementally.

See progress tracker: [`SPEC_COMPATIBILITY.md`](SPEC_COMPATIBILITY.md)

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

## Working with existing JSON tags

`go-toon` intentionally does not enable JSON-tag fallback by default. This keeps schema behavior explicit and predictable for generated and reflection-based paths.

If your models already use `json` tags, prefer one of these explicit approaches:

1. Duplicate tags on shared models:

```go
type User struct {
	ID   int    `json:"id" toon:"id"`
	Name string `json:"name" toon:"name"`
}
```

2. Use an adapter type for external structs you cannot modify:

```go
type ExternalUser struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

//toon:generate
type ToonUser struct {
	ID   int    `toon:"id"`
	Name string `toon:"name"`
}

func NewToonUser(v ExternalUser) ToonUser {
	return ToonUser{ID: v.ID, Name: v.Name}
}
```

3. Implement custom `Marshaler` / `Unmarshaler` for edge cases:

```go
type JSONCompatUser struct {
	ID   int
	Name string
}

func (u JSONCompatUser) MarshalTOON() ([]byte, error) {
	// map fields manually the way your compatibility layer requires
	return []byte("user{id,name}:1,Alice"), nil
}

func (u *JSONCompatUser) UnmarshalTOON(data []byte) error {
	// parse according to your custom compatibility contract
	u.ID, u.Name = 1, "Alice"
	return nil
}
```

This direction preserves `go-toon` performance and avoids surprising behavior changes from implicit tag inference.

## Benchmarks (reflect vs generated)

Numbers from the current project README:

| Operation | Time | Allocations |
| --- | --- | --- |
| JSON Marshal | 276 ns/op | 2 allocs |
| TOON Reflect | 163 ns/op | 2 allocs |
| TOON Generated | **42 ns/op** | **0 allocs** |

## Comparison with alternatives

`toon-format/toon-go` (official implementation) provides a solid and convenient API, and is a great default for general use.

If your primary goal is maximum throughput and minimal allocations, generated methods in `go-toon` can be a better fit:

- Official reflect-based path: around **163 ns/op**, **2 allocs**
- `go-toon` generated path: around **42 ns/op**, **0 allocs**

This is roughly a **4x speedup** with **zero allocations** on the generated path.

## License

MIT
