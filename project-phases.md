# Gin Docs — Project Phases & Tasks

## Phase Overview

| Phase | Name | Description | Estimated Tasks |
|-------|------|-------------|-----------------|
| 1 | Foundation & Project Setup | Module init, config types, mount skeleton | 7 tasks |
| 2 | Route Introspection Engine | Parse Gin route tree, extract metadata | 6 tasks |
| 3 | Struct Reflection & Schema Engine | Reflect Go types into OpenAPI schemas | 9 tasks |
| 4 | OpenAPI 3.1 Spec Generator | Assemble the full OpenAPI document | 8 tasks |
| 5 | Swagger UI Integration | Embed and serve Swagger UI | 6 tasks |
| 6 | Scalar UI Integration | Embed and serve Scalar as alternative | 5 tasks |
| 7 | Smart Inference Engine | Auto-generate descriptions, status codes | 6 tasks |
| 8 | GORM Model Integration | Extract schemas from GORM models | 5 tasks |
| 9 | Override & Customization System | Manual overrides, docs tag, branding | 7 tasks |
| 10 | Export & Integration Features | Postman export, spec download, webhooks | 5 tasks |
| 11 | Demo Application & Examples | Working demo app with realistic models | 5 tasks |
| 12 | Testing, Docs & Release | Tests, README, CI/CD, publish | 7 tasks |

---

## Phase 1: Foundation & Project Setup

**Goal:** Initialize the Go module, define all configuration types, and create the `Mount()` function skeleton that registers placeholder routes on a Gin router.

### Tasks

#### 1.1 — Initialize Go Module
- Create `go.mod` with module path `github.com/MUKE-coder/gin-docs`
- Add initial dependencies: `github.com/gin-gonic/gin`, `gorm.io/gorm`
- Create the `gindocs/` package directory
- Create `main.go` at root as a demo entry point (placeholder)

#### 1.2 — Define Configuration Types (`gindocs/config.go`)
- Define `Config` struct with fields:
  - `Prefix string` (default: `/docs`)
  - `Title string` (default: auto-detected from module name)
  - `Description string`
  - `Version string` (default: `1.0.0`)
  - `UI UIType` — enum: `UISwagger` (default), `UIScalar`
  - `DevMode bool` — enable live reload (default: based on `GIN_MODE`)
  - `ReadOnly bool` — disable "Try It" in production
  - `Auth AuthConfig` — for "Try It" auth headers
  - `Contact ContactInfo` — name, email, URL
  - `License LicenseInfo` — name, URL
  - `Logo string` — URL to custom logo
  - `ExcludeRoutes []string` — glob patterns to hide routes
  - `ExcludePrefixes []string` — prefixes to exclude (e.g., `/debug`)
  - `Servers []ServerInfo` — base URLs for "Try It"
  - `CustomSections []Section` — markdown guides/changelogs
- Define `UIType` constants: `UISwagger`, `UIScalar`
- Define `AuthConfig` with `Type` (Bearer, APIKey, Basic), `HeaderName`, `Scheme`
- Define `ContactInfo`, `LicenseInfo`, `ServerInfo`, `Section` structs
- Implement `applyDefaults()` to fill in sensible defaults

#### 1.3 — Define OpenAPI Types (`gindocs/openapi_types.go`)
- Define Go structs that represent the OpenAPI 3.1 specification:
  - `OpenAPISpec` — root document
  - `InfoObject`, `ServerObject`, `PathItemObject`
  - `OperationObject`, `ParameterObject`, `RequestBodyObject`
  - `ResponseObject`, `MediaTypeObject`, `SchemaObject`
  - `ComponentsObject`, `SecuritySchemeObject`, `TagObject`
  - `ExampleObject`, `ExternalDocObject`
- Ensure all structs serialize to valid OpenAPI 3.1 JSON via `json` tags
- Use `omitempty` for optional fields

