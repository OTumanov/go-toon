# go-toon

[English version](README.en.md)

**TOON (Token-Oriented Object Notation)** — высокопроизводительная реализация формата TOON для Go. Библиотека ориентирована на сценарии с LLM (ChatGPT, Claude и т.д.), где объем контекста напрямую влияет на стоимость.

## Официальные ссылки TOON

- Экосистема TOON: [github.com/toon-format](https://github.com/toon-format)
- Официальная Go-реализация: [github.com/toon-format/toon-go](https://github.com/toon-format/toon-go)

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

## Бенчмарки (reflect vs generated)

Числа из текущего README проекта:

| Операция | Время | Аллокации |
| --- | --- | --- |
| JSON Marshal | 276 ns/op | 2 allocs |
| TOON Reflect | 163 ns/op | 2 allocs |
| TOON Generated | **42 ns/op** | **0 allocs** |

## Сравнение с альтернативами

`toon-format/toon-go` (официальная реализация) дает удобный и качественный API, и это хороший выбор для большинства сценариев.

Если приоритет — максимальная производительность и минимум аллокаций, кодогенерация в `go-toon` может быть предпочтительнее:

- Официальный путь на reflection: около **163 ns/op**, **2 allocs**
- Сгенерированный путь в `go-toon`: около **42 ns/op**, **0 allocs**

Это примерно **в 4 раза быстрее** и **без выделений памяти** на generated-пути.

## Лицензия

MIT
