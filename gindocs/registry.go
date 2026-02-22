package gindocs

import (
	"reflect"
	"sync"
)

// TypeRegistry manages schema deduplication and $ref generation.
type TypeRegistry struct {
	mu      sync.RWMutex
	schemas map[string]*SchemaObject
	// seen tracks types currently being processed (for circular reference detection).
	seen map[reflect.Type]bool
}

// newTypeRegistry creates a new TypeRegistry.
func newTypeRegistry() *TypeRegistry {
	return &TypeRegistry{
		schemas: make(map[string]*SchemaObject),
		seen:    make(map[reflect.Type]bool),
	}
}

// Register adds a schema to the registry under the given name.
func (r *TypeRegistry) Register(name string, schema *SchemaObject) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.schemas[name] = schema
}

// Get retrieves a schema by name.
func (r *TypeRegistry) Get(name string) (*SchemaObject, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.schemas[name]
	return s, ok
}

// Has checks if a schema with the given name exists.
func (r *TypeRegistry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.schemas[name]
	return ok
}

// All returns a copy of all registered schemas.
func (r *TypeRegistry) All() map[string]*SchemaObject {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make(map[string]*SchemaObject, len(r.schemas))
	for k, v := range r.schemas {
		result[k] = v
	}
	return result
}

// RefPath returns the OpenAPI $ref path for a named schema.
func RefPath(name string) string {
	return "#/components/schemas/" + name
}

// SchemaRef returns a SchemaObject that is a $ref to a named component.
func SchemaRef(name string) *SchemaObject {
	return &SchemaObject{
		Ref: RefPath(name),
	}
}

// markSeen marks a type as currently being processed (for circular ref detection).
func (r *TypeRegistry) markSeen(t reflect.Type) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seen[t] = true
}

// unmarkSeen removes a type from the processing set.
func (r *TypeRegistry) unmarkSeen(t reflect.Type) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.seen, t)
}

// isSeen checks if a type is currently being processed.
func (r *TypeRegistry) isSeen(t reflect.Type) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.seen[t]
}

// schemaName generates a schema name from a reflect.Type.
func schemaName(t reflect.Type) string {
	// Dereference pointers.
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := t.Name()
	if name == "" {
		// Anonymous struct — use a generated name.
		return "AnonymousStruct"
	}

	return name
}