#### 1.4 — Create Mount Function Skeleton (`gindocs/mount.go`)
- Implement `Mount(router *gin.Engine, db *gorm.DB, configs ...Config)` 
- Accept variadic config (use first or default)
- Apply config defaults
- Register placeholder routes:
  - `GET /docs` — serve UI (placeholder HTML)
  - `GET /docs/openapi.json` — serve spec (placeholder)
  - `GET /docs/openapi.yaml` — serve spec as YAML
  - `GET /docs/assets/*filepath` — serve static assets
- Return a `*GinDocs` instance for programmatic access
- Store reference to the router and db for later introspection

#### 1.5 — Create Internal State Manager (`gindocs/engine.go`)
- Define `GinDocs` struct holding:
  - `config Config`
  - `router *gin.Engine`
  - `db *gorm.DB`
  - `spec *OpenAPISpec` — the generated spec (cached)
  - `typeRegistry map[reflect.Type]*SchemaObject` — deduplication cache
  - `overrides map[string]*OperationOverride` — per-route overrides
  - `mu sync.RWMutex` — for concurrent access
- Implement `Generate()` method that orchestrates the full pipeline
- Implement `Refresh()` for dev-mode live reload
- Implement `Spec() *OpenAPISpec` getter

#### 1.6 — Create Makefile & Project Structure
- Create `Makefile` with targets: `build`, `run`, `test`, `lint`, `tidy`
- Create directory structure:
  - `gindocs/` — main package
  - `examples/basic/` — basic example
  - `examples/full/` — full-featured example
  - `docs/` — documentation
- Create `.github/workflows/ci.yml` placeholder
- Create `.gitignore`

#### 1.7 — Verify Foundation
- Run `go mod tidy`
- Ensure `Mount()` compiles and serves placeholder page at `/docs`
- Verify config defaults work correctly
- Write one basic test: `TestMountRegistersRoutes`

---

## Phase 2: Route Introspection Engine

**Goal:** Parse the Gin route tree to extract all registered routes with their HTTP methods, path parameters, handler names, and middleware chains.

### Tasks

#### 2.1 — Introspect Gin Route Tree (`gindocs/introspect.go`)
- Call `router.Routes()` to get all registered `gin.RouteInfo` entries
- For each route, extract:
  - `Method` (GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD)
  - `Path` (Gin path format: `/users/:id`)
  - `Handler` (fully qualified handler function name)
  - `HandlerFunc` (the actual function, for reflection)
- Convert Gin path params (`:id`, `*filepath`) to OpenAPI format (`{id}`, `{filepath}`)
- Filter out excluded routes based on `Config.ExcludeRoutes` and `Config.ExcludePrefixes`
- Filter out the docs routes themselves (`/docs/*`)

#### 2.2 — Extract Path Parameters
- Parse `:param` and `*param` from route paths
- Create `ParameterObject` for each with:
  - `In: "path"`
  - `Required: true`
  - `Name` from the param name
  - `Schema` inferred from name (`:id` → integer, others → string)
- Handle wildcard params (`*filepath`) as catch-all strings

#### 2.3 — Detect Route Groups & Generate Tags
- Analyze route paths to detect logical groups:
  - `/api/users/*` → tag "Users"
  - `/api/posts/*` → tag "Posts"
- Use the first significant path segment after common prefixes (`/api/`, `/v1/`, etc.)
- Capitalize and singularize/pluralize appropriately
- Create `TagObject` for each group with auto-generated description
- Allow override via config

#### 2.4 — Extract Handler Function References
- Use `runtime.FuncForPC` to get the handler function's name and package
- Use `reflect` to get the handler's type information
- Store function name for display in docs (e.g., `controllers.CreateUser`)
- Attempt to extract the source file and line number for dev mode

#### 2.5 — Detect Middleware Chain
- For each route, identify middleware in the handler chain
- Detect common middleware patterns:
  - Auth middleware → mark route as requiring authentication
  - Rate limiting → note in description
  - Sentinel middleware → auto-add security scheme
- Store middleware metadata per route for later use

#### 2.6 — Build Route Metadata Registry
- Create `RouteMetadata` struct:
  - `Method`, `Path`, `OpenAPIPath`
  - `HandlerName`, `HandlerFunc`
  - `PathParams []ParameterObject`
  - `Tags []string`
  - `Middleware []string`
  - `RequiresAuth bool`
