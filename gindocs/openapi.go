package gindocs

import (
	"sort"
	"strings"
)

// assembleSpec builds a complete OpenAPI 3.1 specification from discovered routes,
// registered schemas, and configuration.
func (gd *GinDocs) assembleSpec() *OpenAPISpec {
	title := gd.config.Title
	if title == "" {
		title = "API Documentation"
	}

	spec := &OpenAPISpec{
		OpenAPI: "3.1.0",
		Info: InfoObject{
			Title:       title,
			Description: gd.config.Description,
			Version:     gd.config.Version,
		},
		Paths: make(map[string]*PathItem),
		Components: &ComponentsObject{
			Schemas: make(map[string]*SchemaObject),
		},
	}

	// Add contact info.
	if gd.config.Contact != (ContactInfo{}) {
		spec.Info.Contact = &ContactObject{
			Name:  gd.config.Contact.Name,
			URL:   gd.config.Contact.URL,
			Email: gd.config.Contact.Email,
		}
	}

	// Add license info.
	if gd.config.License != (LicenseInfo{}) {
		spec.Info.License = &LicenseObject{
			Name: gd.config.License.Name,
			URL:  gd.config.License.URL,
		}
	}

	// Add servers.
	for _, s := range gd.config.Servers {
		spec.Servers = append(spec.Servers, ServerObject{
			URL:         s.URL,
			Description: s.Description,
		})
	}

	// Add security schemes based on config.
	if gd.config.Auth.Type != AuthNone {
		spec.Components.SecuritySchemes = make(map[string]*SecuritySchemeObject)
		switch gd.config.Auth.Type {
		case AuthBearer:
			scheme := "bearer"
			if gd.config.Auth.Scheme != "" {
				scheme = gd.config.Auth.Scheme
			}
			spec.Components.SecuritySchemes["bearerAuth"] = &SecuritySchemeObject{
				Type:         "http",
				Scheme:       scheme,
				BearerFormat: gd.config.Auth.BearerFormat,
			}
		case AuthAPIKey:
			name := "X-API-Key"
			if gd.config.Auth.Name != "" {
				name = gd.config.Auth.Name
			}
			in := "header"
			if gd.config.Auth.In != "" {
				in = gd.config.Auth.In
			}
			spec.Components.SecuritySchemes["apiKeyAuth"] = &SecuritySchemeObject{
				Type: "apiKey",
				Name: name,
				In:   in,
			}
		case AuthBasic:
			spec.Components.SecuritySchemes["basicAuth"] = &SecuritySchemeObject{
				Type:   "http",
				Scheme: "basic",
			}
		}
	}

	// Register GORM models as schemas.
	gd.registerGORMModels()

	// Introspect routes.
	routes := gd.introspect()

	// Build operations for each route.
	tagSet := make(map[string]bool)

	for _, route := range routes {
		pathItem, ok := spec.Paths[route.OpenAPIPath]
		if !ok {
			pathItem = &PathItem{}
			spec.Paths[route.OpenAPIPath] = pathItem
		}

		op := gd.buildOperation(route)

		pathItem.SetOperation(route.Method, op)

		for _, tag := range op.Tags {
			tagSet[tag] = true
		}
	}

	// Build sorted tag list.
	var tagNames []string
	for tag := range tagSet {
		tagNames = append(tagNames, tag)
	}
	sort.Strings(tagNames)
	for _, name := range tagNames {
		spec.Tags = append(spec.Tags, TagObject{Name: name})
	}

	// Copy registered schemas to components.
	if gd.registry != nil {
		for name, schema := range gd.registry.All() {
			spec.Components.Schemas[name] = schema
		}
	}

	return spec
}

// buildOperation creates an OperationObject for a route.
func (gd *GinDocs) buildOperation(route RouteMetadata) *OperationObject {
	op := &OperationObject{
		Tags:        route.Tags,
		Summary:     generateSummary(route.Method, route.Path),
		OperationID: generateOperationID(route.Method, route.Path),
		Responses:   make(map[string]*Response),
	}

	// Add path parameters.
	for _, param := range route.PathParams {
		op.Parameters = append(op.Parameters, ParameterObject{
			Name:        param,
			In:          "path",
			Required:    true,
			Description: inferParamDescription(param),
			Schema:      inferParamSchema(param),
		})
	}

	// Add inferred query parameters.
	queryParams := inferQueryParams(route.Method, route.Path)
	op.Parameters = append(op.Parameters, queryParams...)

	// Infer response status codes.
	statusCodes := inferStatusCodes(route.Method, route.PathParams)
	for code, desc := range statusCodes {
		op.Responses[code] = &Response{
			Description: desc,
		}
	}

	// Apply route and group overrides.
	gd.applyRouteOverrides(route.Method, route.Path, op)

	return op
}

// inferParamDescription generates a description for a path parameter.
func inferParamDescription(param string) string {
	lower := strings.ToLower(param)
	switch {
	case lower == "id":
		return "Unique identifier"
	case strings.HasSuffix(lower, "_id") || strings.HasSuffix(lower, "id"):
		resource := strings.TrimSuffix(strings.TrimSuffix(lower, "_id"), "id")
		if resource != "" {
			return capitalize(resource) + " identifier"
		}
		return "Resource identifier"
	case lower == "slug":
		return "URL-friendly identifier"
	default:
		return capitalize(param) + " value"
	}
}

// inferParamSchema generates a schema for a path parameter.
func inferParamSchema(param string) *SchemaObject {
	lower := strings.ToLower(param)
	if lower == "id" || strings.HasSuffix(lower, "_id") || strings.HasSuffix(lower, "id") {
		return &SchemaObject{Type: "integer", Format: "int64"}
	}
	return &SchemaObject{Type: "string"}
}

// inferStatusCodes returns appropriate status codes for an HTTP method.
func inferStatusCodes(method string, pathParams []string) map[string]string {
	codes := make(map[string]string)

	switch method {
	case "GET":
		codes["200"] = "Successful response"
	case "POST":
		codes["201"] = "Resource created"
	case "PUT", "PATCH":
		codes["200"] = "Resource updated"
	case "DELETE":
		codes["204"] = "Resource deleted"
	default:
		codes["200"] = "Successful response"
	}

	// Add error responses for methods with request bodies.
	if method == "POST" || method == "PUT" || method == "PATCH" {
		codes["400"] = "Invalid request body"
	}

	// Add 404 for routes with path params.
	if len(pathParams) > 0 {
		codes["404"] = "Resource not found"
	}

	// Always add 500.
	codes["500"] = "Internal server error"

	return codes
}
