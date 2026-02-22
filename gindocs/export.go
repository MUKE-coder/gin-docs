package gindocs

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PostmanCollection represents a Postman v2.1 collection.
type PostmanCollection struct {
	Info PostmanInfo   `json:"info"`
	Item []PostmanItem `json:"item"`
}

// PostmanInfo holds collection metadata.
type PostmanInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Schema      string `json:"schema"`
}

// PostmanItem represents a folder or request in a Postman collection.
type PostmanItem struct {
	Name    string            `json:"name"`
	Item    []PostmanItem     `json:"item,omitempty"`
	Request *PostmanRequest   `json:"request,omitempty"`
}

// PostmanRequest represents a Postman request.
type PostmanRequest struct {
	Method      string          `json:"method"`
	Header      []PostmanHeader `json:"header,omitempty"`
	Body        *PostmanBody    `json:"body,omitempty"`
	URL         PostmanURL      `json:"url"`
	Description string          `json:"description,omitempty"`
}

// PostmanHeader represents a request header.
type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

// PostmanBody represents a request body.
type PostmanBody struct {
	Mode string `json:"mode"`
	Raw  string `json:"raw,omitempty"`
	Options *PostmanBodyOptions `json:"options,omitempty"`
}

// PostmanBodyOptions holds body format options.
type PostmanBodyOptions struct {
	Raw PostmanRawOptions `json:"raw"`
}

// PostmanRawOptions holds raw body language setting.
type PostmanRawOptions struct {
	Language string `json:"language"`
}

// PostmanURL represents a Postman URL.
type PostmanURL struct {
	Raw      string   `json:"raw"`
	Protocol string   `json:"protocol,omitempty"`
	Host     []string `json:"host,omitempty"`
	Path     []string `json:"path,omitempty"`
}

