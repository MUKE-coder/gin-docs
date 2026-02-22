# Gin Docs — Auto API Documentation for Go/Gin

## What Is Gin Docs?

Gin Docs is a **zero-annotation API documentation generator** for Go applications built with **Gin** and **GORM**. It automatically inspects your Gin route tree, reads request/response structs via reflection, and generates a fully interactive API documentation page — mountable with a single function call.

```go
gindocs.Mount(router, db, gindocs.Config{})
// Docs → http://localhost:8080/docs
```

## The Problem

Go developers using Gin face a painful documentation workflow:

1. **swaggo/swag requires comment annotations everywhere.** Every handler needs verbose `// @Summary`, `// @Param`, `// @Success` comments that clutter the codebase and must be manually kept in sync with code changes. When they drift (and they always drift), docs become misleading — worse than no docs at all.

2. **No reflection-based solution exists for Gin.** Frameworks like FastAPI (Python), NestJS (TypeScript), and Hono generate docs automatically from types. Go developers don't have an equivalent. They either spend hours writing annotations, or they simply don't document their APIs.

3. **Swagger UI is the only option.** Developers who want modern, beautiful docs (like Scalar, Redocly, or Stoplight) have to manually generate OpenAPI specs and host separate UI tooling. There's no drop-in solution.

4. **GORM model awareness is nonexistent.** Existing tools don't understand GORM models, relationships, or tags. Developers end up duplicating struct definitions — one for GORM, one for docs.

5. **"Try it" functionality requires extra setup.** Testing endpoints from docs typically requires hosting Swagger UI separately and configuring CORS, auth, etc.

## The Solution

Gin Docs solves all of these by:

