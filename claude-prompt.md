# Claude Code Prompt — Build Gin Docs

Copy and paste this prompt into Claude Code to build the Gin Docs tool phase by phase.

---

## The Prompt

```
You are building a Go open-source tool called **Gin Docs** — a zero-annotation API documentation generator for Go applications using Gin and GORM. The module path is `github.com/MUKE-coder/gin-docs` and the main package is `gindocs/`.

## WHAT THIS TOOL DOES

Gin Docs automatically introspects a Gin router's route tree, uses Go reflection to read request/response structs and their tags, generates a valid OpenAPI 3.1 specification, and serves an interactive documentation UI — all mountable with a single function call:

```go
gindocs.Mount(router, db, gindocs.Config{})
// Docs at http://localhost:8080/docs
```

It supports **two UI options**: Swagger UI (default) and Scalar (modern alternative). Developers can switch between them via config or query parameter.

## DESIGN PRINCIPLES

1. **Zero annotations** — no comment-based documentation. Everything is inferred from code.
2. **One-line mount** — follows the same `Mount()` pattern as my other tools (github.com/MUKE-coder/gorm-studio and github.com/MUKE-coder/sentinel).
3. **Convention over configuration** — sensible defaults for everything, but fully configurable.
4. **Reflection-first** — use Go's `reflect` package to extract struct schemas, tags, and types.
5. **Standards-compliant** — generate valid OpenAPI 3.1 specs that work with any tooling.
6. **Production-ready** — proper error handling, caching, concurrent safety, no panics.

## ARCHITECTURE OVERVIEW

```
gin-docs/
├── gindocs/                    # Main package (this is what users import)
│   ├── mount.go                # Mount() entry point — registers all routes on Gin
│   ├── config.go               # Config struct, UIType constants, defaults
│   ├── engine.go               # GinDocs struct — orchestrates everything
│   ├── openapi_types.go        # Go structs representing OpenAPI 3.1 spec objects
│   ├── introspect.go           # Route tree introspection — reads Gin's Routes()
│   ├── reflect.go              # Handler analysis, request/response type detection
│   ├── schemas.go              # Go type → OpenAPI SchemaObject converter
│   ├── registry.go             # TypeRegistry for dedup and $ref management
│   ├── tags.go                 # Struct tag parser (json, binding, form, query, gorm, docs)
│   ├── openapi.go              # Full OpenAPI 3.1 spec assembler
│   ├── inference.go            # Smart description, summary, status code generation
│   ├── gorm.go                 # GORM model schema extraction and relationship detection
│   ├── overrides.go            # Route-level and group-level override API
│   ├── handlers.go             # HTTP handlers (serve UI, spec JSON/YAML, exports)
│   ├── ui_swagger.go           # Swagger UI HTML template and embedding
│   ├── ui_scalar.go            # Scalar UI HTML template and embedding
│   ├── export.go               # Postman and Insomnia collection generators
│   └── gorm.go                 # GORM-specific tag parsing and relationship detection
├── examples/
│   ├── basic/main.go           # Minimal example
│   └── full/main.go            # All features configured
├── main.go                     # Demo application with realistic blog API
├── go.mod
├── go.sum
├── README.md
├── CONTRIBUTING.md
├── SECURITY.md
├── LICENSE
└── Makefile
```

## CONFIGURATION

```go
type Config struct {
    // URL prefix for docs (default: "/docs")
    Prefix string
    
    // API metadata
    Title       string  // Default: auto-detect from module name
    Description string
    Version     string  // Default: "1.0.0"
    
    // UI selection: UISwagger (default) or UIScalar
    UI UIType
    
    // Dev mode: re-introspect on every request (default: auto from GIN_MODE)
    DevMode bool
    
    // Disable "Try It" functionality
    ReadOnly bool
    
    // Authentication for "Try It"
    Auth AuthConfig
    
    // API server URLs for "Try It"
    Servers []ServerInfo
    
    // Contact and license info
    Contact ContactInfo
    License LicenseInfo
    
    // Custom logo URL
    Logo string
    
    // Routes to exclude from docs
    ExcludeRoutes   []string  // Glob patterns
    ExcludePrefixes []string  // Path prefixes
    
    // GORM models to register as schemas
    Models []interface{}
    
    // Custom documentation sections (rendered as markdown)
    CustomSections []Section
    
    // Custom CSS to inject into the UI
    CustomCSS string
}

type UIType int
const (
    UISwagger UIType = iota  // Default
    UIScalar
)
```

## MOUNT FUNCTION SIGNATURE

```go
// Mount registers Gin Docs routes on the given router.
// db is optional (pass nil if not using GORM models).
// configs is variadic — pass zero or one Config.
func Mount(router *gin.Engine, db *gorm.DB, configs ...Config) *GinDocs
```

The returned `*GinDocs` supports a fluent override API:

```go
docs := gindocs.Mount(router, db, config)

// Route-level overrides
docs.Route("POST /api/users").
    Summary("Register a new user").
    Description("Creates a user account and sends verification email").
    RequestBody(CreateUserInput{}).
    Response(201, User{}, "User created").
    Response(409, nil, "Email already in use").
    Tags("Authentication")