- Build a registry: `map[string]*RouteMetadata` keyed by `METHOD:PATH`
- Implement `Introspect()` method on `GinDocs` that populates this registry
- Write tests: `TestIntrospectRoutes`, `TestPathParamExtraction`, `TestRouteGrouping`

---

## Phase 3: Struct Reflection & Schema Engine

**Goal:** Build a reflection engine that converts Go structs into OpenAPI schema objects by reading struct tags (`json`, `binding`, `form`, `query`, `uri`, `gorm`, `docs`).

### Tasks

#### 3.1 — Go Type → OpenAPI Schema Converter (`gindocs/schemas.go`)
- Implement `typeToSchema(t reflect.Type) *SchemaObject` recursive converter
- Handle primitive types:
  - `string` → `{type: "string"}`
  - `int`, `int32`, `int64` → `{type: "integer", format: "int32"/"int64"}`
  - `uint`, `uint32`, `uint64` → `{type: "integer", format: "int32"/"int64", minimum: 0}`
  - `float32`, `float64` → `{type: "number", format: "float"/"double"}`
  - `bool` → `{type: "boolean"}`
  - `time.Time` → `{type: "string", format: "date-time"}`
  - `uuid.UUID` → `{type: "string", format: "uuid"}`
- Handle complex types:
  - `[]T` → `{type: "array", items: schemaOf(T)}`
  - `map[K]V` → `{type: "object", additionalProperties: schemaOf(V)}`
  - `*T` → `schemaOf(T)` with `nullable: true`
  - Struct → `{type: "object", properties: {...}}`
- Use type registry for deduplication and `$ref` references

#### 3.2 — Struct Tag Parser (`gindocs/tags.go`)
- Parse `json:"name,omitempty"` → property name, omitempty flag
- Parse `binding:"required,min=1,max=100,email,oneof=active inactive"` → required flag, validation rules
  - `required` → add to `required` array
  - `min`, `max` → `minimum`, `maximum` for numbers; `minLength`, `maxLength` for strings
  - `email` → `format: "email"`
  - `oneof=a b c` → `enum: ["a", "b", "c"]`
  - `gt`, `gte`, `lt`, `lte` → `exclusiveMinimum`, `minimum`, `exclusiveMaximum`, `maximum`
- Parse `form:"field_name"` → alternative field name for form data
- Parse `query:"field_name"` → query parameter
- Parse `uri:"field_name"` → path parameter
- Parse `gorm:"size:200;uniqueIndex;not null;default:'pending'"` → `maxLength`, description notes
- Parse custom `docs:"..."` tag (see Phase 9)

#### 3.3 — Struct Field Iterator
- Implement `structToSchema(t reflect.Type) *SchemaObject`
- Iterate over all exported fields
- Handle embedded structs (flatten fields into parent)
- Handle `json:"-"` (skip field)
- Handle `json:",omitempty"` (mark as optional)
- Build `properties` map and `required` array
- Support nested struct references via `$ref` to components

#### 3.4 — Type Registry & Component Refs (`gindocs/registry.go`)
- Implement `TypeRegistry` that maps `reflect.Type` → schema name
- When a struct is first encountered, generate its schema and store in `components/schemas`
- On subsequent encounters, return a `$ref: "#/components/schemas/TypeName"` reference
- Handle naming conflicts (same name, different packages) by prefixing with package name
- Handle anonymous structs by generating a name from context

#### 3.5 — Detect Request Types from Handler Patterns (`gindocs/reflect.go`)
- **Pattern 1: ShouldBindJSON / ShouldBind** — detect the struct type passed to bind
  - Use a registration mechanism: `gindocs.RegisterRoute("POST /users", RequestType{}, ResponseType{})`
  - Or: analyze handler function parameter types if using a typed handler pattern
- **Pattern 2: Convention-based** — look for structs named `CreateUserRequest`, `UpdateUserResponse` in the same package
- **Pattern 3: Middleware-based** — provide a `DocRequestBody[T any]()` helper that registers types
- Implement all three patterns with graceful fallback

