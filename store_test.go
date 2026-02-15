package spine

import (
	"testing"
)

func TestStoreSetAndGet(t *testing.T) {
	s := NewStore()
	s.Set("name", "alice")
	v, ok := s.Get("name")
	if !ok || v != "alice" {
		t.Fatalf("expected alice, got %v", v)
	}

	// Overwrite
	s.Set("name", "bob")
	v, ok = s.Get("name")
	if !ok || v != "bob" {
		t.Fatalf("expected bob after overwrite, got %v", v)
	}

	// Missing key
	_, ok = s.Get("missing")
	if ok {
		t.Fatal("expected missing key to return false")
	}
}

func TestStoreDelete(t *testing.T) {
	s := NewStore()
	s.Set("a", 1)

	if !s.Delete("a") {
		t.Fatal("expected delete to return true for existing key")
	}
	if s.Has("a") {
		t.Fatal("key should be gone after delete")
	}
	if s.Delete("a") {
		t.Fatal("expected delete to return false for missing key")
	}
}

func TestStoreHas(t *testing.T) {
	s := NewStore()
	if s.Has("x") {
		t.Fatal("expected Has to return false for empty store")
	}
	s.Set("x", 42)
	if !s.Has("x") {
		t.Fatal("expected Has to return true after Set")
	}
}

func TestStoreLen(t *testing.T) {
	s := NewStore()
	if s.Len() != 0 {
		t.Fatalf("expected 0, got %d", s.Len())
	}
	s.Set("a", 1)
	s.Set("b", 2)
	if s.Len() != 2 {
		t.Fatalf("expected 2, got %d", s.Len())
	}
	s.Delete("a")
	if s.Len() != 1 {
		t.Fatalf("expected 1, got %d", s.Len())
	}
}

func TestStoreKeys(t *testing.T) {
	s := NewStore()
	s.Set("c", 3)
	s.Set("a", 1)
	s.Set("b", 2)
	keys := s.Keys()
	if len(keys) != 3 || keys[0] != "a" || keys[1] != "b" || keys[2] != "c" {
		t.Fatalf("expected [a b c], got %v", keys)
	}
}

func TestStoreClear(t *testing.T) {
	s := NewStore()
	s.Set("a", 1)
	s.Set("b", 2)
	s.Clear()
	if s.Len() != 0 {
		t.Fatalf("expected 0 after clear, got %d", s.Len())
	}
}

func TestStoreListPagination(t *testing.T) {
	s := NewStore()
	s.Set("a", 1)
	s.Set("b", 2)
	s.Set("c", 3)
	s.Set("d", 4)
	s.Set("e", 5)

	// First page
	p := s.List(0, 2)
	if len(p.Items) != 2 || p.Total != 5 || p.Offset != 0 || !p.HasMore {
		t.Fatalf("unexpected first page: %+v", p)
	}
	if p.Items[0].Key != "a" || p.Items[1].Key != "b" {
		t.Fatalf("unexpected keys: %v %v", p.Items[0].Key, p.Items[1].Key)
	}

	// Second page
	p = s.List(2, 2)
	if len(p.Items) != 2 || p.Offset != 2 || !p.HasMore {
		t.Fatalf("unexpected second page: %+v", p)
	}
	if p.Items[0].Key != "c" || p.Items[1].Key != "d" {
		t.Fatalf("unexpected keys: %v %v", p.Items[0].Key, p.Items[1].Key)
	}

	// Last page
	p = s.List(4, 2)
	if len(p.Items) != 1 || p.HasMore {
		t.Fatalf("unexpected last page: %+v", p)
	}
	if p.Items[0].Key != "e" {
		t.Fatalf("unexpected key: %v", p.Items[0].Key)
	}

	// Offset beyond total
	p = s.List(10, 2)
	if len(p.Items) != 0 || p.HasMore {
		t.Fatalf("expected empty page for offset beyond total: %+v", p)
	}
}

func TestStoreListAll(t *testing.T) {
	s := NewStore()
	s.Set("x", 1)
	s.Set("y", 2)
	s.Set("z", 3)

	p := s.List(0, 0)
	if len(p.Items) != 3 || p.HasMore {
		t.Fatalf("expected all items: %+v", p)
	}
}

