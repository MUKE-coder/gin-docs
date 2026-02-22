package gindocs

import (
	"reflect"
	"strings"
	"time"
)

// registerGORMModels processes registered GORM models and creates schema variants.
func (gd *GinDocs) registerGORMModels() {
	if len(gd.config.Models) == 0 {
		return
	}

	for _, model := range gd.config.Models {
		t := reflect.TypeOf(model)
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		if t.Kind() != reflect.Struct {
			continue
		}

		name := t.Name()
		if name == "" {
			continue
		}

		// Generate full model schema (for responses).
		typeToSchema(t, gd.registry)

		// Generate Create variant (without auto-fields).
		createSchema := generateCreateVariant(t, gd.registry)
		gd.registry.Register("Create"+name, createSchema)

		// Generate Update variant (all fields optional).
		updateSchema := generateUpdateVariant(t, gd.registry)
		gd.registry.Register("Update"+name, updateSchema)
	}
}

// generateCreateVariant creates a schema variant for creating a resource.
// Excludes ID, CreatedAt, UpdatedAt, DeletedAt, and other auto-generated fields.
func generateCreateVariant(t reflect.Type, registry *TypeRegistry) *SchemaObject {
	schema := &SchemaObject{
		Type:       "object",
		Properties: make(map[string]*SchemaObject),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// Handle embedded structs.
		if field.Anonymous {
			embeddedType := field.Type
			for embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}
			if embeddedType.Kind() == reflect.Struct && specialTypeSchema(embeddedType) == nil {
				// Recurse into embedded struct but skip auto-fields.
				processCreateFields(embeddedType, schema, registry)
				continue
			}
		}

		if shouldSkipForCreate(field) {
			continue
		}

		tagInfo := mergeTags(
			field.Tag.Get("json"),
			field.Tag.Get("binding"),
			field.Tag.Get("gorm"),
			field.Tag.Get("docs"),
		)

		if tagInfo.JSONSkip || tagInfo.GORMSkip || tagInfo.Hidden {
			continue
		}

		propName := tagInfo.JSONName
		if propName == "" {
			propName = field.Name
		}

		fieldSchema := fieldToSchema(field.Type, tagInfo, registry)
		schema.Properties[propName] = fieldSchema

		if tagInfo.Required {
			schema.Required = append(schema.Required, propName)
		}
	}

	return schema
}

// processCreateFields processes struct fields for the create variant.
func processCreateFields(t reflect.Type, schema *SchemaObject, registry *TypeRegistry) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		if field.Anonymous {
			embeddedType := field.Type
			for embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}
			if embeddedType.Kind() == reflect.Struct && specialTypeSchema(embeddedType) == nil {
				processCreateFields(embeddedType, schema, registry)
				continue
			}
		}

		if shouldSkipForCreate(field) {
			continue
		}

		tagInfo := mergeTags(
			field.Tag.Get("json"),
			field.Tag.Get("binding"),
			field.Tag.Get("gorm"),
			field.Tag.Get("docs"),
		)

		if tagInfo.JSONSkip || tagInfo.GORMSkip || tagInfo.Hidden {
			continue
		}

		propName := tagInfo.JSONName
		if propName == "" {
			propName = field.Name
		}

		fieldSchema := fieldToSchema(field.Type, tagInfo, registry)
		schema.Properties[propName] = fieldSchema

		if tagInfo.Required {
			schema.Required = append(schema.Required, propName)
		}
	}
}

// shouldSkipForCreate determines if a field should be excluded from create variants.
func shouldSkipForCreate(field reflect.StructField) bool {
	name := field.Name
	gormTag := strings.ToLower(field.Tag.Get("gorm"))

	// Skip primary keys.
	if strings.Contains(gormTag, "primarykey") || strings.Contains(gormTag, "primary_key") {
		return true
	}

	// Skip auto-timestamps.
	if strings.Contains(gormTag, "autocreatetime") || strings.Contains(gormTag, "autoupdatetime") {
		return true
	}

	// Skip common auto-generated fields by name.
	switch name {
	case "ID", "CreatedAt", "UpdatedAt", "DeletedAt":
		return true
	}

	// Skip gorm.DeletedAt type.
	if field.Type.String() == "gorm.DeletedAt" {
		return true
	}

	// Skip time.Time fields named like timestamps.
	if field.Type == reflect.TypeOf(time.Time{}) || field.Type == reflect.TypeOf(&time.Time{}) {
		lower := strings.ToLower(name)
		if strings.Contains(lower, "created") || strings.Contains(lower, "updated") || strings.Contains(lower, "deleted") {
			return true
		}
	}

	return false
}

