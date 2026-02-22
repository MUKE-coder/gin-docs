package gindocs

import (
	"reflect"
	"testing"
	"time"
)

func TestTypeToSchema_Primitives(t *testing.T) {
	registry := newTypeRegistry()

	tests := []struct {
		name       string
		goType     reflect.Type
		wantType   string
		wantFormat string
	}{
		{"bool", reflect.TypeOf(false), "boolean", ""},
		{"int", reflect.TypeOf(0), "integer", "int32"},
		{"int32", reflect.TypeOf(int32(0)), "integer", "int32"},
		{"int64", reflect.TypeOf(int64(0)), "integer", "int64"},
		{"uint", reflect.TypeOf(uint(0)), "integer", "int32"},
		{"uint64", reflect.TypeOf(uint64(0)), "integer", "int64"},
		{"float32", reflect.TypeOf(float32(0)), "number", "float"},
		{"float64", reflect.TypeOf(float64(0)), "number", "double"},
		{"string", reflect.TypeOf(""), "string", ""},
		{"time.Time", reflect.TypeOf(time.Time{}), "string", "date-time"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schema := typeToSchema(tt.goType, registry)
			if schema.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", schema.Type, tt.wantType)
			}
			if schema.Format != tt.wantFormat {
				t.Errorf("Format = %q, want %q", schema.Format, tt.wantFormat)
			}
		})
	}
}

func TestTypeToSchema_Slice(t *testing.T) {
	registry := newTypeRegistry()

	schema := typeToSchema(reflect.TypeOf([]string{}), registry)
	if schema.Type != "array" {
		t.Errorf("Type = %q, want %q", schema.Type, "array")
	}
	if schema.Items == nil || schema.Items.Type != "string" {
		t.Error("Items should be string schema")
	}
}

func TestTypeToSchema_ByteSlice(t *testing.T) {
	registry := newTypeRegistry()

	schema := typeToSchema(reflect.TypeOf([]byte{}), registry)
	if schema.Type != "string" {
		t.Errorf("Type = %q, want %q", schema.Type, "string")
	}
	if schema.Format != "byte" {
		t.Errorf("Format = %q, want %q", schema.Format, "byte")
	}
}

func TestTypeToSchema_Map(t *testing.T) {
	registry := newTypeRegistry()

	schema := typeToSchema(reflect.TypeOf(map[string]int{}), registry)
	if schema.Type != "object" {
		t.Errorf("Type = %q, want %q", schema.Type, "object")
	}
	if schema.AdditionalProperties == nil || schema.AdditionalProperties.Type != "integer" {
		t.Error("AdditionalProperties should be integer schema")
	}
}

func TestTypeToSchema_Pointer(t *testing.T) {
	registry := newTypeRegistry()

	var s *string
	schema := typeToSchema(reflect.TypeOf(s), registry)
	if schema.Type != "string" {
		t.Errorf("Type = %q, want %q", schema.Type, "string")
	}
}

type TestUser struct {
	ID    uint   `json:"id" gorm:"primarykey"`
	Name  string `json:"name" binding:"required,min=2,max=100"`
	Email string `json:"email" binding:"required,email" gorm:"size:200;uniqueIndex"`
	Role  string `json:"role" binding:"oneof=admin user moderator" gorm:"default:'user'"`
	Age   int    `json:"age,omitempty" binding:"gte=0,lte=150"`
	Bio   string `json:"bio,omitempty" gorm:"type:text" docs:"description:User biography"`
}

