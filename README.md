# Gin Docs

[![Go Reference](https://pkg.go.dev/badge/github.com/MUKE-coder/gin-docs.svg)](https://pkg.go.dev/github.com/MUKE-coder/gin-docs)
[![Go Report Card](https://goreportcard.com/badge/github.com/MUKE-coder/gin-docs)](https://goreportcard.com/report/github.com/MUKE-coder/gin-docs)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

**Zero-annotation API documentation generator for Go applications using Gin and GORM.**

Gin Docs automatically introspects your Gin router, reads struct tags via reflection, generates a valid OpenAPI 3.1 specification, and serves interactive documentation — all with a single function call.

## Features

- **Zero annotations** — no comment-based docs. Everything is inferred from code.
- **One-line mount** — `gindocs.Mount(router, db)` and you're done.
- **Two UI options** — Swagger UI (default) and Scalar, switchable via config or query param.
- **Full OpenAPI 3.1** — valid spec that works with any tooling.
- **GORM integration** — auto-generates Create/Update schema variants from your models.
- **Smart inference** — auto-generates summaries, status codes, examples, and parameter descriptions.
- **Fluent override API** — customize any route's documentation with a builder pattern.
- **Export support** — download as Postman or Insomnia collections.
- **Production-ready** — concurrent-safe, no panics, proper caching.

## Quick Start

```bash
go get github.com/MUKE-coder/gin-docs
```

```go
package main

import (
    "net/http"

    "github.com/MUKE-coder/gin-docs/gindocs"
    "github.com/gin-gonic/gin"
)

type User struct {
    ID    uint   `json:"id" gorm:"primarykey"`
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
}

func main() {
    r := gin.Default()

    r.GET("/api/users", func(c *gin.Context) {
        c.JSON(http.StatusOK, []User{{ID: 1, Name: "John", Email: "john@example.com"}})
    })

    r.POST("/api/users", func(c *gin.Context) {
        var u User
        if err := c.ShouldBindJSON(&u); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        u.ID = 1
        c.JSON(http.StatusCreated, u)
    })

    r.GET("/api/users/:id", func(c *gin.Context) {
        c.JSON(http.StatusOK, User{ID: 1, Name: "John", Email: "john@example.com"})
    })

    // One line to add docs!
    gindocs.Mount(r, nil)

    r.Run(":8080")
    // Visit http://localhost:8080/docs
}
```

## Configuration

```go
gindocs.Mount(router, db, gindocs.Config{
    Title:       "My API",
    Description: "API description in markdown.",
    Version:     "1.0.0",
    UI:          gindocs.UIScalar, // or gindocs.UISwagger (default)
    DevMode:     true,             // Re-introspect on every request
    ReadOnly:    false,            // Disable "Try It" if true
    Auth: gindocs.AuthConfig{
        Type:         gindocs.AuthBearer,
        BearerFormat: "JWT",
    },
    Servers: []gindocs.ServerInfo{
        {URL: "http://localhost:8080", Description: "Local"},
    },
    Models: []interface{}{User{}, Post{}}, // GORM models
})
```

### Config Reference

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `Prefix` | `string` | `"/docs"` | URL prefix for docs endpoints |
| `Title` | `string` | `"API Documentation"` | API title |
| `Description` | `string` | `""` | API description (markdown) |
| `Version` | `string` | `"1.0.0"` | API version |
| `UI` | `UIType` | `UISwagger` | UI to serve (`UISwagger` or `UIScalar`) |
| `DevMode` | `bool` | `false` | Re-generate spec on every request |
| `ReadOnly` | `bool` | `false` | Disable "Try It" functionality |
| `Auth` | `AuthConfig` | `AuthNone` | Authentication config for UI |
| `Servers` | `[]ServerInfo` | `[]` | API server URLs |
| `Models` | `[]interface{}` | `[]` | GORM models to register as schemas |
| `ExcludeRoutes` | `[]string` | `[]` | Glob patterns to exclude |
| `ExcludePrefixes` | `[]string` | `[]` | Path prefixes to exclude |
| `CustomSections` | `[]Section` | `[]` | Extra docs sections (markdown) |
| `CustomCSS` | `string` | `""` | Custom CSS for the UI |

## Struct Tags

Gin Docs reads these struct tags to generate accurate schemas:

```go
type User struct {
    ID    uint   `json:"id" gorm:"primarykey"`
    Name  string `json:"name" binding:"required,min=2,max=100"`
    Email string `json:"email" binding:"required,email" gorm:"size:200;uniqueIndex"`
    Role  string `json:"role" binding:"oneof=admin user" gorm:"default:'user'"`
    Bio   string `json:"bio" docs:"description:User biography,example:A Go developer"`
}
```

| Tag | Effect |
|-----|--------|
| `json:"name,omitempty"` | Property name, marks as optional |
| `json:"-"` | Skip field |
| `binding:"required"` | Adds to `required` array |
| `binding:"email"` | Sets `format: "email"` |
| `binding:"oneof=a b c"` | Sets `enum` |
| `binding:"min=N,max=M"` | Sets `minimum`/`maximum` or `minLength`/`maxLength` |
| `gorm:"primarykey"` | Marks as `readOnly` |
| `gorm:"size:N"` | Sets `maxLength` |
| `gorm:"uniqueIndex"` | Adds "Must be unique" to description |
| `gorm:"default:'val'"` | Sets `default` |
| `gorm:"autoCreateTime"` | Marks as `readOnly` |
| `docs:"description:...,example:...,deprecated,hidden"` | Direct schema control |

## Route Overrides

Customize documentation for specific routes:

```go
docs := gindocs.Mount(r, nil, config)

docs.Route("POST /api/users").
    Summary("Register a new user").
    Description("Creates a user account.").
    RequestBody(CreateUserInput{}).
    Response(201, User{}, "User created").
    Response(400, nil, "Validation error").
    Tags("Authentication")

docs.Group("/api/admin/*").
    Tags("Admin").
    Security("bearerAuth")
```

## Doc Middleware

Document routes inline with a middleware helper:

```go
r.POST("/api/users",
    gindocs.Doc(gindocs.DocConfig{
        Summary:     "Create user",
        RequestBody: CreateUserInput{},
        Response:    User{},
    }),
    createUserHandler,
)
```

## GORM Models

When you register GORM models, Gin Docs automatically generates three schema variants:

- **`User`** — full model (for responses, includes ID and timestamps)
- **`CreateUser`** — without ID, CreatedAt, UpdatedAt (for request bodies)
- **`UpdateUser`** — all fields optional (for PATCH requests)

```go
gindocs.Mount(r, db, gindocs.Config{
    Models: []interface{}{User{}, Post{}, Comment{}},
})
```

## UI Switching

Switch between Swagger UI and Scalar:

```go
// Via config
gindocs.Config{UI: gindocs.UIScalar}

// Via query parameter
// http://localhost:8080/docs?ui=scalar
// http://localhost:8080/docs?ui=swagger
```

## Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/docs` | Documentation UI |
| GET | `/docs/openapi.json` | OpenAPI 3.1 spec (JSON) |
| GET | `/docs/openapi.yaml` | OpenAPI 3.1 spec (YAML) |
| GET | `/docs/export/postman` | Postman v2.1 collection |
| GET | `/docs/export/insomnia` | Insomnia v4 export |

## Examples

- [Basic example](examples/basic/main.go) — minimal setup
- [Full example](examples/full/main.go) — all features configured
- [Demo app](main.go) — realistic blog API

## License

MIT License. See [LICENSE](LICENSE) for details.
