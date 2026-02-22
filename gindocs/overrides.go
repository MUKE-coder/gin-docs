package gindocs

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// RouteOverride holds documentation overrides for a specific route.
type RouteOverride struct {
	gd     *GinDocs
	method string
	path   string

	summary     *string
	description *string
	tags        []string
	deprecated  *bool
	security    []string

	requestBodyType reflect.Type
	responses       []responseOverride
}

type responseOverride struct {
	statusCode  int
	bodyType    reflect.Type
	description string
}

// GroupOverride holds documentation overrides for a route group.
type GroupOverride struct {
	gd      *GinDocs
	pattern string

	tags     []string
	security []string
}

// Route returns a RouteOverride builder for the specified "METHOD /path" key.
func (gd *GinDocs) Route(key string) *RouteOverride {
	parts := strings.SplitN(key, " ", 2)
	method := "GET"
	path := key
	if len(parts) == 2 {
		method = strings.ToUpper(parts[0])
		path = parts[1]
	}

	override := &RouteOverride{
		gd:     gd,
		method: method,
		path:   path,
	}

	if gd.routeOverrides == nil {
		gd.routeOverrides = make(map[string]*RouteOverride)
	}

	overrideKey := method + " " + path
	gd.routeOverrides[overrideKey] = override

	return override
}

// Summary sets the operation summary.
func (r *RouteOverride) Summary(s string) *RouteOverride {
	r.summary = &s
	return r
}

// Description sets the operation description.
func (r *RouteOverride) Description(d string) *RouteOverride {
	r.description = &d
	return r
}

// Tags sets the operation tags.
func (r *RouteOverride) Tags(tags ...string) *RouteOverride {
	r.tags = append(r.tags, tags...)
	return r
}

// Deprecated marks the operation as deprecated.
func (r *RouteOverride) Deprecated(d bool) *RouteOverride {
	r.deprecated = &d
	return r
}

// Security sets security scheme names for this route.
func (r *RouteOverride) Security(schemes ...string) *RouteOverride {
	r.security = append(r.security, schemes...)
	return r
}

// RequestBody registers the request body type for this route.
func (r *RouteOverride) RequestBody(v interface{}) *RouteOverride {
	r.requestBodyType = reflect.TypeOf(v)
	return r
}

// Response registers a response for this route.
func (r *RouteOverride) Response(statusCode int, body interface{}, description string) *RouteOverride {
	var bodyType reflect.Type
	if body != nil {
		bodyType = reflect.TypeOf(body)
	}
	r.responses = append(r.responses, responseOverride{
		statusCode:  statusCode,
		bodyType:    bodyType,
		description: description,
	})
	return r
}

// Group returns a GroupOverride builder for routes matching the given pattern.
func (gd *GinDocs) Group(pattern string) *GroupOverride {
	override := &GroupOverride{
		gd:      gd,
		pattern: pattern,
	}

	if gd.groupOverrides == nil {
		gd.groupOverrides = make(map[string]*GroupOverride)
	}
	gd.groupOverrides[pattern] = override

	return override
}

// Tags sets the tags for all routes in the group.
func (g *GroupOverride) Tags(tags ...string) *GroupOverride {
	g.tags = append(g.tags, tags...)
	return g
}

// Security sets security scheme names for all routes in the group.
func (g *GroupOverride) Security(schemes ...string) *GroupOverride {
	g.security = append(g.security, schemes...)
	return g
}

// DocConfig holds inline documentation configuration for the Doc() middleware.
type DocConfig struct {
	// Summary is the operation summary.
	Summary string
	// Description is the operation description.
	Description string
	// Tags are the operation tags.
	Tags []string
	// RequestBody is the request body type (pass a struct instance).
	RequestBody interface{}
	// Response is the response body type (pass a struct instance).
	Response interface{}
	// ResponseCode is the primary success status code.
	ResponseCode int
	// Deprecated marks the operation as deprecated.
	Deprecated bool
}

// Doc returns a Gin middleware that registers inline documentation for a route.
func Doc(cfg DocConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("gindocs:config", cfg)
		c.Next()
	}
}

// applyRouteOverrides applies route and group overrides to an operation.
func (gd *GinDocs) applyRouteOverrides(method, path string, op *OperationObject) {
	// Apply group overrides first.
	for pattern, override := range gd.groupOverrides {
		if matchGroupPattern(path, pattern) {
			if len(override.tags) > 0 {
				op.Tags = override.tags
			}
			if len(override.security) > 0 {
				for _, scheme := range override.security {
					op.Security = append(op.Security, SecurityRequirement{
						scheme: []string{},
					})
				}
			}
		}
	}

	// Apply route-level overrides (higher priority).
	key := method + " " + path
	override, ok := gd.routeOverrides[key]
	if !ok {
		return
	}

	if override.summary != nil {
		op.Summary = *override.summary
	}
	if override.description != nil {
		op.Description = *override.description
	}
	if len(override.tags) > 0 {
		op.Tags = override.tags
	}
	if override.deprecated != nil {
		op.Deprecated = *override.deprecated
	}
	if len(override.security) > 0 {
		op.Security = nil
		for _, scheme := range override.security {
			op.Security = append(op.Security, SecurityRequirement{
				scheme: []string{},
			})
		}
	}

	// Apply request body override.
	if override.requestBodyType != nil {
		schema := typeToSchema(override.requestBodyType, gd.registry)
		op.RequestBody = &RequestBodyObject{
			Required: true,
			Content: map[string]MediaType{
				"application/json": {Schema: schema},
			},
		}
	}

	// Apply response overrides.
	if len(override.responses) > 0 {
		op.Responses = make(map[string]*Response)
		for _, resp := range override.responses {
			code := strconv.Itoa(resp.statusCode)
			response := &Response{
				Description: resp.description,
			}
			if resp.bodyType != nil {
				schema := typeToSchema(resp.bodyType, gd.registry)
				response.Content = map[string]MediaType{
					"application/json": {Schema: schema},
				}
			}
			op.Responses[code] = response
		}
	}
}

// matchGroupPattern checks if a path matches a group pattern.
func matchGroupPattern(path, pattern string) bool {
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(path, prefix)
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	return path == pattern
}
