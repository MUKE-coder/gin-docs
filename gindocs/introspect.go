package gindocs

import (
	"path"
	"strings"
)

// RouteMetadata holds parsed information about a single route.
type RouteMetadata struct {
	// Method is the HTTP method (GET, POST, PUT, PATCH, DELETE, etc.).
	Method string

	// Path is the Gin route path (e.g., "/api/users/:id").
	Path string

	// OpenAPIPath is the path converted to OpenAPI format (e.g., "/api/users/{id}").
	OpenAPIPath string

	// HandlerName is the fully qualified handler function name.
	HandlerName string

	// PathParams lists path parameter names extracted from the route.
	PathParams []string

	// Tags are auto-detected operation tags (from route groups).
	Tags []string
}

// introspect reads all routes from the Gin router and builds RouteMetadata entries.
func (gd *GinDocs) introspect() []RouteMetadata {
	routes := gd.router.Routes()
	result := make([]RouteMetadata, 0, len(routes))

	for _, r := range routes {
		// Skip documentation routes themselves.
		if gd.isDocRoute(r.Path) {
			continue
		}

		// Skip excluded routes.
		if gd.isExcluded(r.Path) {
			continue
		}

		meta := RouteMetadata{
			Method:      r.Method,
			Path:        r.Path,
			OpenAPIPath: ginPathToOpenAPI(r.Path),
			HandlerName: r.Handler,
			PathParams:  extractPathParams(r.Path),
			Tags:        inferTags(r.Path),
		}

		result = append(result, meta)
	}

	return result
}

// ginPathToOpenAPI converts Gin's :param and *param syntax to OpenAPI {param}.
func ginPathToOpenAPI(ginPath string) string {
	segments := strings.Split(ginPath, "/")
	for i, seg := range segments {
		if strings.HasPrefix(seg, ":") {
			segments[i] = "{" + seg[1:] + "}"
		} else if strings.HasPrefix(seg, "*") {
			segments[i] = "{" + seg[1:] + "}"
		}
	}
	return strings.Join(segments, "/")
}

// extractPathParams returns the names of all path parameters in a Gin route.
func extractPathParams(ginPath string) []string {
	var params []string
	segments := strings.Split(ginPath, "/")
	for _, seg := range segments {
		if strings.HasPrefix(seg, ":") {
			params = append(params, seg[1:])
		} else if strings.HasPrefix(seg, "*") {
			params = append(params, seg[1:])
		}
	}
	return params
}

// inferTags auto-detects tags from the route path.
// Uses the first meaningful path segment after common API prefixes.
func inferTags(routePath string) []string {
	// Clean the path.
	p := strings.TrimPrefix(routePath, "/")
	segments := strings.Split(p, "/")

	// Skip common API prefixes.
	startIdx := 0
	for i, seg := range segments {
		lower := strings.ToLower(seg)
		if lower == "api" || lower == "v1" || lower == "v2" || lower == "v3" {
			startIdx = i + 1
			continue
		}
		break
	}

	if startIdx >= len(segments) {
		return nil
	}

	tag := segments[startIdx]
	// Don't use path params as tags.
	if strings.HasPrefix(tag, ":") || strings.HasPrefix(tag, "*") {
		return nil
	}

	// Capitalize first letter for nicer display.
	tag = capitalizeTag(tag)

	return []string{tag}
}

// capitalizeTag capitalizes the first letter and handles common transformations.
func capitalizeTag(s string) string {
	if s == "" {
		return s
	}
	// Convert kebab-case or snake_case to Title Case.
	s = strings.ReplaceAll(s, "-", " ")
	s = strings.ReplaceAll(s, "_", " ")

	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}

	return strings.Join(words, " ")
}

// isDocRoute checks if a path belongs to the documentation routes.
func (gd *GinDocs) isDocRoute(routePath string) bool {
	prefix := gd.config.Prefix
	return routePath == prefix ||
		routePath == prefix+"/" ||
		strings.HasPrefix(routePath, prefix+"/")
}

// isExcluded checks if a route should be excluded from documentation.
func (gd *GinDocs) isExcluded(routePath string) bool {
	// Check prefix exclusions.
	for _, prefix := range gd.config.ExcludePrefixes {
		if strings.HasPrefix(routePath, prefix) {
			return true
		}
	}

	// Check glob pattern exclusions.
	for _, pattern := range gd.config.ExcludeRoutes {
		if matched, _ := path.Match(pattern, routePath); matched {
			return true
		}
	}

	return false
}
