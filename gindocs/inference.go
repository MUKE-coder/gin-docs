package gindocs

import (
	"strings"
)

// exampleValues maps common field names to realistic example values.
var exampleValues = map[string]interface{}{
	// Identity
	"id":      1,
	"uuid":    "550e8400-e29b-41d4-a716-446655440000",
	"slug":    "example-slug",

	// Personal info
	"name":       "John Doe",
	"first_name": "John",
	"last_name":  "Doe",
	"username":   "johndoe",
	"email":      "user@example.com",
	"phone":      "+1-555-123-4567",
	"age":        30,
	"avatar":     "https://example.com/avatar.jpg",
	"bio":        "A short biography about the user.",
	"password":   "********",

	// Content
	"title":       "Sample Title",
	"description": "A detailed description of the resource.",
	"content":     "Lorem ipsum dolor sit amet, consectetur adipiscing elit.",
	"body":        "The main body content goes here.",
	"summary":     "A brief summary.",
	"message":     "Hello, world!",
	"comment":     "This is a comment.",
	"text":        "Some text content.",

	// URLs and paths
	"url":          "https://example.com",
	"website":      "https://example.com",
	"image":        "https://example.com/image.jpg",
	"image_url":    "https://example.com/image.jpg",
	"avatar_url":   "https://example.com/avatar.jpg",
	"callback_url": "https://example.com/callback",

	// Dates/times
	"created_at": "2025-01-15T10:30:00Z",
	"updated_at": "2025-01-15T12:00:00Z",
	"deleted_at": "2025-01-15T14:00:00Z",
	"date":       "2025-01-15",
	"start_date": "2025-01-01",
	"end_date":   "2025-12-31",
	"timestamp":  "2025-01-15T10:30:00Z",

	// Booleans
	"is_active":   true,
	"is_admin":    false,
	"is_verified": true,
	"published":   true,
	"enabled":     true,
	"active":      true,
	"visible":     true,

	// Numbers
	"count":    42,
	"total":    100,
	"page":     1,
	"per_page": 20,
	"limit":    10,
	"offset":   0,
	"score":    85.5,
	"price":    29.99,
	"amount":   100.00,
	"quantity": 1,
	"rating":   4.5,
	"weight":   1.5,
	"order":    1,
	"sort":     "created_at",
	"priority": 1,

	// Auth
	"token":         "eyJhbGciOiJIUzI1NiIs...",
	"access_token":  "eyJhbGciOiJIUzI1NiIs...",
	"refresh_token": "dGhpcyBpcyBhIHJlZnJl...",
	"api_key":       "sk-1234567890abcdef",

	// Common fields
	"role":     "user",
	"status":   "active",
	"type":     "default",
	"category": "general",
	"tags":     []string{"tag1", "tag2"},
	"color":    "#3498db",
	"code":     "ABC123",
	"label":    "Example Label",
	"key":      "example_key",
	"value":    "example_value",

	// Address
	"address":  "123 Main St",
	"city":     "San Francisco",
	"state":    "CA",
	"country":  "US",
	"zip":      "94102",
	"zip_code": "94102",
	"lat":      37.7749,
	"lng":      -122.4194,
	"latitude": 37.7749,
	"longitude": -122.4194,
}

// inferExampleValue generates an example value for a field based on its name and type.
func inferExampleValue(fieldName, schemaType, format string) interface{} {
	lower := strings.ToLower(fieldName)

	// Check exact match first.
	if v, ok := exampleValues[lower]; ok {
		return v
	}

	// Check suffix matches.
	for key, v := range exampleValues {
		if strings.HasSuffix(lower, "_"+key) || strings.HasSuffix(lower, key) {
			return v
		}
	}

	// Fallback by type and format.
	switch schemaType {
	case "string":
		switch format {
		case "date-time":
			return "2025-01-15T10:30:00Z"
		case "date":
			return "2025-01-15"
		case "email":
			return "user@example.com"
		case "uri", "url":
			return "https://example.com"
		case "uuid":
			return "550e8400-e29b-41d4-a716-446655440000"
		case "ipv4":
			return "192.168.1.1"
		case "ipv6":
			return "::1"
		case "byte":
			return "SGVsbG8gV29ybGQ="
		case "binary":
			return "binary data"
		default:
			return "string"
		}
	case "integer":
		return 1
	case "number":
		return 1.0
	case "boolean":
		return true
	case "array":
		return []interface{}{}
	case "object":
		return map[string]interface{}{}
	}

	return nil
}

// inferQueryParams generates common query parameters based on the route and method.
func inferQueryParams(method, path string) []ParameterObject {
	var params []ParameterObject

	// Only add query params for GET list endpoints.
	if method != "GET" {
		return params
	}

	segments := strings.Split(strings.TrimPrefix(path, "/"), "/")
	lastSeg := segments[len(segments)-1]

	// If the last segment is not a param, it's a list endpoint.
	if !strings.HasPrefix(lastSeg, ":") && !strings.HasPrefix(lastSeg, "*") {
		// Check if it looks like a search endpoint.
		if strings.Contains(strings.ToLower(lastSeg), "search") {
			params = append(params, ParameterObject{
				Name:        "q",
				In:          "query",
				Description: "Search query string",
				Schema:      &SchemaObject{Type: "string"},
			})
		}
	}

	return params
}