- **Introspecting the Gin route tree** at startup to discover all registered routes, methods, and middleware
- **Using Go reflection** to read request body structs, query parameters, path parameters, and response types from handler signatures and struct tags (`json`, `binding`, `form`, `query`, `gorm`)
- **Generating a compliant OpenAPI 3.1 spec** automatically — no comments, no annotations, no code generation step
- **Serving an interactive documentation UI** (Swagger UI or Scalar, user's choice) directly from your app
- **Understanding GORM models** to auto-document response schemas, relationships, and field validations
- **Providing a "Try It" playground** so developers and API consumers can test endpoints directly from the docs

## Key Features

### Core Documentation Engine
- **Zero-annotation OpenAPI generation** from Gin routes and Go structs
- **Automatic request/response schema detection** via reflection
- **GORM model awareness** — reads `gorm` tags for field types, relationships, validations
- **Binding tag support** — reads `binding:"required"`, `validate`, `form`, `query`, `uri` tags
- **Enum detection** — auto-detects const blocks as enum values
- **Nested struct resolution** — handles embedded structs, pointers, slices, maps

### Documentation UI
- **Dual UI support** — Swagger UI (default) or Scalar (modern, beautiful)
- **Switchable at runtime** — toggle between UIs from config or query param
- **Built-in "Try It" playground** — test any endpoint directly from the docs
- **Authentication support** — configure auth headers (Bearer, API Key, Basic) for "Try It"
- **Dark/Light mode** — both UIs support theme switching
- **Responsive design** — works on mobile for on-the-go API reference

### Developer Experience
- **One-line mount** — `gindocs.Mount(router, db, gindocs.Config{})`
- **Route grouping** — auto-groups endpoints by Gin route groups or custom tags
- **Middleware detection** — shows which middleware applies to each route (auth, rate limit, etc.)
- **Live reload in dev** — re-introspects routes on page refresh in dev mode
- **OpenAPI spec export** — download the generated spec as JSON or YAML at `/docs/openapi.json`
- **Versioning support** — document multiple API versions side by side

### Smart Inference
- **HTTP method semantics** — infers descriptions from method + path (e.g., `DELETE /users/:id` → "Delete a user by ID")
- **Status code inference** — auto-generates response codes based on method (POST → 201, GET → 200, DELETE → 204)
- **Error response schemas** — auto-includes common Gin error responses (400, 401, 404, 500)
- **Parameter detection** — reads `:id` path params, `c.Query()`, `c.Bind()` calls

### Customization & Overrides
- **Struct tag overrides** — use `docs:"description:...,example:...,deprecated"` for fine-tuning
- **Route-level overrides** — programmatically override any auto-generated docs per route
- **Custom sections** — add markdown-based guides, changelogs, and authentication docs
- **Branding** — custom logo, title, description, contact info, license in the spec
- **Exclude routes** — hide internal/debug routes from docs

### Integration
- **GORM model schemas** — pass your models and they become reusable schema components
- **Sentinel integration** — if using Sentinel, auto-documents security middleware
- **GORM Studio link** — optionally link to GORM Studio for database browsing
- **Webhook for spec changes** — notify CI/CD when the spec changes (for contract testing)
- **Postman/Insomnia export** — download collections for popular API clients

## Benefits

| Benefit | Description |
|---------|-------------|
| **Zero maintenance** | Docs stay in sync because they're generated from the actual code |
| **Zero clutter** | No annotation comments polluting your handlers |
| **Zero setup** | One line to mount, works with any existing Gin app |
| **Beautiful by default** | Scalar UI is modern and polished out of the box |
| **Production ready** | Disable "Try It" in production, keep docs read-only |
| **Standards compliant** | Generates valid OpenAPI 3.1 specs usable anywhere |
| **Team friendly** | New team members can explore the entire API instantly |

## Target Users

- **Go developers** building REST APIs with Gin + GORM
- **Backend teams** that need API documentation without the overhead
- **Startups** that move fast and can't afford to maintain separate docs
- **Open source projects** that want great docs with minimal effort

## How It Differs from Existing Tools

| Feature | swaggo/swag | go-swagger | Gin Docs |
|---------|-------------|------------|----------|
| Annotations required | Yes (verbose) | Yes (verbose) | **No** |
| Reflection-based | No | Partial | **Yes** |
| GORM awareness | No | No | **Yes** |
| Built-in UI | Swagger only | Swagger only | **Swagger + Scalar** |
| One-line mount | No | No | **Yes** |
| "Try It" built-in | Partial | Partial | **Yes** |
| Live reload | No | No | **Yes** |
| Go struct tag driven | No | Partial | **Yes** |

## Technical Approach

1. **Route Introspection** — Use Gin's `Routes()` method to enumerate all registered routes at mount time
2. **Handler Analysis** — Use reflection to inspect handler function signatures and extract bound structs
3. **Struct Tag Parsing** — Parse `json`, `binding`, `form`, `query`, `uri`, `gorm`, `docs` tags to build schemas
4. **OpenAPI Generation** — Build a compliant OpenAPI 3.1 document from the introspected metadata
5. **UI Embedding** — Serve Swagger UI or Scalar via `go:embed` with the generated spec injected
6. **Type Registry** — Maintain a registry of Go types → OpenAPI schemas for deduplication and `$ref` usage

## Repository Structure

```
gin-docs/
├── gindocs/
│   ├── mount.go          # Mount() entry point
│   ├── config.go         # Configuration types
│   ├── introspect.go     # Route tree introspection
│   ├── reflect.go        # Struct reflection and tag parsing
│   ├── openapi.go        # OpenAPI 3.1 spec generation
│   ├── schemas.go        # Type → Schema conversion
│   ├── inference.go      # Smart description/status inference
│   ├── overrides.go      # Manual override system
│   ├── ui_swagger.go     # Swagger UI embedding
│   ├── ui_scalar.go      # Scalar UI embedding
│   ├── handlers.go       # HTTP handlers (serve UI, spec, assets)
│   ├── middleware.go      # Route metadata extraction middleware
│   ├── export.go         # Postman/Insomnia collection export
│   ├── tags.go           # docs:"..." struct tag parser
│   └── gorm.go           # GORM model schema extraction
├── docs/
│   ├── README.md
│   ├── configuration.md
│   └── examples.md
├── examples/
│   ├── basic/
│   ├── full/
│   └── with-sentinel/
├── go.mod
├── go.sum
├── main.go               # Demo application
├── README.md
├── LICENSE
├── CONTRIBUTING.md
├── SECURITY.md
└── Makefile
```

## License

MIT — consistent with GORM Studio and Sentinel.