#### 3.6 — Query Parameter Extraction
- Detect `query:"param_name"` tags in structs used with `c.ShouldBindQuery()`
- Create `ParameterObject` with `In: "query"` for each
- Detect common patterns: `page`, `page_size`, `sort_by`, `sort_order`, `search`, `filter_*`
- Auto-add pagination parameters when detected

#### 3.7 — Response Schema Inference
- For registered response types, generate full response schemas
- Auto-generate common response wrappers:
  - Success: `{data: T}` or `{data: T, message: string}`
  - List: `{data: []T, total: int, page: int, page_size: int}`
  - Error: `{error: string}` or `{error: string, details: [...]}`
- Detect if handler returns `c.JSON(status, gin.H{...})` — mark as `object` with unknown schema
- Allow explicit response type registration

#### 3.8 — Enum Detection
- Scan for `const` blocks that define values for custom types
- Example: `type Status string; const (Active Status = "active"; Inactive Status = "inactive")`
- Map these to `enum` values in the schema
- Use `iota`-based enums for integer types
- Store in a global enum registry for reuse

#### 3.9 — Write Schema Tests
- Test primitive type conversion
- Test struct with all tag types
- Test nested structs and `$ref` resolution
- Test embedded structs
- Test slices, maps, pointers
- Test enum detection
- Test binding tag → validation rule mapping
- Test circular reference handling (e.g., User has Posts, Post has User)

---

## Phase 4: OpenAPI 3.1 Spec Generator

**Goal:** Assemble the introspected routes and reflected schemas into a complete, valid OpenAPI 3.1 specification document.

### Tasks

#### 4.1 — Build Spec Root (`gindocs/openapi.go`)
- Create `generateSpec()` method on `GinDocs`
- Build root `OpenAPISpec`:
  - `openapi: "3.1.0"`
  - `info` from config (title, description, version, contact, license)
  - `servers` from config (default: current host)
  - `paths` — populated from routes
  - `components` — schemas, security schemes
  - `tags` — from route groups
  - `security` — global security requirements

#### 4.2 — Generate Path Items
- For each route in the metadata registry:
  - Create or get `PathItemObject` for the OpenAPI path
  - Create `OperationObject` for the HTTP method
  - Set `operationId` from handler name (e.g., `createUser`)
  - Set `tags` from route group
  - Set `summary` and `description` from inference engine (Phase 7)
  - Set `parameters` from path params + query params
  - Set `requestBody` if method is POST/PUT/PATCH
  - Set `responses` from response schema or defaults
  - Apply any overrides (Phase 9)

#### 4.3 — Generate Request Bodies
- For routes with registered request types:
  - Create `RequestBodyObject` with `application/json` media type
  - Reference the schema via `$ref`
  - Set `required: true` for POST/PUT
- For routes with `multipart/form-data` (file uploads):
  - Detect `*multipart.FileHeader` fields
  - Create appropriate schema with `format: "binary"`

#### 4.4 — Generate Response Objects
- For each operation, generate responses:
  - Primary success response (200/201/204 based on method)
  - Common error responses (400, 401, 403, 404, 500)
  - Use registered response types when available
  - Fall back to generic schemas when not
- Create reusable error response components

#### 4.5 — Generate Security Schemes
- Based on `Config.Auth`:
  - `Bearer` → `{type: "http", scheme: "bearer", bearerFormat: "JWT"}`
  - `APIKey` → `{type: "apiKey", in: "header", name: "X-API-Key"}`
  - `Basic` → `{type: "http", scheme: "basic"}`
- Auto-detect from middleware (e.g., if auth middleware is present)
- Apply security requirements to routes that have auth middleware

#### 4.6 — Generate Tags with Descriptions
- Create `TagObject` for each route group
- Auto-generate descriptions: "Operations related to {tag_name}"
- Support custom descriptions via config
- Sort tags alphabetically
- Add `externalDocs` if configured

