package gindocs

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"
)

// registerHandlers sets up all documentation-related HTTP handlers on the router.
func (gd *GinDocs) registerHandlers() {
	prefix := gd.config.Prefix

	gd.router.GET(prefix, gd.handleUI)
	gd.router.GET(prefix+"/", gd.handleUI)
	gd.router.GET(prefix+"/openapi.json", gd.handleSpecJSON)
	gd.router.GET(prefix+"/openapi.yaml", gd.handleSpecYAML)
	gd.router.GET(prefix+"/export/postman", gd.handleExportPostman)
	gd.router.GET(prefix+"/export/insomnia", gd.handleExportInsomnia)
}

// handleUI serves the documentation UI page.
func (gd *GinDocs) handleUI(c *gin.Context) {
	uiType := gd.config.UI
	if q := c.Query("ui"); q != "" {
		switch q {
		case "scalar":
			uiType = UIScalar
		case "swagger":
			uiType = UISwagger
		}
	}

	specURL := gd.config.Prefix + "/openapi.json"
	title := gd.config.Title
	if title == "" {
		title = "API Documentation"
	}

	var html string
	switch uiType {
	case UIScalar:
		html = renderScalarHTML(title, specURL, gd.config)
	default:
		html = renderSwaggerHTML(title, specURL, gd.config)
	}

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// handleSpecJSON serves the OpenAPI specification as JSON.
func (gd *GinDocs) handleSpecJSON(c *gin.Context) {
	spec := gd.getSpec()

	data, err := json.MarshalIndent(spec, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal spec"})
		return
	}

	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}

// handleSpecYAML serves the OpenAPI specification as YAML.
func (gd *GinDocs) handleSpecYAML(c *gin.Context) {
	spec := gd.getSpec()

	data, err := specToYAML(spec)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to marshal spec"})
		return
	}

	c.Header("Cache-Control", "no-cache")
	c.Data(http.StatusOK, "application/x-yaml; charset=utf-8", data)
}

// handleExportPostman exports the API as a Postman v2.1 collection.
func (gd *GinDocs) handleExportPostman(c *gin.Context) {
	spec := gd.getSpec()
	collection := generatePostmanCollection(spec)

	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate Postman collection"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\"postman_collection.json\"")
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}

// handleExportInsomnia exports the API as an Insomnia v4 export.
func (gd *GinDocs) handleExportInsomnia(c *gin.Context) {
	spec := gd.getSpec()
	export := generateInsomniaExport(spec)

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate Insomnia export"})
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\"insomnia_export.json\"")
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}