func TestStoreRange(t *testing.T) {
	s := NewStore()
	s.Set("a", 1)
	s.Set("b", 2)
	s.Set("c", 3)

	// Full iteration
	var visited []string
	s.Range(func(key string, value any) bool {
		visited = append(visited, key)
		return true
	})
	if len(visited) != 3 || visited[0] != "a" || visited[1] != "b" || visited[2] != "c" {
		t.Fatalf("expected [a b c], got %v", visited)
	}

	// Early stop
	visited = nil
	s.Range(func(key string, value any) bool {
		visited = append(visited, key)
		return key != "b"
	})
	if len(visited) != 2 || visited[0] != "a" || visited[1] != "b" {
		t.Fatalf("expected [a b], got %v", visited)
	}
}

func TestStoreSchema(t *testing.T) {
	s := NewStore()
	if s.GetSchema() != nil {
		t.Fatal("expected nil schema initially")
	}
	schema := Schema{
		"name": {Type: FieldString, Required: true},
	}
	s.SetSchema(schema)
	got := s.GetSchema()
	if got == nil || got["name"].Type != FieldString {
		t.Fatalf("unexpected schema: %v", got)
	}
}

func TestStoreValidateRequired(t *testing.T) {
	s := NewStore()
	s.SetSchema(Schema{
		"name": {Type: FieldString, Required: true},
		"age":  {Type: FieldInt, Required: false},
	})

	errs := s.Validate()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(errs), errs)
	}

	s.Set("name", "alice")
	errs = s.Validate()
	if errs != nil {
		t.Fatalf("expected no errors, got %v", errs)
	}
}

func TestStoreValidateTypes(t *testing.T) {
	s := NewStore()
	s.SetSchema(Schema{
		"s": {Type: FieldString},
		"i": {Type: FieldInt},
		"f": {Type: FieldFloat},
		"b": {Type: FieldBool},
		"y": {Type: FieldBytes},
		"l": {Type: FieldSlice},
		"m": {Type: FieldMap},
		"a": {Type: FieldAny},
	})

	// All correct types
	s.Set("s", "hello")
	s.Set("i", 42)
	s.Set("f", 3.14)
	s.Set("b", true)
	s.Set("y", []byte{1, 2})
	s.Set("l", []string{"a", "b"})
	s.Set("m", map[string]any{"k": "v"})
	s.Set("a", struct{}{})

	errs := s.Validate()
	if errs != nil {
		t.Fatalf("expected no errors with correct types, got %v", errs)
	}

	// Wrong type for string field
	s.Set("s", 123)
	errs = s.Validate()
	if len(errs) != 1 {
		t.Fatalf("expected 1 error for wrong type, got %d: %v", len(errs), errs)
	}

	// Int subtypes
	s.Set("s", "ok")
	s.Set("i", int64(100))
	errs = s.Validate()
	if errs != nil {
		t.Fatalf("expected int64 to match FieldInt, got %v", errs)
	}
}

func TestStoreValidateOpenWorld(t *testing.T) {
	s := NewStore()
	s.SetSchema(Schema{
		"name": {Type: FieldString, Required: true},
	})
	s.Set("name", "alice")
	s.Set("extra", 999)

	errs := s.Validate()
	if errs != nil {
		t.Fatalf("extra keys should be allowed, got %v", errs)
	}
}

func TestStoreValidateNoSchema(t *testing.T) {
	s := NewStore()
	s.Set("anything", "goes")
	errs := s.Validate()
	if errs != nil {
		t.Fatalf("expected nil with no schema, got %v", errs)
	}
}

func TestStoreCopy(t *testing.T) {
	s := NewStore()
	s.Set("a", 1)
	s.Set("b", 2)
	s.SetSchema(Schema{"a": {Type: FieldInt, Required: true}})

	c := s.Copy()

	// Modify original
	s.Set("a", 99)
	s.Delete("b")

	v, ok := c.Get("a")
	if !ok || v != 1 {
		t.Fatalf("copy should be independent, got %v", v)
	}
	if !c.Has("b") {
		t.Fatal("copy should still have key b")
	}
	if c.GetSchema() == nil {
		t.Fatal("copy should have schema")
	}
}