#### 4.7 — Spec Serialization
- Implement `ToJSON() ([]byte, error)` — serialize to JSON with proper indentation
- Implement `ToYAML() ([]byte, error)` — serialize to YAML (use `gopkg.in/yaml.v3`)
- Ensure all `$ref` paths are valid
- Validate the spec against OpenAPI 3.1 rules (basic validation)
- Cache the serialized spec, invalidate on `Refresh()`

#### 4.8 — Write Spec Generation Tests
- Test minimal spec generation (empty router)
- Test spec with CRUD routes
- Test spec with path params, query params, request bodies
- Test spec with security schemes
- Test JSON and YAML serialization
- Validate generated spec against OpenAPI 3.1 schema (if validator available)
- Test that excluded routes don't appear

---

## Phase 5: Swagger UI Integration

**Goal:** Embed Swagger UI and serve it as the default documentation interface with "Try It" support.

### Tasks

#### 5.1 — Embed Swagger UI Assets
- Download Swagger UI dist files (latest stable)
- Use `go:embed` to embed them in the binary
- Create `gindocs/ui_swagger.go` with embedded filesystem
- Ensure assets are served efficiently with proper cache headers

#### 5.2 — Create Swagger UI HTML Template
- Create an HTML template that loads Swagger UI
- Inject the OpenAPI spec URL (`/docs/openapi.json`)
- Configure Swagger UI options:
  - `deepLinking: true`
  - `displayRequestDuration: true`
  - `filter: true` (search bar)
  - `showExtensions: true`
  - `tryItOutEnabled` based on `Config.ReadOnly`
  - `persistAuthorization: true`
- Support dark mode via CSS injection
- Add custom CSS for branding (logo, colors)

#### 5.3 — Serve Swagger UI Handler
- Implement handler for `GET /docs` that serves the HTML page
- Implement handler for `GET /docs/assets/*` that serves embedded assets
- Set proper `Content-Type` headers
- Add `Cache-Control` headers (long cache for assets, no-cache for HTML in dev)

#### 5.4 — Configure "Try It" Authentication
- Pre-configure auth in Swagger UI from `Config.Auth`
- Support Bearer token input
- Support API key header input
- Support Basic auth input
- Auto-populate server URLs from `Config.Servers`

#### 5.5 — Add Custom Branding
- Inject custom logo if `Config.Logo` is set
- Inject custom title
- Add "Powered by Gin Docs" footer (optional, removable via config)
- Support custom CSS injection via `Config.CustomCSS`

#### 5.6 — Test Swagger UI
- Test that `/docs` returns HTML with Swagger UI
- Test that `/docs/openapi.json` returns valid JSON
- Test that assets are served correctly
- Test dark mode toggle
- Test "Try It" is disabled when `ReadOnly: true`

---

## Phase 6: Scalar UI Integration

**Goal:** Add Scalar as an alternative, modern documentation UI that can be selected via config.

### Tasks

#### 6.1 — Embed Scalar Assets
- Scalar can be loaded via CDN or embedded
- Create `gindocs/ui_scalar.go`
- Use Scalar's standalone HTML approach (single HTML file with CDN reference)
- Alternatively, embed the JS bundle for offline use

#### 6.2 — Create Scalar HTML Template
- Create HTML template that initializes Scalar
- Configure Scalar options:
  - `spec` → point to `/docs/openapi.json`
  - `theme` → default dark or light
  - `layout` → "modern" (sidebar) or "classic"
  - `hideModels` → false
  - `searchHotKey` → "k"
  - `showSidebar` → true
  - Authentication configuration
- Support proxy for "Try It" requests if needed

#### 6.3 — Serve Scalar UI Handler
- If `Config.UI == UIScalar`, serve Scalar instead of Swagger
- Support query param override: `/docs?ui=scalar` or `/docs?ui=swagger`
- Both UIs share the same OpenAPI spec endpoint

#### 6.4 — UI Switching Mechanism
- Implement seamless switching between UIs
- Persist preference in a cookie or query param
- Add a small toggle button in both UIs to switch
- Ensure both UIs work with the same spec