// Group-level overrides
docs.Group("/api/admin/*").
    Tags("Admin").
    Security("bearerAuth")
```

## TYPE REGISTRATION SYSTEM

Since Go doesn't allow runtime inspection of what type is passed to `c.ShouldBindJSON()`, provide a registration mechanism:

```go
// Option 1: Via route overrides (recommended)
docs.Route("POST /api/users").
    RequestBody(CreateUserInput{}).
    Response(200, User{})

// Option 2: Via inline helper middleware
router.POST("/api/users", gindocs.Doc(gindocs.DocConfig{
    Summary:     "Create user",
    RequestBody: CreateUserInput{},
    Response:    User{},
}), handlers.CreateUser)

// Option 3: Convention-based (auto-detect structs named CreateUserRequest, etc.)
// This is best-effort and used as fallback
```

## STRUCT TAG SUPPORT

The schema engine must parse ALL of these tags:

```go
type User struct {
    ID        uint      `json:"id" gorm:"primarykey" docs:"description:Unique identifier"`
    Name      string    `json:"name" binding:"required,min=2,max=100" docs:"example:John Doe"`
    Email     string    `json:"email" binding:"required,email" gorm:"size:200;uniqueIndex"`
    Role      string    `json:"role" binding:"oneof=admin user moderator" gorm:"default:'user'"`
    Age       int       `json:"age,omitempty" binding:"gte=0,lte=150"`
    Avatar    string    `json:"avatar,omitempty" docs:"format:uri"`
    Bio       string    `json:"bio,omitempty" gorm:"type:text" docs:"description:User biography"`
    IsActive  bool      `json:"is_active" gorm:"default:true"`
    Score     float64   `json:"score" binding:"min=0,max=100"`
    Tags      []string  `json:"tags" gorm:"serializer:json"`
    CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime" docs:"description:Account creation timestamp"`
    UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
```

Tag parsing rules:
- `json:"name,omitempty"` → property name is "name", field is optional
- `json:"-"` → skip this field entirely
- `binding:"required"` → add to `required` array
- `binding:"min=N,max=M"` → `minimum`/`maximum` for numbers, `minLength`/`maxLength` for strings
- `binding:"email"` → `format: "email"`
- `binding:"oneof=a b c"` → `enum: ["a", "b", "c"]`
- `binding:"gte=N"` → `minimum: N`
- `gorm:"primarykey"` → mark as ID, readOnly
- `gorm:"size:N"` → `maxLength: N`
- `gorm:"uniqueIndex"` → note "Must be unique" in description
- `gorm:"default:'value'"` → `default: "value"`
- `gorm:"autoCreateTime"` / `autoUpdateTime"` → `readOnly: true`
- `gorm:"-"` → skip field
- `docs:"description:...,example:...,deprecated,hidden,format:...,enum:a|b|c"`

## SMART INFERENCE RULES

Auto-generate summaries from HTTP method + path:
- `GET /users` → "List all users"
- `GET /users/:id` → "Get a user by ID"
- `POST /users` → "Create a new user"
- `PUT /users/:id` → "Update a user by ID"
- `PATCH /users/:id` → "Partially update a user by ID"
- `DELETE /users/:id` → "Delete a user by ID"
- `GET /users/:id/posts` → "List posts for a user"

Auto-infer response status codes:
- GET → 200, POST → 201, PUT/PATCH → 200, DELETE → 204
- Add 400 for routes with request bodies
- Add 401/403 for authenticated routes
- Add 404 for routes with path params
- Add 500 always

Auto-generate examples from field names/types:
- `Name string` → "John Doe"
- `Email string` → "user@example.com"
- `ID uint` → 1
- `CreatedAt time.Time` → "2025-01-15T10:30:00Z"

## GORM MODEL FEATURES

When GORM models are registered:
1. Parse GORM tags for constraints and metadata
2. Detect relationships (HasOne, HasMany, BelongsTo, Many2Many)
3. Generate schema variants:
   - `User` — full model (for responses)
   - `CreateUser` — without ID, CreatedAt, UpdatedAt, DeletedAt
   - `UpdateUser` — all fields optional
4. Handle circular references via `$ref`

## UI REQUIREMENTS

### Swagger UI (Default)
- Embed Swagger UI dist via `go:embed`
- Serve at `/docs` (or configured prefix)
- Enable "Try It Out" unless `ReadOnly: true`
- Support auth configuration (Bearer, API Key, Basic)
- Dark mode support
- Custom branding (logo, title)

### Scalar UI (Alternative)
- Load via CDN `<script>` tag (or embed for offline)
- Modern, clean design
- Same spec endpoint (`/docs/openapi.json`)
- Authentication support
- Theme configuration

### UI Switching
- Config: `Config.UI = gindocs.UIScalar`
- Query param: `/docs?ui=scalar` or `/docs?ui=swagger`
- Both use the same OpenAPI spec

## ENDPOINTS TO REGISTER

| Method | Path | Description |
|--------|------|-------------|
| GET | `/docs` | Serve documentation UI (Swagger or Scalar) |
| GET | `/docs/openapi.json` | OpenAPI 3.1 spec as JSON |
| GET | `/docs/openapi.yaml` | OpenAPI 3.1 spec as YAML |
| GET | `/docs/export/postman` | Postman v2.1 collection |
| GET | `/docs/export/insomnia` | Insomnia v4 export |
| GET | `/docs/assets/*filepath` | Static assets (Swagger UI files) |

## DEMO APPLICATION

Build a demo `main.go` at the project root with a realistic blog API:

Models: User, Post, Comment, Tag, Category (with GORM relationships)
Routes:
- POST /api/auth/register, POST /api/auth/login
- GET/POST /api/users, GET/PUT/DELETE /api/users/:id
- GET/POST /api/posts, GET/PUT/DELETE /api/posts/:id
- GET/POST /api/posts/:id/comments
- GET /api/tags, GET /api/categories
- GET /api/search?q=...
- POST /api/upload (file upload example)

Include auth middleware on write routes, pagination on list routes.
Mount Gin Docs with full config, custom sections, route overrides, and GORM models.

## IMPLEMENTATION ORDER

Build in this exact order (each step should compile and be testable):

### Step 1: Foundation
- Initialize go.mod, create all directories
- Define Config, OpenAPI types, UIType constants
- Create Mount() skeleton that serves a placeholder HTML page at /docs
- Create GinDocs engine struct
- Verify: `go run main.go` → visit localhost:8080/docs → see placeholder

### Step 2: Route Introspection
- Implement Introspect() — read router.Routes()
- Extract path params, convert to OpenAPI format
- Auto-detect route groups as tags
- Build RouteMetadata registry
- Verify: print discovered routes to console

### Step 3: Schema Engine
- Implement typeToSchema() for all Go types
- Implement struct tag parsing (json, binding, form, query, uri)
- Implement TypeRegistry with $ref support
- Handle embedded structs, pointers, slices, maps
- Write tests for schema generation

### Step 4: OpenAPI Spec Generation
- Assemble full OpenAPI 3.1 spec from routes + schemas
- Generate path items, operations, parameters
- Generate request bodies and responses
- Generate security schemes
- Serve at /docs/openapi.json
- Verify: fetch spec and validate with an OpenAPI validator

### Step 5: Smart Inference
- Implement summary generation from method + path
- Implement status code inference
- Implement example value generation
- Implement parameter description inference
- Apply inference to all operations in the spec

### Step 6: Swagger UI
- Embed Swagger UI assets
- Create HTML template with configuration
- Serve at /docs with "Try It" support
- Add auth support, dark mode, branding
- Verify: visit /docs → see full interactive documentation

### Step 7: Scalar UI
- Create Scalar HTML template
- Support config-based and query-param UI switching
- Ensure "Try It" works
- Verify: visit /docs?ui=scalar → see Scalar documentation

### Step 8: GORM Integration
- Parse gorm tags
- Detect relationships
- Generate model variants (Create, Update, Full)
- Register models as schema components
- Verify: GORM models appear as schemas in the spec

### Step 9: Override System
- Implement docs:"..." struct tag
- Implement Route().Summary().RequestBody() fluent API
- Implement Group() overrides
- Implement Doc() middleware helper
- Custom sections support
- Verify: overrides appear in generated spec

### Step 10: Exports & Polish
- Implement Postman collection export
- Implement Insomnia export  
- Implement YAML spec endpoint
- Dev mode live reload
- Verify: download Postman collection and import it

### Step 11: Demo & Examples
- Build full blog API demo in main.go
- Create examples/basic/ and examples/full/
- Test the complete flow end to end

### Step 12: Documentation & Release
- Write comprehensive README.md with badges, quick start, config reference
- Write CONTRIBUTING.md, SECURITY.md, LICENSE (MIT)
- Create Makefile with build, test, lint targets
- Create .github/workflows/ci.yml
- Ensure all tests pass, go vet, golangci-lint clean

## QUALITY STANDARDS

- Every public function has a doc comment
- No panics — return errors gracefully
- Use sync.RWMutex for concurrent spec access
- Table-driven tests for all parsers
- go vet and golangci-lint clean
- No external dependencies beyond Gin, GORM, and YAML parser
- go:embed for all static assets (no runtime file reads)
- Proper HTTP cache headers on all responses
- JSON output is indented for readability (openapi.json)

## IMPORTANT NOTES

- Do NOT use swaggo/swag or any annotation-based approach. The whole point is ZERO annotations.
- The module path is `github.com/MUKE-coder/gin-docs` and users import `github.com/MUKE-coder/gin-docs/gindocs`
- Follow the same patterns as my other tools: simple Mount() function, embedded UI, zero-config defaults
- The OpenAPI spec MUST be valid 3.1 — test this
- Both Swagger UI and Scalar must work with the same spec
- Handle edge cases: empty router, no models, circular refs, anonymous structs, interface{} fields

Start with Step 1 and proceed sequentially. After each step, ensure the project compiles, the demo runs, and any new functionality is tested. Ask me if you need clarification on any requirement.
```
