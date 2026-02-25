package toon

import (
	"reflect"
	"strings"
	"sync"
)

// structInfo holds cached metadata for a struct type
type structInfo struct {
	name   string
	fields []fieldInfo
}

// fieldInfo holds metadata for a single field
type fieldInfo struct {
	name  string
	index int
}

// cache uses sync.Map for better concurrent performance
var cache sync.Map // map[reflect.Type]*structInfo

// getStructInfo returns cached struct info for a type
func getStructInfo(t reflect.Type) *structInfo {
	if info, ok := cache.Load(t); ok {
		return info.(*structInfo)
	}

	// Build info
	info := buildStructInfo(t)
	
	// Store in cache (if another goroutine stored first, use that)
	if actual, loaded := cache.LoadOrStore(t, info); loaded {
		return actual.(*structInfo)
	}
	return info
}

// buildStructInfo creates structInfo from reflect.Type
func buildStructInfo(t reflect.Type) *structInfo {
	info := &structInfo{
		name:   strings.ToLower(t.Name()),
		fields: make([]fieldInfo, 0, t.NumField()),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Parse tag
		name := field.Name
		tag := field.Tag.Get("toon")
		if tag == "-" {
			continue // Skip field
		}
		if tag != "" {
			name = tag
		} else {
			name = strings.ToLower(field.Name)
		}

		info.fields = append(info.fields, fieldInfo{
			name:  name,
			index: i,
		})
	}

	return info
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

// structCache is kept for backward compatibility with decoder
type structCache struct{}

var defaultCache = &structCache{}

type fieldMap map[string]int
