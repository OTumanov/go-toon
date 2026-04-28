# go-toon

[![CI](https://github.com/OTumanov/go-toon/actions/workflows/ci.yml/badge.svg)](https://github.com/OTumanov/go-toon/actions/workflows/ci.yml)
[![Spec Fixtures](https://github.com/OTumanov/go-toon/actions/workflows/spec-fixtures.yml/badge.svg)](https://github.com/OTumanov/go-toon/actions/workflows/spec-fixtures.yml)

[English version](README.en.md)

**TOON (Token-Oriented Object Notation)** — высокопроизводительная реализация формата TOON для Go. Библиотека ориентирована на сценарии с LLM (ChatGPT, Claude и т.д.), где объем контекста напрямую влияет на стоимость.

## Официальные ссылки TOON

- Экосистема TOON: [github.com/toon-format](https://github.com/toon-format)
- Официальная Go-реализация: [github.com/toon-format/toon-go](https://github.com/toon-format/toon-go)
- Официальные spec-тесты: [github.com/toon-format/spec/tree/main/tests](https://github.com/toon-format/spec/tree/main/tests)

## Статус совместимости со спецификацией

`go-toon` подключает в CI официальный набор `toon-format/spec/tests`.

Текущий статус: интеграция fixtures автоматизирована; полные поведенческие
проверки соответствия добавляются поэтапно.

Трекер прогресса: [`SPEC_COMPATIBILITY.md`](SPEC_COMPATIBILITY.md)

## Почему TOON вместо JSON?

- **Меньше токенов**: формат header/body помогает заметно уменьшить размер контекста.
- **Высокая производительность**: кодогенерация дает быстрые пути encode/decode.
- **Проверка схемы**: в сгенерированном коде заголовок проверяется через hash.

## Установка

```bash
go get github.com/OTumanov/go-toon
go install github.com/OTumanov/go-toon/cmd/toongen@latest
```

## Формат TOON в одном взгляде

Пример структуры:

```text
user{id,name,age}:42,Alice,30
```

Пример слайса:

```text
user[2]{id,name,age}:1,Alice,30
2,Bob,25
```

## Быстрый старт (runtime reflection API)

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

## Быстрый старт (кодогенерация)

1. Пометьте структуры комментарием `//toon:generate`:

```go
//toon:generate
type User struct {
	ID   int    `toon:"id"`
	Name string `toon:"name"`
	Age  int    `toon:"age"`
}
```

2. Запустите генератор:

```bash
toongen -i ./example -o ./example/toon_gen.go
```

3. Используйте сгенерированные методы:

```go
u := User{ID: 1, Name: "Alice", Age: 30}

data, _ := u.MarshalTOON()
tokens := u.ToonTokenCount()

var decoded User
_ = decoded.UnmarshalTOON(data)
```

## Потоковый API

```go
enc := toon.NewEncoder(writer)
_ = enc.Encode(users)

dec := toon.NewDecoder(reader)
var out []User
_ = dec.Decode(&out)
```

## Кастомное кодирование полей

Если тип поля реализует `Marshaler` / `Unmarshaler`, библиотека автоматически использует эти интерфейсы при кодировании и декодировании.

## Работа с существующими JSON-тегами

В `go-toon` JSON-fallback намеренно не включен по умолчанию. Это сохраняет явное и предсказуемое поведение схемы как для codegen, так и для reflection-пути.

Если ваши модели уже используют `json`-теги, лучше применять один из явных вариантов:

1. Дублировать теги в общих моделях:

```go
type User struct {
	ID   int    `json:"id" toon:"id"`
	Name string `json:"name" toon:"name"`
}
```

2. Использовать адаптерный тип для внешних структур, которые нельзя менять:

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

3. Реализовать кастомные `Marshaler` / `Unmarshaler` для edge-case сценариев:

```go
type JSONCompatUser struct {
	ID   int
	Name string
}

func (u JSONCompatUser) MarshalTOON() ([]byte, error) {
	// вручную отображаем поля по правилам вашего compatibility-слоя
	return []byte("user{id,name}:1,Alice"), nil
}

func (u *JSONCompatUser) UnmarshalTOON(data []byte) error {
	// парсим по вашему кастомному контракту совместимости
	u.ID, u.Name = 1, "Alice"
	return nil
}
```

Такой подход сохраняет преимущества `go-toon` по производительности и исключает неожиданные изменения поведения из-за неявного выбора тегов.

## Бенчмарки (reflect vs generated)

Команда запуска:

```bash
go test -bench BenchmarkCompare -benchmem -run ^$ ./...
```

Окружение: `darwin/arm64`, Apple M4, Go из `go.mod`.

| Marshal | Время | Память | Аллокации |
| --- | --- | --- | --- |
| `encoding/json` | 78.78 ns/op | 48 B/op | 1 allocs/op |
| `json-iterator/go` | 107.4 ns/op | 48 B/op | 1 allocs/op |
| `go-toon` reflect | 118.2 ns/op | 88 B/op | 5 allocs/op |
| `go-toon` generated | **52.72 ns/op** | 256 B/op | 1 allocs/op |

| Unmarshal | Время | Память | Аллокации |
| --- | --- | --- | --- |
| `encoding/json` | 387.5 ns/op | 256 B/op | 6 allocs/op |
| `json-iterator/go` | 128.2 ns/op | 48 B/op | 5 allocs/op |
| `go-toon` reflect | 177.5 ns/op | 269 B/op | 6 allocs/op |
| `go-toon` generated | **41.50 ns/op** | **5 B/op** | **1 allocs/op** |

Код бенчмарков: `benchmark_compare_test.go`.

## Сравнение с альтернативами

`toon-format/toon-go` (официальная реализация) дает удобный и качественный API, и это хороший выбор для большинства сценариев.

Если приоритет — максимальная производительность и минимум аллокаций, кодогенерация в `go-toon` может быть предпочтительнее:

- Официальный путь на reflection: около **118.2 ns/op**, **5 allocs**
- Сгенерированный путь в `go-toon`: около **52.72 ns/op**, **1 allocs**

Самый сильный результат в текущем наборе — generated-декодирование (`41.50 ns/op`, `1 allocs/op`).

## Лицензия

MIT