// generateUpdateVariant creates a schema variant for updating a resource.
// All fields are optional (no required array).
func generateUpdateVariant(t reflect.Type, registry *TypeRegistry) *SchemaObject {
	schema := &SchemaObject{
		Type:       "object",
		Properties: make(map[string]*SchemaObject),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		// Handle embedded structs.
		if field.Anonymous {
			embeddedType := field.Type
			for embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}
			if embeddedType.Kind() == reflect.Struct && specialTypeSchema(embeddedType) == nil {
				processUpdateFields(embeddedType, schema, registry)
				continue
			}
		}

		if shouldSkipForCreate(field) {
			continue
		}

		tagInfo := mergeTags(
			field.Tag.Get("json"),
			field.Tag.Get("binding"),
			field.Tag.Get("gorm"),
			field.Tag.Get("docs"),
		)

		if tagInfo.JSONSkip || tagInfo.GORMSkip || tagInfo.Hidden {
			continue
		}

		propName := tagInfo.JSONName
		if propName == "" {
			propName = field.Name
		}

		fieldSchema := fieldToSchema(field.Type, tagInfo, registry)
		// Clear readOnly for update variants.
		if fieldSchema.Ref == "" {
			fieldSchema.ReadOnly = false
		}
		schema.Properties[propName] = fieldSchema
		// No required fields in update variant.
	}

	return schema
}

// processUpdateFields processes struct fields for the update variant.
func processUpdateFields(t reflect.Type, schema *SchemaObject, registry *TypeRegistry) {
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		if field.Anonymous {
			embeddedType := field.Type
			for embeddedType.Kind() == reflect.Ptr {
				embeddedType = embeddedType.Elem()
			}
			if embeddedType.Kind() == reflect.Struct && specialTypeSchema(embeddedType) == nil {
				processUpdateFields(embeddedType, schema, registry)
				continue
			}
		}

		if shouldSkipForCreate(field) {
			continue
		}

		tagInfo := mergeTags(
			field.Tag.Get("json"),
			field.Tag.Get("binding"),
			field.Tag.Get("gorm"),
			field.Tag.Get("docs"),
		)

		if tagInfo.JSONSkip || tagInfo.GORMSkip || tagInfo.Hidden {
			continue
		}

		propName := tagInfo.JSONName
		if propName == "" {
			propName = field.Name
		}

		fieldSchema := fieldToSchema(field.Type, tagInfo, registry)
		if fieldSchema.Ref == "" {
			fieldSchema.ReadOnly = false
		}
		schema.Properties[propName] = fieldSchema
	}
}

// detectRelationships analyzes a struct for GORM relationships.
// Returns relationship metadata that can be used to enhance API documentation.
func detectRelationships(t reflect.Type) []RelationshipInfo {
	var relationships []RelationshipInfo

	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return relationships
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		ft := field.Type
		for ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}

		gormTag := field.Tag.Get("gorm")

		switch {
		case ft.Kind() == reflect.Slice && ft.Elem().Kind() == reflect.Struct:
			// HasMany or Many2Many.
			relType := RelHasMany
			if strings.Contains(strings.ToLower(gormTag), "many2many") {
				relType = RelMany2Many
			}
			relationships = append(relationships, RelationshipInfo{
				FieldName:    field.Name,
				Type:         relType,
				RelatedModel: ft.Elem().Name(),
			})

		case ft.Kind() == reflect.Struct && !isSpecialType(ft):
			// HasOne or BelongsTo.
			// If there's a corresponding ForeignKey field, it's BelongsTo.
			fkName := field.Name + "ID"
			relType := RelHasOne
			for j := 0; j < t.NumField(); j++ {
				if t.Field(j).Name == fkName {
					relType = RelBelongsTo
					break
				}
			}
			if strings.Contains(strings.ToLower(gormTag), "foreignkey") {
				relType = RelBelongsTo
			}
			relationships = append(relationships, RelationshipInfo{
				FieldName:    field.Name,
				Type:         relType,
				RelatedModel: ft.Name(),
			})
		}
	}

	return relationships
}

// isSpecialType checks if a type is a known special type (like time.Time).
func isSpecialType(t reflect.Type) bool {
	return specialTypeSchema(t) != nil
}

// RelType represents the type of a GORM relationship.
type RelType int

const (
	// RelHasOne represents a has-one relationship.
	RelHasOne RelType = iota
	// RelHasMany represents a has-many relationship.
	RelHasMany
	// RelBelongsTo represents a belongs-to relationship.
	RelBelongsTo
	// RelMany2Many represents a many-to-many relationship.
	RelMany2Many
)

// RelationshipInfo describes a detected GORM relationship.
type RelationshipInfo struct {
	// FieldName is the struct field name.
	FieldName string
	// Type is the relationship type.
	Type RelType
	// RelatedModel is the name of the related model.
	RelatedModel string
}
