package gindocs

// OpenAPISpec represents a complete OpenAPI 3.1 specification.
type OpenAPISpec struct {
	OpenAPI      string                `json:"openapi"`
	Info         InfoObject            `json:"info"`
	Servers      []ServerObject        `json:"servers,omitempty"`
	Paths        map[string]*PathItem  `json:"paths"`
	Components   *ComponentsObject     `json:"components,omitempty"`
	Security     []SecurityRequirement `json:"security,omitempty"`
	Tags         []TagObject           `json:"tags,omitempty"`
	ExternalDocs *ExternalDocsObject   `json:"externalDocs,omitempty"`
}

// InfoObject provides metadata about the API.
type InfoObject struct {
	Title          string         `json:"title"`
	Description    string         `json:"description,omitempty"`
	TermsOfService string         `json:"termsOfService,omitempty"`
	Contact        *ContactObject `json:"contact,omitempty"`
	License        *LicenseObject `json:"license,omitempty"`
	Version        string         `json:"version"`
}

// ContactObject holds contact information.
type ContactObject struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

// LicenseObject holds license information.
type LicenseObject struct {
	Name string `json:"name"`
	URL  string `json:"url,omitempty"`
}

// ServerObject describes a server.
type ServerObject struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`
}

// PathItem describes operations available on a single path.
type PathItem struct {
	Get     *OperationObject `json:"get,omitempty"`
	Post    *OperationObject `json:"post,omitempty"`
	Put     *OperationObject `json:"put,omitempty"`
	Patch   *OperationObject `json:"patch,omitempty"`
	Delete  *OperationObject `json:"delete,omitempty"`
	Head    *OperationObject `json:"head,omitempty"`
	Options *OperationObject `json:"options,omitempty"`
}

// SetOperation sets the operation for the given HTTP method on the path item.
func (p *PathItem) SetOperation(method string, op *OperationObject) {
	switch method {
	case "GET":
		p.Get = op
	case "POST":
		p.Post = op
	case "PUT":
		p.Put = op
	case "PATCH":
		p.Patch = op
	case "DELETE":
		p.Delete = op
	case "HEAD":
		p.Head = op
	case "OPTIONS":
		p.Options = op
	}
}

// OperationObject describes a single API operation on a path.
type OperationObject struct {
	Tags         []string              `json:"tags,omitempty"`
	Summary      string                `json:"summary,omitempty"`
	Description  string                `json:"description,omitempty"`
	OperationID  string                `json:"operationId,omitempty"`
	Parameters   []ParameterObject     `json:"parameters,omitempty"`
	RequestBody  *RequestBodyObject    `json:"requestBody,omitempty"`
	Responses    map[string]*Response  `json:"responses"`
	Security     []SecurityRequirement `json:"security,omitempty"`
	Deprecated   bool                  `json:"deprecated,omitempty"`
	ExternalDocs *ExternalDocsObject   `json:"externalDocs,omitempty"`
}

// ParameterObject describes a single operation parameter.
type ParameterObject struct {
	Name        string        `json:"name"`
	In          string        `json:"in"` // "query", "header", "path", "cookie"
	Description string        `json:"description,omitempty"`
	Required    bool          `json:"required,omitempty"`
	Deprecated  bool          `json:"deprecated,omitempty"`
	Schema      *SchemaObject `json:"schema,omitempty"`
	Example     interface{}   `json:"example,omitempty"`
}

// RequestBodyObject describes a request body.
type RequestBodyObject struct {
	Description string                `json:"description,omitempty"`
	Content     map[string]MediaType  `json:"content"`
	Required    bool                  `json:"required,omitempty"`
}

// MediaType describes a media type with a schema and examples.
type MediaType struct {
	Schema  *SchemaObject `json:"schema,omitempty"`
	Example interface{}   `json:"example,omitempty"`
}

// Response describes a single response from an API operation.
type Response struct {
	Description string               `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
	Headers     map[string]*Header   `json:"headers,omitempty"`
}

// Header describes a response header.
type Header struct {
	Description string        `json:"description,omitempty"`
	Schema      *SchemaObject `json:"schema,omitempty"`
}

// SchemaObject represents a JSON Schema object (OpenAPI 3.1 compatible).
type SchemaObject struct {
	// Reference
	Ref string `json:"$ref,omitempty"`

	// Type
	Type   string `json:"type,omitempty"`
	Format string `json:"format,omitempty"`

	// Metadata
	Title       string      `json:"title,omitempty"`
	Description string      `json:"description,omitempty"`
	Default     interface{} `json:"default,omitempty"`
	Example     interface{} `json:"example,omitempty"`
	Deprecated  bool        `json:"deprecated,omitempty"`
	ReadOnly    bool        `json:"readOnly,omitempty"`
	WriteOnly   bool        `json:"writeOnly,omitempty"`
	Nullable    bool        `json:"nullable,omitempty"`

	// Validation — numbers
	Minimum          *float64 `json:"minimum,omitempty"`
	Maximum          *float64 `json:"maximum,omitempty"`
	ExclusiveMinimum *float64 `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum *float64 `json:"exclusiveMaximum,omitempty"`
	MultipleOf       *float64 `json:"multipleOf,omitempty"`

	// Validation — strings
	MinLength *int   `json:"minLength,omitempty"`
	MaxLength *int   `json:"maxLength,omitempty"`
	Pattern   string `json:"pattern,omitempty"`

	// Validation — arrays
	Items    *SchemaObject `json:"items,omitempty"`
	MinItems *int          `json:"minItems,omitempty"`
	MaxItems *int          `json:"maxItems,omitempty"`

	// Validation — objects
	Properties           map[string]*SchemaObject `json:"properties,omitempty"`
	Required             []string                 `json:"required,omitempty"`
	AdditionalProperties *SchemaObject            `json:"additionalProperties,omitempty"`

	// Enum
	Enum []interface{} `json:"enum,omitempty"`

	// Composition
	AllOf []*SchemaObject `json:"allOf,omitempty"`
	OneOf []*SchemaObject `json:"oneOf,omitempty"`
	AnyOf []*SchemaObject `json:"anyOf,omitempty"`
}

// ComponentsObject holds reusable components.
type ComponentsObject struct {
	Schemas         map[string]*SchemaObject         `json:"schemas,omitempty"`
	SecuritySchemes map[string]*SecuritySchemeObject  `json:"securitySchemes,omitempty"`
	Parameters      map[string]*ParameterObject      `json:"parameters,omitempty"`
	RequestBodies   map[string]*RequestBodyObject     `json:"requestBodies,omitempty"`
	Responses       map[string]*Response              `json:"responses,omitempty"`
}

// SecuritySchemeObject defines a security scheme.
type SecuritySchemeObject struct {
	Type         string `json:"type"`
	Description  string `json:"description,omitempty"`
	Name         string `json:"name,omitempty"`   // for apiKey
	In           string `json:"in,omitempty"`     // for apiKey: "header", "query", "cookie"
	Scheme       string `json:"scheme,omitempty"` // for http: "bearer", "basic"
	BearerFormat string `json:"bearerFormat,omitempty"`
}

// SecurityRequirement maps security scheme names to required scopes.
type SecurityRequirement map[string][]string

// TagObject describes a tag used for API operation grouping.
type TagObject struct {
	Name         string              `json:"name"`
	Description  string              `json:"description,omitempty"`
	ExternalDocs *ExternalDocsObject `json:"externalDocs,omitempty"`
}

// ExternalDocsObject describes external documentation.
type ExternalDocsObject struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}
