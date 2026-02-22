package gindocs

import (
	"reflect"
	"strconv"
	"strings"
	"time"
)

// typeToSchema converts a Go reflect.Type to an OpenAPI SchemaObject.
// It registers struct types in the registry and returns $ref for known types.
func typeToSchema(t reflect.Type, registry *TypeRegistry) *SchemaObject {
	// Dereference pointers.
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Handle special types first.
	if schema := specialTypeSchema(t); schema != nil {
		return schema
	}

	switch t.Kind() {
	case reflect.Bool:
		return &SchemaObject{Type: "boolean"}

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		return &SchemaObject{Type: "integer", Format: "int32"}

	case reflect.Int64, reflect.Uint64:
		return &SchemaObject{Type: "integer", Format: "int64"}

	case reflect.Float32:
		return &SchemaObject{Type: "number", Format: "float"}

	case reflect.Float64:
		return &SchemaObject{Type: "number", Format: "double"}

	case reflect.String:
		return &SchemaObject{Type: "string"}

	case reflect.Slice, reflect.Array:
		elemSchema := typeToSchema(t.Elem(), registry)
		// []byte is a string (base64)
		if t.Elem().Kind() == reflect.Uint8 {
			return &SchemaObject{Type: "string", Format: "byte"}
		}
		return &SchemaObject{
			Type:  "array",
			Items: elemSchema,
		}

	case reflect.Map:
		valSchema := typeToSchema(t.Elem(), registry)
		return &SchemaObject{
			Type:                 "object",
			AdditionalProperties: valSchema,
		}

	case reflect.Struct:
		return structToSchema(t, registry)

	case reflect.Interface:
		// interface{} / any
		return &SchemaObject{}

	default:
		return &SchemaObject{Type: "string"}
	}
}

// specialTypeSchema handles well-known types that need special treatment.
func specialTypeSchema(t reflect.Type) *SchemaObject {
	// time.Time → string with date-time format.
	if t == reflect.TypeOf(time.Time{}) {
		return &SchemaObject{Type: "string", Format: "date-time"}
	}

	// Check for types that implement encoding.TextMarshaler (they serialize as strings).
	textMarshalerType := reflect.TypeOf((*interface{ MarshalText() ([]byte, error) })(nil)).Elem()
	if t.Implements(textMarshalerType) || reflect.PtrTo(t).Implements(textMarshalerType) {
		return &SchemaObject{Type: "string"}
	}

	return nil
}

// structToSchema converts a struct type to an OpenAPI SchemaObject.
// Registers the struct in the registry and returns a $ref.
func structToSchema(t reflect.Type, registry *TypeRegistry) *SchemaObject {
	// Dereference pointers.
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	name := schemaName(t)

	// If already registered, return a $ref.
	if registry.Has(name) {
		return SchemaRef(name)
	}

	// Check for circular references.
	if registry.isSeen(t) {
		return SchemaRef(name)
	}

	// Mark as being processed.
	registry.markSeen(t)
	defer registry.unmarkSeen(t)

	schema := &SchemaObject{
		Type:       "object",
		Properties: make(map[string]*SchemaObject),
	}

	// Process all fields including embedded structs.
	processStructFields(t, schema, registry)

	// Register the schema.
	registry.Register(name, schema)

	return SchemaRef(name)
}

// processStructFields processes struct fields, handling embedded structs recursively.
func processStructFields(t reflect.Type, schema *SchemaObject, registry *TypeRegistry) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields.
		if !field.IsExported() {
			continue
		}

		// Handle embedded structs.
		if field.Anonymous {
			embeddedType := field.Type
			for embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}
			if embeddedType.Kind() == reflect.Struct {
				// Check if it's a special type (like time.Time).
				if specialTypeSchema(embeddedType) == nil {
					processStructFields(embeddedType, schema, registry)
					continue
				}
			}
		}

		// Parse all tags.
		tagInfo := mergeTags(
			field.Tag.Get("json"),
			field.Tag.Get("binding"),
			field.Tag.Get("gorm"),
			field.Tag.Get("docs"),
		)

		// Skip hidden or skipped fields.
		if tagInfo.JSONSkip || tagInfo.GORMSkip || tagInfo.Hidden {
			continue
		}

		// Determine property name.
		propName := tagInfo.JSONName
		if propName == "" {
			propName = field.Name
		}

		// Generate schema for the field type.
		fieldSchema := fieldToSchema(field.Type, tagInfo, registry)

		schema.Properties[propName] = fieldSchema

		// Add to required list.
		if tagInfo.Required {
			schema.Required = append(schema.Required, propName)
		}
	}
}