#### 6.5 — Test Scalar UI
- Test that `/docs?ui=scalar` returns Scalar HTML
- Test that config-based UI selection works
- Test that "Try It" works in Scalar
- Test theme configuration

---

## Phase 7: Smart Inference Engine

**Goal:** Automatically generate human-readable descriptions, summaries, and status codes from route patterns and handler names.

### Tasks

#### 7.1 — HTTP Method + Path → Summary (`gindocs/inference.go`)
- Implement rules:
  - `GET /users` → "List all users"
  - `GET /users/:id` → "Get a user by ID"
  - `POST /users` → "Create a new user"
  - `PUT /users/:id` → "Update a user by ID"
  - `PATCH /users/:id` → "Partially update a user by ID"
  - `DELETE /users/:id` → "Delete a user by ID"
  - `GET /users/:id/posts` → "List posts for a user"
  - `POST /users/:id/posts` → "Create a post for a user"
- Handle nested resources up to 3 levels deep
- Singularize/pluralize resource names intelligently
- Handle special paths: `/login`, `/register`, `/health`, `/metrics`, `/search`

#### 7.2 — Handler Name → Description
- Parse handler function name: `controllers.CreateUser` → "Create User"
- Use CamelCase splitting: `GetUserPosts` → "Get User Posts"
- Fall back to method + path inference if handler name is generic (e.g., `func1`)

#### 7.3 — Status Code Inference
- Default success codes:
  - `GET` → 200
  - `POST` → 201
  - `PUT/PATCH` → 200
  - `DELETE` → 204 (or 200 if response body present)
- Default error codes for all routes:
  - 400 Bad Request (for routes with request bodies)
  - 401 Unauthorized (for authenticated routes)
  - 403 Forbidden (for authenticated routes)
  - 404 Not Found (for routes with path params)
  - 500 Internal Server Error (always)

#### 7.4 — Parameter Description Inference
- Common parameter names:
  - `id` → "Unique identifier"
  - `page` → "Page number for pagination"
  - `page_size` / `limit` → "Number of items per page"
  - `sort_by` → "Field to sort by"
  - `sort_order` → "Sort direction (asc or desc)"
  - `search` / `q` → "Search query string"
  - `filter_*` → "Filter by {field_name}"
- Infer from field type: `email` field → "Email address"

#### 7.5 — Request/Response Example Generation
- Generate example values based on field types and names:
  - `Name string` → `"John Doe"`
  - `Email string` → `"user@example.com"`
  - `Age int` → `25`
  - `CreatedAt time.Time` → `"2025-01-15T10:30:00Z"`
  - `ID uint` → `1`
  - `Price float64` → `29.99`
  - `IsActive bool` → `true`
- Use enum values for enum fields
- Generate realistic examples for known patterns (phone, URL, UUID)

#### 7.6 — Write Inference Tests
- Test all method + path combinations
- Test handler name parsing
- Test status code inference
- Test parameter descriptions
- Test example generation
- Edge cases: unconventional paths, deeply nested routes

---

## Phase 8: GORM Model Integration

**Goal:** Extract rich schema information from GORM models, including relationships, field constraints, and default values.

### Tasks

#### 8.1 — GORM Tag Parser (`gindocs/gorm.go`)
- Parse `gorm:"..."` tag components:
  - `primarykey` / `primaryKey` → mark as ID field
  - `size:N` → `maxLength: N`
  - `type:text` → note in description
  - `uniqueIndex` → `description: "Must be unique"`
  - `not null` → add to `required`
  - `default:'value'` → `default: "value"`
  - `column:name` → use as property name if no `json` tag
  - `autoIncrement` → `readOnly: true`
  - `autoCreateTime` / `autoUpdateTime` → `readOnly: true`
  - `-` → skip field

#### 8.2 — Relationship Detection
- Detect GORM relationship types:
  - `HasOne` → nested object reference
  - `HasMany` → array of references
  - `BelongsTo` → object reference
  - `Many2Many` → array of references
- Use foreign key tags to link models
- Create proper `$ref` relationships in schemas
- Handle circular references (User → Posts → User) with lazy refs