// generatePostmanCollection creates a Postman v2.1 collection from the spec.
func generatePostmanCollection(spec *OpenAPISpec) *PostmanCollection {
	collection := &PostmanCollection{
		Info: PostmanInfo{
			Name:        spec.Info.Title,
			Description: spec.Info.Description,
			Schema:      "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
	}

	// Determine base URL.
	baseURL := "http://localhost:8080"
	if len(spec.Servers) > 0 {
		baseURL = spec.Servers[0].URL
	}

	// Group requests by tag.
	tagFolders := make(map[string]*PostmanItem)
	var ungrouped []PostmanItem

	for path, pathItem := range spec.Paths {
		operations := []struct {
			method string
			op     *OperationObject
		}{
			{"GET", pathItem.Get},
			{"POST", pathItem.Post},
			{"PUT", pathItem.Put},
			{"PATCH", pathItem.Patch},
			{"DELETE", pathItem.Delete},
			{"HEAD", pathItem.Head},
			{"OPTIONS", pathItem.Options},
		}

		for _, entry := range operations {
			if entry.op == nil {
				continue
			}

			item := createPostmanItem(entry.method, path, baseURL, entry.op)

			if len(entry.op.Tags) > 0 {
				tag := entry.op.Tags[0]
				folder, ok := tagFolders[tag]
				if !ok {
					folder = &PostmanItem{Name: tag}
					tagFolders[tag] = folder
				}
				folder.Item = append(folder.Item, item)
			} else {
				ungrouped = append(ungrouped, item)
			}
		}
	}

	// Add folders to collection.
	for _, folder := range tagFolders {
		collection.Item = append(collection.Item, *folder)
	}
	collection.Item = append(collection.Item, ungrouped...)

	return collection
}

// createPostmanItem creates a Postman request item from an operation.
func createPostmanItem(method, path, baseURL string, op *OperationObject) PostmanItem {
	// Convert OpenAPI path params to Postman format.
	postmanPath := path
	postmanPath = strings.ReplaceAll(postmanPath, "{", ":")
	postmanPath = strings.ReplaceAll(postmanPath, "}", "")

	name := op.Summary
	if name == "" {
		name = method + " " + path
	}

	rawURL := baseURL + postmanPath
	pathSegments := strings.Split(strings.TrimPrefix(postmanPath, "/"), "/")

	item := PostmanItem{
		Name: name,
		Request: &PostmanRequest{
			Method:      method,
			Description: op.Description,
			URL: PostmanURL{
				Raw:  rawURL,
				Path: pathSegments,
			},
			Header: []PostmanHeader{
				{Key: "Content-Type", Value: "application/json", Type: "text"},
				{Key: "Accept", Value: "application/json", Type: "text"},
			},
		},
	}

	// Add request body for appropriate methods.
	if op.RequestBody != nil && (method == "POST" || method == "PUT" || method == "PATCH") {
		item.Request.Body = &PostmanBody{
			Mode: "raw",
			Raw:  "{}",
			Options: &PostmanBodyOptions{
				Raw: PostmanRawOptions{Language: "json"},
			},
		}
	}

	return item
}

// InsomniaExport represents an Insomnia v4 export.
type InsomniaExport struct {
	Type      string           `json:"_type"`
	Format    int              `json:"__export_format"`
	Source    string           `json:"__export_source"`
	Resources []InsomniaResource `json:"resources"`
}

// InsomniaResource represents a resource in an Insomnia export.
type InsomniaResource struct {
	ID          string      `json:"_id"`
	Type        string      `json:"_type"`
	ParentID    string      `json:"parentId,omitempty"`
	Name        string      `json:"name"`
	Description string      `json:"description,omitempty"`
	URL         string      `json:"url,omitempty"`
	Method      string      `json:"method,omitempty"`
	Body        interface{} `json:"body,omitempty"`
	Headers     []InsomniaHeader `json:"headers,omitempty"`
}

// InsomniaHeader represents a header in an Insomnia request.
type InsomniaHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// generateInsomniaExport creates an Insomnia v4 export from the spec.
func generateInsomniaExport(spec *OpenAPISpec) *InsomniaExport {
	export := &InsomniaExport{
		Type:   "export",
		Format: 4,
		Source: "gindocs",
	}

	baseURL := "http://localhost:8080"
	if len(spec.Servers) > 0 {
		baseURL = spec.Servers[0].URL
	}

	// Add workspace.
	workspaceID := "wrk_gindocs"
	export.Resources = append(export.Resources, InsomniaResource{
		ID:          workspaceID,
		Type:        "workspace",
		Name:        spec.Info.Title,
		Description: spec.Info.Description,
	})

	// Add folders for each tag.
	tagFolderIDs := make(map[string]string)
	for _, tag := range spec.Tags {
		folderID := fmt.Sprintf("fld_%s", strings.ToLower(tag.Name))
		tagFolderIDs[tag.Name] = folderID
		export.Resources = append(export.Resources, InsomniaResource{
			ID:       folderID,
			Type:     "request_group",
			ParentID: workspaceID,
			Name:     tag.Name,
		})
	}

	// Add requests.
	requestIdx := 0
	for path, pathItem := range spec.Paths {
		operations := []struct {
			method string
			op     *OperationObject
		}{
			{"GET", pathItem.Get},
			{"POST", pathItem.Post},
			{"PUT", pathItem.Put},
			{"PATCH", pathItem.Patch},
			{"DELETE", pathItem.Delete},
		}

		for _, entry := range operations {
			if entry.op == nil {
				continue
			}

			requestIdx++
			reqID := fmt.Sprintf("req_%d", requestIdx)

			parentID := workspaceID
			if len(entry.op.Tags) > 0 {
				if fid, ok := tagFolderIDs[entry.op.Tags[0]]; ok {
					parentID = fid
				}
			}

			// Convert OpenAPI path params to Insomnia format.
			insomniaPath := path
			insomniaPath = strings.ReplaceAll(insomniaPath, "{", "{{ ")
			insomniaPath = strings.ReplaceAll(insomniaPath, "}", " }}")

			name := entry.op.Summary
			if name == "" {
				name = entry.method + " " + path
			}

			resource := InsomniaResource{
				ID:       reqID,
				Type:     "request",
				ParentID: parentID,
				Name:     name,
				URL:      baseURL + insomniaPath,
				Method:   entry.method,
				Headers: []InsomniaHeader{
					{Name: "Content-Type", Value: "application/json"},
					{Name: "Accept", Value: "application/json"},
				},
			}

			if entry.op.RequestBody != nil {
				resource.Body = map[string]interface{}{
					"mimeType": "application/json",
					"text":     "{}",
				}
			}

			export.Resources = append(export.Resources, resource)
		}
	}

	return export
}

// specToYAML converts an OpenAPI spec to a basic YAML representation.
// Uses a simple JSON-to-YAML converter to avoid external dependencies.
func specToYAML(spec *OpenAPISpec) ([]byte, error) {
	data, err := json.Marshal(spec)
	if err != nil {
		return nil, err
	}

	var obj interface{}
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}

	var buf strings.Builder
	writeYAML(&buf, obj, 0)
	return []byte(buf.String()), nil
}

