package toon

import (
	"reflect"
	"strings"
	"sync"
)

// structInfo holds cached metadata for a struct type
type structInfo struct {
	name      string
	nameBytes []byte
	fields    []fieldInfo
}

// fieldInfo holds metadata for a single field
type fieldInfo struct {
	name      string
	nameBytes []byte // For fast []byte comparison
	index     int
}

// cache uses sync.Map for better concurrent performance
var cache sync.Map // map[reflect.Type]*structInfo

// getStructInfo returns cached struct info for a type
func getStructInfo(t reflect.Type) *structInfo {
	if info, ok := cache.Load(t); ok {
		return info.(*structInfo)
	}

	info := buildStructInfo(t)
	if actual, loaded := cache.LoadOrStore(t, info); loaded {
		return actual.(*structInfo)
	}
	return info
}

// buildStructInfo creates structInfo from reflect.Type
func buildStructInfo(t reflect.Type) *structInfo {
	name := strings.ToLower(t.Name())
	info := &structInfo{
		name:      name,
		nameBytes: []byte(name),
		fields:    make([]fieldInfo, 0, t.NumField()),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		if !field.IsExported() {
			continue
		}

		name := field.Name
		tag := field.Tag.Get("toon")
		if tag == "-" {
			continue
		}
		if tag != "" {
			name = tag
		} else {
			name = strings.ToLower(field.Name)
		}

		info.fields = append(info.fields, fieldInfo{
			name:      name,
			nameBytes: []byte(name),
			index:     i,
		})
	}

	return info
}

// findFieldIndex finds field index by []byte name (zero-copy lookup)
func (info *structInfo) findFieldIndex(name []byte) int {
	for _, f := range info.fields {
		if bytesEqual(f.nameBytes, name) {
			return f.index
		}
	}
	return -1
}

// bytesEqual compares two byte slices (faster than bytes.Equal for small slices)
func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// Legacy support for decoder.go
func (c *structCache) get(t reflect.Type) fieldMap {
	info := getStructInfo(t)
	fm := make(fieldMap)
	for _, f := range info.fields {
		fm[f.name] = f.index
	}
	return fm
}

type structCache struct{}
var defaultCache = &structCache{}
type fieldMap map[string]int