func TestTypeToSchema_Struct(t *testing.T) {
	registry := newTypeRegistry()

	schema := typeToSchema(reflect.TypeOf(TestUser{}), registry)

	// Should return a $ref.
	if schema.Ref == "" {
		t.Error("Struct should return a $ref")
	}
	if schema.Ref != "#/components/schemas/TestUser" {
		t.Errorf("Ref = %q, want %q", schema.Ref, "#/components/schemas/TestUser")
	}

	// Check the registered schema.
	registered, ok := registry.Get("TestUser")
	if !ok {
		t.Fatal("TestUser should be registered")
	}
	if registered.Type != "object" {
		t.Errorf("Type = %q, want %q", registered.Type, "object")
	}

	// Check properties.
	if _, ok := registered.Properties["id"]; !ok {
		t.Error("Should have 'id' property")
	}
	if _, ok := registered.Properties["name"]; !ok {
		t.Error("Should have 'name' property")
	}
	if _, ok := registered.Properties["email"]; !ok {
		t.Error("Should have 'email' property")
	}
	if _, ok := registered.Properties["role"]; !ok {
		t.Error("Should have 'role' property")
	}

	// Check required fields.
	requiredSet := make(map[string]bool)
	for _, r := range registered.Required {
		requiredSet[r] = true
	}
	if !requiredSet["name"] {
		t.Error("'name' should be required")
	}
	if !requiredSet["email"] {
		t.Error("'email' should be required")
	}

	// Check primary key is readOnly.
	idSchema := registered.Properties["id"]
	if !idSchema.ReadOnly {
		t.Error("'id' should be readOnly (primarykey)")
	}

	// Check email has format and uniqueIndex description.
	emailSchema := registered.Properties["email"]
	if emailSchema.Format != "email" {
		t.Errorf("Email format = %q, want %q", emailSchema.Format, "email")
	}
	if emailSchema.MaxLength == nil || *emailSchema.MaxLength != 200 {
		t.Error("Email should have maxLength 200 from gorm:size")
	}

	// Check role has enum.
	roleSchema := registered.Properties["role"]
	if len(roleSchema.Enum) != 3 {
		t.Errorf("Role enum length = %d, want 3", len(roleSchema.Enum))
	}
	if roleSchema.Default != "user" {
		t.Errorf("Role default = %v, want %q", roleSchema.Default, "user")
	}

	// Check bio has description from docs tag.
	bioSchema := registered.Properties["bio"]
	if bioSchema.Description != "User biography" {
		t.Errorf("Bio description = %q, want %q", bioSchema.Description, "User biography")
	}

	// Check age has gte/lte constraints.
	ageSchema := registered.Properties["age"]
	if ageSchema.Minimum == nil || *ageSchema.Minimum != 0 {
		t.Errorf("Age minimum = %v, want 0", ageSchema.Minimum)
	}
	if ageSchema.Maximum == nil || *ageSchema.Maximum != 150 {
		t.Errorf("Age maximum = %v, want 150", ageSchema.Maximum)
	}
}

type TestEmbeddedBase struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

type TestEmbeddedChild struct {
	TestEmbeddedBase
	Name string `json:"name" binding:"required"`
}

func TestTypeToSchema_EmbeddedStruct(t *testing.T) {
	registry := newTypeRegistry()

	typeToSchema(reflect.TypeOf(TestEmbeddedChild{}), registry)

	schema, ok := registry.Get("TestEmbeddedChild")
	if !ok {
		t.Fatal("TestEmbeddedChild should be registered")
	}

	// Should have properties from both embedded and own fields.
	if _, ok := schema.Properties["id"]; !ok {
		t.Error("Should have 'id' from embedded struct")
	}
	if _, ok := schema.Properties["created_at"]; !ok {
		t.Error("Should have 'created_at' from embedded struct")
	}
	if _, ok := schema.Properties["name"]; !ok {
		t.Error("Should have 'name' from own fields")
	}
}

type TestJSONSkip struct {
	ID      uint   `json:"id"`
	Secret  string `json:"-"`
	Visible string `json:"visible"`
}

func TestTypeToSchema_JSONSkip(t *testing.T) {
	registry := newTypeRegistry()

	typeToSchema(reflect.TypeOf(TestJSONSkip{}), registry)

	schema, ok := registry.Get("TestJSONSkip")
	if !ok {
		t.Fatal("TestJSONSkip should be registered")
	}

	if _, ok := schema.Properties["Secret"]; ok {
		t.Error("Secret with json:\"-\" should be skipped")
	}
	if _, ok := schema.Properties["id"]; !ok {
		t.Error("Should have 'id'")
	}
	if _, ok := schema.Properties["visible"]; !ok {
		t.Error("Should have 'visible'")
	}
}

func TestParseJSONTag(t *testing.T) {
	tests := []struct {
		tag       string
		wantName  string
		wantOmit  bool
		wantSkip  bool
	}{
		{"name", "name", false, false},
		{"name,omitempty", "name", true, false},
		{"-", "", false, true},
		{"", "", false, false},
		{"field_name,omitempty", "field_name", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			name, omit, skip := parseJSONTag(tt.tag)
			if name != tt.wantName {
				t.Errorf("name = %q, want %q", name, tt.wantName)
			}
			if omit != tt.wantOmit {
				t.Errorf("omitEmpty = %v, want %v", omit, tt.wantOmit)
			}
			if skip != tt.wantSkip {
				t.Errorf("skip = %v, want %v", skip, tt.wantSkip)
			}
		})
	}
}