#### 8.3 — Model Schema Registration
- When `Mount()` receives models via parameter:
  - `gindocs.Mount(router, db, config, &User{}, &Post{}, &Comment{})`
  - Or via config: `Config.Models: []interface{}{&User{}, &Post{}}`
- Register each model in the type registry
- Generate complete schemas with GORM metadata
- Auto-create "Create" and "Update" variants (without ID, timestamps)

#### 8.4 — Schema Variant Generation
- For each model, generate variants:
  - `User` — full schema (for responses)
  - `CreateUser` — without `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt` (for POST body)
  - `UpdateUser` — all fields optional except identifiers (for PUT/PATCH body)
  - `UserList` — pagination wrapper with array of `User`
- Use `$ref` for field reuse between variants

#### 8.5 — Write GORM Integration Tests
- Test GORM tag parsing
- Test relationship detection
- Test schema variant generation
- Test with real GORM models (User, Post, Comment with relationships)
- Test circular reference handling

---

## Phase 9: Override & Customization System

**Goal:** Allow developers to manually override auto-generated documentation with a simple API and custom struct tags.

### Tasks

#### 9.1 — Custom `docs` Struct Tag (`gindocs/tags.go`)
- Define `docs:"..."` tag format:
  - `docs:"description:User's full name"`
  - `docs:"example:John Doe"`
  - `docs:"deprecated"`
  - `docs:"hidden"` — exclude from schema
  - `docs:"enum:active|inactive|suspended"`
  - `docs:"format:email"`
  - `docs:"minimum:0,maximum:100"`
  - Multiple values: `docs:"description:Full name,example:John Doe"`
- Parse and apply these overrides during schema generation
- These take precedence over auto-inferred values

#### 9.2 — Route-Level Override API
- Implement fluent API for per-route overrides:
  ```go
  docs := gindocs.Mount(router, db, config)
  docs.Route("POST /users").
      Summary("Register a new user").
      Description("Creates a user account and sends a welcome email").
      RequestBody(CreateUserInput{}).
      Response(201, User{}, "User created successfully").
      Response(409, nil, "Email already exists").
      Tags("Authentication").
      Deprecated(false)
  ```
- Store overrides in the `overrides` map
- Apply during spec generation (overrides win over inference)

#### 9.3 — Group-Level Overrides
- Override docs for entire route groups:
  ```go
  docs.Group("/api/admin/*").
      Tags("Admin").
      Security("bearerAuth").
      Description("Administrative endpoints — requires admin role")
  ```
- Apply to all matching routes

#### 9.4 — Custom Markdown Sections
- Support adding custom documentation sections:
  ```go
  config.CustomSections = []gindocs.Section{
      {Title: "Authentication", Content: "## How to authenticate\n..."},
      {Title: "Rate Limits", Content: "..."},
      {Title: "Changelog", ContentFile: "CHANGELOG.md"},
  }
  ```
- Render in Scalar's sidebar or Swagger UI's description
- Support reading from files

#### 9.5 — Branding Configuration
- Custom logo (URL or embedded)
- Custom favicon
- Custom colors/theme
- Custom footer text
- "Powered by" toggle

#### 9.6 — Route Visibility Control
- `ExcludeRoutes: []string{"/health", "/metrics", "/debug/*"}`
- `ExcludePrefixes: []string{"/internal/"}`
- Support glob patterns
- `docs.Route("GET /debug/pprof").Hide()`
- Default exclusions: the docs routes themselves

#### 9.7 — Write Override Tests
- Test `docs` tag parsing and application
- Test route-level overrides
- Test group-level overrides
- Test custom sections rendering
- Test route exclusion/visibility

---

## Phase 10: Export & Integration Features

**Goal:** Add export capabilities, integrations with other tools, and production-ready features.

### Tasks

#### 10.1 — OpenAPI Spec Download Endpoints
- `GET /docs/openapi.json` — JSON format (already exists, polish it)
- `GET /docs/openapi.yaml` — YAML format
- Add `Content-Disposition` header for file download
- Add version in filename: `openapi-v1.0.0.json`
- Add ETag for caching

