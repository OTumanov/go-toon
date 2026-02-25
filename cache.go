package toon

import (
	"reflect"
	"sync"
)

// structCache maps struct types to their field indices for fast lookup
type structCache struct {
	mu    sync.RWMutex
	cache map[reflect.Type]fieldMap
}

// fieldMap maps field names to their struct index
type fieldMap map[string]int

var defaultCache = &structCache{
	cache: make(map[reflect.Type]fieldMap),
}

// get returns cached field map for a struct type
func (c *structCache) get(t reflect.Type) fieldMap {
	c.mu.RLock()
	fm, ok := c.cache[t]
	c.mu.RUnlock()
	
	if ok {
		return fm
	}
	
	// Compute and cache
	c.mu.Lock()
	defer c.mu.Unlock()
	
	// Double-check after acquiring write lock
	if fm, ok := c.cache[t]; ok {
		return fm
	}
	
	fm = buildFieldMap(t)
	c.cache[t] = fm
	return fm
}

// buildFieldMap creates a mapping from field name to struct index
func buildFieldMap(t reflect.Type) fieldMap {
	fm := make(fieldMap)
	
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		
		// Skip unexported fields
		if !field.IsExported() {
			continue
		}
		
		// Use "toon" tag if present, otherwise use field name
		name := field.Name
		if tag := field.Tag.Get("toon"); tag != "" {
			name = tag
		}
		
		fm[name] = i
	}
	
	return fm
}