func TestParseBindingTag(t *testing.T) {
	tests := []struct {
		tag      string
		check    func(TagInfo) bool
		desc     string
	}{
		{"required", func(i TagInfo) bool { return i.Required }, "should be required"},
		{"email", func(i TagInfo) bool { return i.Format == "email" }, "should have email format"},
		{"uri", func(i TagInfo) bool { return i.Format == "uri" }, "should have uri format"},
		{"uuid", func(i TagInfo) bool { return i.Format == "uuid" }, "should have uuid format"},
		{"oneof=a b c", func(i TagInfo) bool { return len(i.Enum) == 3 }, "should have 3 enum values"},
		{"min=5", func(i TagInfo) bool { return i.MinLength != nil && *i.MinLength == 5 }, "should have minLength 5"},
		{"max=100", func(i TagInfo) bool { return i.MaxLength != nil && *i.MaxLength == 100 }, "should have maxLength 100"},
		{"gte=0", func(i TagInfo) bool { return i.Minimum != nil && *i.Minimum == 0 }, "should have minimum 0"},
		{"lte=150", func(i TagInfo) bool { return i.Maximum != nil && *i.Maximum == 150 }, "should have maximum 150"},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			info := parseBindingTag(tt.tag)
			if !tt.check(info) {
				t.Error(tt.desc)
			}
		})
	}
}

func TestParseGORMTag(t *testing.T) {
	tests := []struct {
		tag   string
		check func(TagInfo) bool
		desc  string
	}{
		{"primarykey", func(i TagInfo) bool { return i.PrimaryKey }, "should be primarykey"},
		{"autoCreateTime", func(i TagInfo) bool { return i.AutoCreateTime }, "should be autoCreateTime"},
		{"autoUpdateTime", func(i TagInfo) bool { return i.AutoUpdateTime }, "should be autoUpdateTime"},
		{"size:200", func(i TagInfo) bool { return i.GORMSize != nil && *i.GORMSize == 200 }, "should have size 200"},
		{"uniqueIndex", func(i TagInfo) bool { return i.UniqueIndex }, "should be uniqueIndex"},
		{"default:'user'", func(i TagInfo) bool { return i.GORMDefault != nil && *i.GORMDefault == "user" }, "should have default 'user'"},
		{"-", func(i TagInfo) bool { return i.GORMSkip }, "should be skipped"},
		{"type:text", func(i TagInfo) bool { return i.GORMType == "text" }, "should have type text"},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			info := parseGORMTag(tt.tag)
			if !tt.check(info) {
				t.Error(tt.desc)
			}
		})
	}
}

func TestParseDocsTag(t *testing.T) {
	tests := []struct {
		tag   string
		check func(TagInfo) bool
		desc  string
	}{
		{"description:A test field", func(i TagInfo) bool { return i.Description == "A test field" }, "should have description"},
		{"example:John Doe", func(i TagInfo) bool { return i.Example == "John Doe" }, "should have example"},
		{"deprecated", func(i TagInfo) bool { return i.Deprecated }, "should be deprecated"},
		{"hidden", func(i TagInfo) bool { return i.Hidden }, "should be hidden"},
		{"format:uri", func(i TagInfo) bool { return i.DocsFormat == "uri" }, "should have format"},
		{"enum:a|b|c", func(i TagInfo) bool { return len(i.DocsEnum) == 3 }, "should have enum"},
	}

	for _, tt := range tests {
		t.Run(tt.tag, func(t *testing.T) {
			info := parseDocsTag(tt.tag)
			if !tt.check(info) {
				t.Error(tt.desc)
			}
		})
	}
}

// Test circular reference handling.
type TestNode struct {
	ID       uint      `json:"id"`
	Name     string    `json:"name"`
	Children []TestNode `json:"children"`
}

func TestTypeToSchema_CircularRef(t *testing.T) {
	registry := newTypeRegistry()

	schema := typeToSchema(reflect.TypeOf(TestNode{}), registry)
	if schema.Ref == "" {
		t.Error("Should return a $ref for struct types")
	}

	// Should not panic or loop infinitely.
	registered, ok := registry.Get("TestNode")
	if !ok {
		t.Fatal("TestNode should be registered")
	}

	// Children should be an array with $ref back to TestNode.
	children := registered.Properties["children"]
	if children.Type != "array" {
		t.Errorf("Children type = %q, want %q", children.Type, "array")
	}
	if children.Items == nil {
		t.Fatal("Children items should not be nil")
	}
	if children.Items.Ref != "#/components/schemas/TestNode" {
		t.Errorf("Children items ref = %q, want %q", children.Items.Ref, "#/components/schemas/TestNode")
	}
}
