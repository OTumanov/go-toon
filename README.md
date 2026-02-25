# go-toon

**TOON (Token-Oriented Object Notation)** — высокопроизводительная реализация формата TOON для Go. Оптимизировано для работы с LLM (ChatGPT, Claude), где размер контекста напрямую конвертируется в деньги.

## Почему это лучше JSON?
- **Экономия токенов:** До 40-50% меньше объем данных за счет Header-Body структуры.
- **Производительность:** Кодогенерация обеспечивает Zero-allocation маршалинг.
- **Безопасность:** Compile-time проверка схем через хеширование заголовков.

## Установка
```bash
go get github.com/OTumanov/go-toon
go install github.com/OTumanov/go-toon/cmd/toongen@latest
```

## Быстрый старт

1. Добавьте тег `//toon:generate` к вашей структуре:

```go
//toon:generate
type User struct {
    ID     int    `toon:"id"`
    Name   string `toon:"name"`
    Active bool   `toon:"active"`
}
```

2. Запустите генератор:

```bash
go generate ./...
```

3. Используйте в коде:

```go
u := User{ID: 1, Name: "Alice", Active: true}

// Маршалинг (Zero-allocation)
data, _ := u.MarshalTOON() 

// Оценка токенов для LLM
tokens := u.ToonTokenCount() 
```

## Бенчмарки (Reflect vs Generated)

| Операция | Время | Аллокации |
| --- | --- | --- |
| **JSON Marshal** | 276 ns/op | 2 allocs |
| **TOON Reflect** | 163 ns/op | 2 allocs |
| **TOON Generated** | **42 ns/op** | **0 allocs** |

## Архитектура

Проект следует строгому **ПРАВИЛУ СЛОЕВ**:

- Инфраструктурный слой генерации и парсинга полностью изолирован.
- Поддержка `io.Reader/io.Writer` для потоковой передачи в LLM.
- Возможность кастомного маршалинга через интерфейсы `Marshaler/Unmarshaler`.

## Лицензия

MIT
