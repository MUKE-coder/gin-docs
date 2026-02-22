package gindocs

import (
	"fmt"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GinDocs is the core engine that orchestrates route introspection,
// schema generation, and OpenAPI spec assembly.
type GinDocs struct {
	router *gin.Engine
	db     *gorm.DB
	config Config

	// spec holds the generated OpenAPI specification.
	spec *OpenAPISpec
	// specMu guards concurrent access to the spec.
	specMu sync.RWMutex

	// registry manages schema deduplication and $ref generation.
	registry *TypeRegistry

	// routes holds discovered route metadata after introspection.
	routes []RouteMetadata

	// routeOverrides holds per-route documentation overrides.
	routeOverrides map[string]*RouteOverride

	// groupOverrides holds group-level documentation overrides.
	groupOverrides map[string]*GroupOverride

	// built tracks whether the spec has been generated.
	built bool
}

// newGinDocs creates a new GinDocs engine with the given configuration.
func newGinDocs(router *gin.Engine, db *gorm.DB, config Config) *GinDocs {
	gd := &GinDocs{
		router:   router,
		db:       db,
		config:   config,
		registry: newTypeRegistry(),
	}
	return gd
}

// getSpec returns the current OpenAPI spec, building it if necessary.
func (gd *GinDocs) getSpec() *OpenAPISpec {
	if gd.config.DevMode {
		gd.buildSpec()
		return gd.spec
	}

	gd.specMu.RLock()
	if gd.built {
		defer gd.specMu.RUnlock()
		return gd.spec
	}
	gd.specMu.RUnlock()

	gd.buildSpec()
	return gd.spec
}

// buildSpec generates the OpenAPI specification from the router and models.
func (gd *GinDocs) buildSpec() {
	gd.specMu.Lock()
	defer gd.specMu.Unlock()

	// Reset registry for fresh build.
	gd.registry = newTypeRegistry()

	gd.spec = gd.assembleSpec()
	gd.built = true
}

// generateSummary creates a human-readable summary from method and path.
func generateSummary(method, path string) string {
	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	resource := ""
	parentResource := ""
	for _, seg := range segments {
		if seg == "api" || seg == "v1" || seg == "v2" || seg == "v3" {
			continue
		}
		if strings.HasPrefix(seg, ":") || strings.HasPrefix(seg, "*") {
			continue
		}
		parentResource = resource
		resource = seg
	}

	if resource == "" {
		return method + " " + path
	}

	hasParam := false
	for _, seg := range segments {
		if strings.HasPrefix(seg, ":") || strings.HasPrefix(seg, "*") {
			hasParam = true
			break
		}
	}

	lastSeg := segments[len(segments)-1]
	isDetail := strings.HasPrefix(lastSeg, ":") || strings.HasPrefix(lastSeg, "*")

	singular := singularize(resource)
	switch method {
	case "GET":
		if isDetail {
			return fmt.Sprintf("Get a %s by ID", singular)
		}
		if parentResource != "" && hasParam {
			return fmt.Sprintf("List %s for a %s", resource, singularize(parentResource))
		}
		return fmt.Sprintf("List all %s", resource)
	case "POST":
		return fmt.Sprintf("Create a new %s", singular)
	case "PUT":
		return fmt.Sprintf("Update a %s by ID", singular)
	case "PATCH":
		return fmt.Sprintf("Partially update a %s by ID", singular)
	case "DELETE":
		return fmt.Sprintf("Delete a %s by ID", singular)
	default:
		return method + " " + path
	}
}

// singularize does a basic plurals-to-singular conversion.
func singularize(s string) string {
	if strings.HasSuffix(s, "ies") {
		return s[:len(s)-3] + "y"
	}
	if strings.HasSuffix(s, "ses") || strings.HasSuffix(s, "xes") || strings.HasSuffix(s, "zes") {
		return s[:len(s)-2]
	}
	if strings.HasSuffix(s, "s") && !strings.HasSuffix(s, "ss") {
		return s[:len(s)-1]
	}
	return s
}

// generateOperationID creates a unique operation ID from method and path.
func generateOperationID(method, routePath string) string {
	id := strings.ToLower(method)
	segments := strings.Split(strings.TrimPrefix(routePath, "/"), "/")
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		if strings.HasPrefix(seg, ":") {
			id += "By" + capitalize(seg[1:])
		} else if strings.HasPrefix(seg, "*") {
			id += capitalize(seg[1:])
		} else {
			id += capitalize(seg)
		}
	}
	return id
}

// capitalize capitalizes the first letter of a string.
func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
