package spine

import (
	"fmt"
	"sort"
)

// Store is a standalone key-value metadata store with pagination and schema validation.
type Store struct {
	entries map[string]any
	schema  Schema
}

// Entry represents a single key-value pair in a Store.
type Entry struct {
	Key   string
	Value any
}

// Page represents a paginated view of store entries.
type Page struct {
	Items   []Entry
	Total   int
	Offset  int
	Limit   int
	HasMore bool
}

// FieldType defines the expected type for a schema field.
type FieldType string

const (
	FieldString FieldType = "string"
	FieldInt    FieldType = "int"
	FieldFloat  FieldType = "float"
	FieldBool   FieldType = "bool"
	FieldBytes  FieldType = "bytes"
	FieldSlice  FieldType = "slice"
	FieldMap    FieldType = "map"
	FieldAny   FieldType = "any"
)

// FieldDef defines the type and requirement for a schema field.
type FieldDef struct {
	Type     FieldType `json:"type"`
	Required bool      `json:"required"`
}

// Schema maps field names to their definitions for store validation.
type Schema map[string]FieldDef

// NewStore creates a new empty Store.
func NewStore() *Store {
	return &Store{entries: make(map[string]any)}
}

// Set adds or updates a key-value pair.
func (s *Store) Set(key string, value any) {
	s.entries[key] = value
}

// Get returns the value for the given key and whether it exists.
func (s *Store) Get(key string) (any, bool) {
	v, ok := s.entries[key]
	return v, ok
}

// Delete removes a key. Returns true if the key existed.
func (s *Store) Delete(key string) bool {
	_, ok := s.entries[key]
	if ok {
		delete(s.entries, key)
	}
	return ok
}

// Has returns true if the key exists.
func (s *Store) Has(key string) bool {
	_, ok := s.entries[key]
	return ok
}

// Len returns the number of entries.
func (s *Store) Len() int {
	return len(s.entries)
}

// Keys returns all keys in sorted order.
func (s *Store) Keys() []string {
	keys := make([]string, 0, len(s.entries))
	for k := range s.entries {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Clear removes all entries.
func (s *Store) Clear() {
	s.entries = make(map[string]any)
}

// List returns a paginated view of store entries sorted by key.
// If limit <= 0, all entries from offset onward are returned.
func (s *Store) List(offset, limit int) Page {
	keys := s.Keys()
	total := len(keys)

	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}

	remaining := keys[offset:]

	var selected []string
	if limit <= 0 {
		selected = remaining
	} else if limit > len(remaining) {
		selected = remaining
	} else {
		selected = remaining[:limit]
	}

	items := make([]Entry, len(selected))
	for i, k := range selected {
		items[i] = Entry{Key: k, Value: s.entries[k]}
	}

	hasMore := offset+len(selected) < total

	return Page{
		Items:   items,
		Total:   total,
		Offset:  offset,
		Limit:   limit,
		HasMore: hasMore,
	}
}

// Range iterates over entries in sorted key order.
// If fn returns false, iteration stops.
func (s *Store) Range(fn func(key string, value any) bool) {
	keys := s.Keys()
	for _, k := range keys {
		if !fn(k, s.entries[k]) {
			return
		}
	}
}

// SetSchema attaches a validation schema to this store.
func (s *Store) SetSchema(schema Schema) {
	s.schema = schema
}

// GetSchema returns the current schema, or nil if none is set.
func (s *Store) GetSchema() Schema {
	return s.schema
}

// Validate checks all entries against the schema.
// Returns nil if no schema is set or all entries are valid.
func (s *Store) Validate() []error {
	if s.schema == nil {
		return nil
	}

	var errs []error

	// Check required fields and type constraints.
	keys := make([]string, 0, len(s.schema))
	for k := range s.schema {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		def := s.schema[key]
		val, exists := s.entries[key]

		if !exists {
			if def.Required {
				errs = append(errs, fmt.Errorf("missing required field %q", key))
			}
			continue
		}

		if def.Type == FieldAny {
			continue
		}

		if !matchesType(val, def.Type) {
			errs = append(errs, fmt.Errorf("field %q: expected type %s, got %T", key, def.Type, val))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errs
}

func matchesType(val any, ft FieldType) bool {
	switch ft {
	case FieldString:
		_, ok := val.(string)
		return ok
	case FieldInt:
		switch val.(type) {
		case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
			return true
		}
		return false
	case FieldFloat:
		switch val.(type) {
		case float32, float64:
			return true
		}
		return false
	case FieldBool:
		_, ok := val.(bool)
		return ok
	case FieldBytes:
		_, ok := val.([]byte)
		return ok
	case FieldSlice:
		switch val.(type) {
		case []any, []string, []int, []float64, []byte, []bool:
			return true
		}
		return false
	case FieldMap:
		switch val.(type) {
		case map[string]any, map[string]string, map[string]int:
			return true
		}
		return false
	case FieldAny:
		return true
	}
	return false
}

// Copy returns a structural copy of the store. Values are shallow-copied.
func (s *Store) Copy() *Store {
	c := NewStore()
	for k, v := range s.entries {
		c.entries[k] = v
	}
	if s.schema != nil {
		sc := make(Schema, len(s.schema))
		for k, v := range s.schema {
			sc[k] = v
		}
		c.schema = sc
	}
	return c
}