// fieldToSchema generates a schema for a struct field, applying tag constraints.
func fieldToSchema(t reflect.Type, tags TagInfo, registry *TypeRegistry) *SchemaObject {
	// Get the base schema from the type.
	baseSchema := typeToSchema(t, registry)

	// If it's a $ref, we can't add constraints directly.
	// We need to use the base schema as-is.
	if baseSchema.Ref != "" {
		// Apply description via wrapper if needed.
		if tags.Description != "" || tags.Deprecated {
			return &SchemaObject{
				AllOf:       []*SchemaObject{baseSchema},
				Description: tags.Description,
				Deprecated:  tags.Deprecated,
			}
		}
		return baseSchema
	}

	// Apply tag constraints to the schema.
	applyTagConstraints(baseSchema, tags)

	return baseSchema
}

// applyTagConstraints applies parsed tag information to a schema.
func applyTagConstraints(schema *SchemaObject, tags TagInfo) {
	// Description.
	if tags.Description != "" {
		schema.Description = tags.Description
	}

	// Add "Must be unique" to description if unique index.
	if tags.UniqueIndex {
		if schema.Description != "" {
			schema.Description += ". Must be unique"
		} else {
			schema.Description = "Must be unique"
		}
	}

	// Format.
	if tags.Format != "" {
		schema.Format = tags.Format
	}

	// Enum.
	if len(tags.Enum) > 0 {
		for _, v := range tags.Enum {
			schema.Enum = append(schema.Enum, v)
		}
	}

	// Numeric constraints — only apply to number/integer types.
	if schema.Type == "integer" || schema.Type == "number" {
		schema.Minimum = tags.Minimum
		schema.Maximum = tags.Maximum
	}

	// String constraints — only apply to string types.
	if schema.Type == "string" {
		schema.MinLength = tags.MinLength
		schema.MaxLength = tags.MaxLength

		// GORM size as maxLength.
		if tags.GORMSize != nil && schema.MaxLength == nil {
			schema.MaxLength = tags.GORMSize
		}
	}

	// Default value.
	if tags.GORMDefault != nil {
		schema.Default = parseDefaultValue(*tags.GORMDefault, schema.Type)
	}

	// ReadOnly for primary keys and auto-timestamps.
	if tags.PrimaryKey || tags.AutoCreateTime || tags.AutoUpdateTime {
		schema.ReadOnly = true
	}

	// Deprecated.
	if tags.Deprecated {
		schema.Deprecated = true
	}

	// Example.
	if tags.Example != "" {
		schema.Example = parseExampleValue(tags.Example, schema.Type)
	}
}

// parseDefaultValue converts a string default to the appropriate Go type.
func parseDefaultValue(val, schemaType string) interface{} {
	switch schemaType {
	case "integer":
		if v, err := strconv.ParseInt(val, 10, 64); err == nil {
			return v
		}
	case "number":
		if v, err := strconv.ParseFloat(val, 64); err == nil {
			return v
		}
	case "boolean":
		if v, err := strconv.ParseBool(val); err == nil {
			return v
		}
	}
	return val
}

// parseExampleValue converts a string example to the appropriate Go type.
func parseExampleValue(val, schemaType string) interface{} {
	switch schemaType {
	case "integer":
		if v, err := strconv.ParseInt(val, 10, 64); err == nil {
			return v
		}
	case "number":
		if v, err := strconv.ParseFloat(val, 64); err == nil {
			return v
		}
	case "boolean":
		if v, err := strconv.ParseBool(val); err == nil {
			return v
		}
	}
	return val
}

// TypeOf is a helper that returns the reflect.Type for a value, useful for
// registering types without creating instances.
func TypeOf(v interface{}) reflect.Type {
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t
}

// SchemaFromType generates an OpenAPI schema for a Go value and registers it.
func SchemaFromType(v interface{}, registry *TypeRegistry) *SchemaObject {
	t := reflect.TypeOf(v)
	return typeToSchema(t, registry)
}

// isStringType checks if a schema type represents string-like values.
func isStringType(fieldName string) bool {
	lower := strings.ToLower(fieldName)
	return strings.Contains(lower, "name") ||
		strings.Contains(lower, "email") ||
		strings.Contains(lower, "title") ||
		strings.Contains(lower, "description") ||
		strings.Contains(lower, "content") ||
		strings.Contains(lower, "body") ||
		strings.Contains(lower, "bio")
}