// writeYAML writes a Go value as YAML to the builder.
func writeYAML(buf *strings.Builder, v interface{}, indent int) {
	prefix := strings.Repeat("  ", indent)

	switch val := v.(type) {
	case map[string]interface{}:
		if len(val) == 0 {
			buf.WriteString("{}\n")
			return
		}
		for key, value := range val {
			buf.WriteString(prefix)
			buf.WriteString(key)
			buf.WriteString(":")

			switch value.(type) {
			case map[string]interface{}, []interface{}:
				buf.WriteString("\n")
				writeYAML(buf, value, indent+1)
			default:
				buf.WriteString(" ")
				writeYAML(buf, value, indent+1)
			}
		}

	case []interface{}:
		if len(val) == 0 {
			buf.WriteString("[]\n")
			return
		}
		for _, item := range val {
			buf.WriteString(prefix)
			buf.WriteString("- ")
			switch item.(type) {
			case map[string]interface{}:
				// Inline first key, indent rest.
				m := item.(map[string]interface{})
				first := true
				for key, value := range m {
					if first {
						buf.WriteString(key)
						buf.WriteString(":")
						switch value.(type) {
						case map[string]interface{}, []interface{}:
							buf.WriteString("\n")
							writeYAML(buf, value, indent+2)
						default:
							buf.WriteString(" ")
							writeYAML(buf, value, indent+2)
						}
						first = false
					} else {
						buf.WriteString(prefix)
						buf.WriteString("  ")
						buf.WriteString(key)
						buf.WriteString(":")
						switch value.(type) {
						case map[string]interface{}, []interface{}:
							buf.WriteString("\n")
							writeYAML(buf, value, indent+2)
						default:
							buf.WriteString(" ")
							writeYAML(buf, value, indent+2)
						}
					}
				}
			default:
				writeYAML(buf, item, indent+1)
			}
		}

	case string:
		// Check if we need quoting.
		if needsYAMLQuoting(val) {
			buf.WriteString(fmt.Sprintf("%q", val))
		} else {
			buf.WriteString(val)
		}
		buf.WriteString("\n")

	case float64:
		if val == float64(int64(val)) {
			buf.WriteString(fmt.Sprintf("%d", int64(val)))
		} else {
			buf.WriteString(fmt.Sprintf("%g", val))
		}
		buf.WriteString("\n")

	case bool:
		if val {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		buf.WriteString("\n")

	case nil:
		buf.WriteString("null\n")

	default:
		buf.WriteString(fmt.Sprintf("%v\n", val))
	}
}

// needsYAMLQuoting checks if a string needs to be quoted in YAML.
func needsYAMLQuoting(s string) bool {
	if s == "" {
		return true
	}
	if s == "true" || s == "false" || s == "null" || s == "yes" || s == "no" {
		return true
	}
	if strings.ContainsAny(s, ":#{}[]|>&*!%@`'\"\\,\n") {
		return true
	}
	return false
}