#### 10.2 — Postman Collection Export
- `GET /docs/export/postman` — generates Postman v2.1 collection
- Map OpenAPI operations to Postman requests
- Include examples, auth configuration
- Support environment variables for base URL and auth tokens

#### 10.3 — Insomnia Export
- `GET /docs/export/insomnia` — generates Insomnia v4 export
- Similar mapping as Postman
- Include workspace and environment setup

#### 10.4 — Dev Mode Live Reload
- When `Config.DevMode` is true:
  - Re-run introspection on every page load
  - Add auto-refresh script to the UI HTML
  - Detect route changes and regenerate spec
- When false:
  - Generate spec once at startup
  - Serve from cache

#### 10.5 — Sentinel & GORM Studio Integration
- If Sentinel is detected (mounted on same router):
  - Auto-document security middleware in the spec
  - Link to Sentinel dashboard from docs
  - Add security notes to protected routes
- If GORM Studio is mounted:
  - Add link to Studio from docs
  - Cross-reference model schemas

---

## Phase 11: Demo Application & Examples

**Goal:** Create a realistic demo application and example projects that showcase all features.

### Tasks

#### 11.1 — Demo Application (`main.go`)
- Build a realistic blog API with:
  - User registration/login
  - CRUD for Posts, Comments, Tags
  - File upload endpoint
  - Pagination, filtering, sorting
  - Auth middleware on write endpoints
- Mount Gin Docs with full configuration
- Mount GORM Studio alongside (show integration)
- Include seed data

#### 11.2 — Basic Example (`examples/basic/`)
- Minimal setup: 3 models, 5 routes, one-line mount
- Show default behavior with zero config
- Include README with setup instructions

#### 11.3 — Full Example (`examples/full/`)
- All features configured:
  - Custom branding
  - Route overrides
  - Custom sections (authentication guide, changelog)
  - Scalar UI
  - Auth configuration
  - GORM model registration
- Include README

#### 11.4 — With Sentinel Example (`examples/with-sentinel/`)
- Show Gin Docs + Sentinel working together
- Auto-documented security middleware
- Cross-linking between dashboards

#### 11.5 — Screenshots & Demo GIF
- Capture screenshots of both Swagger and Scalar UIs
- Create a GIF/video showing:
  - Mounting in one line
  - Browsing docs
  - "Try It" in action
  - Switching between UIs

---

## Phase 12: Testing, Documentation & Release

**Goal:** Comprehensive testing, documentation, CI/CD, and publishing.

### Tasks

#### 12.1 — Unit Tests
- Achieve 80%+ code coverage
- Test every public function
- Test edge cases: empty router, no models, circular refs, giant schemas
- Test configuration defaults and validation
- Use table-driven tests throughout

#### 12.2 — Integration Tests
- Full Mount → Generate → Serve cycle
- Test generated spec against OpenAPI 3.1 validator
- Test both UIs render correctly
- Test "Try It" functionality
- Test export endpoints

#### 12.3 — README.md
- Badges: Go version, license, CI status, Go Report Card
- Quick start (3 steps)
- Feature overview with screenshots
- Full configuration reference
- Override API documentation
- FAQ section
- Contributing guide

#### 12.4 — Documentation Site (`docs/`)
- `docs/configuration.md` — full config reference
- `docs/overrides.md` — override API guide
- `docs/struct-tags.md` — all supported struct tags
- `docs/examples.md` — usage examples
- `docs/architecture.md` — how it works internally

#### 12.5 — CONTRIBUTING.md & SECURITY.md
- Contribution guidelines
- Code style guide
- Security reporting process
- Consistent with GORM Studio and Sentinel

#### 12.6 — CI/CD Pipeline (`.github/workflows/`)
- Run tests on push/PR
- Lint with `golangci-lint`
- Build verification
- Tag-based releases with goreleaser
- Publish to pkg.go.dev

#### 12.7 — Release v0.1.0
- Tag initial release
- Ensure `go get github.com/MUKE-coder/gin-docs/gindocs` works
- Announce on Go communities
- Submit to awesome-go list
