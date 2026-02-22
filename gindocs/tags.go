package gindocs

import (
	"strconv"
	"strings"
)

// TagInfo holds parsed information from struct field tags.
type TagInfo struct {
	// JSON tag
	JSONName  string
	OmitEmpty bool
	JSONSkip  bool

	// Binding/validate tag
	Required    bool
	MinLength   *int
	MaxLength   *int
	Minimum     *float64
	Maximum     *float64
	Enum        []string
	Format      string // e.g., "email", "uri", "uuid"
	Pattern     string
	BindingSkip bool

	// GORM tag
	PrimaryKey     bool
	AutoCreateTime bool
	AutoUpdateTime bool
	GORMSize       *int
	GORMDefault    *string
	UniqueIndex    bool
	GORMSkip       bool
	GORMType       string

	// Docs tag
	Description string
	Example     string
	Deprecated  bool
	Hidden      bool
	DocsFormat  string
	DocsEnum    []string
}

// parseJSONTag parses a json struct tag value.
func parseJSONTag(tag string) (name string, omitEmpty bool, skip bool) {
	if tag == "" {
		return "", false, false
	}
	if tag == "-" {
		return "", false, true
	}

	parts := strings.Split(tag, ",")
	name = parts[0]

	for _, opt := range parts[1:] {
		if opt == "omitempty" {
			omitEmpty = true
		}
	}

	return name, omitEmpty, false
}

// parseBindingTag parses a binding or validate struct tag value.
func parseBindingTag(tag string) TagInfo {
	var info TagInfo
	if tag == "" || tag == "-" {
		if tag == "-" {
			info.BindingSkip = true
		}
		return info
	}

	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		switch {
		case part == "required":
			info.Required = true
		case part == "email":
			info.Format = "email"
		case part == "url" || part == "uri" || part == "http_url":
			info.Format = "uri"
		case part == "uuid" || part == "uuid3" || part == "uuid4" || part == "uuid5":
			info.Format = "uuid"
		case part == "ipv4":
			info.Format = "ipv4"
		case part == "ipv6":
			info.Format = "ipv6"
		case part == "ip":
			info.Format = "ipv4"
		case part == "datetime":
			info.Format = "date-time"
		case strings.HasPrefix(part, "oneof="):
			values := strings.TrimPrefix(part, "oneof=")
			info.Enum = strings.Fields(values)
		case strings.HasPrefix(part, "min="):
			if v, err := strconv.Atoi(strings.TrimPrefix(part, "min=")); err == nil {
				info.MinLength = intPtr(v)
				f := float64(v)
				info.Minimum = &f
			}
		case strings.HasPrefix(part, "max="):
			if v, err := strconv.Atoi(strings.TrimPrefix(part, "max=")); err == nil {
				info.MaxLength = intPtr(v)
				f := float64(v)
				info.Maximum = &f
			}
		case strings.HasPrefix(part, "gte="):
			if v, err := strconv.ParseFloat(strings.TrimPrefix(part, "gte="), 64); err == nil {
				info.Minimum = &v
			}
		case strings.HasPrefix(part, "gt="):
			if v, err := strconv.ParseFloat(strings.TrimPrefix(part, "gt="), 64); err == nil {
				info.Minimum = &v
			}
		case strings.HasPrefix(part, "lte="):
			if v, err := strconv.ParseFloat(strings.TrimPrefix(part, "lte="), 64); err == nil {
				info.Maximum = &v
			}
		case strings.HasPrefix(part, "lt="):
			if v, err := strconv.ParseFloat(strings.TrimPrefix(part, "lt="), 64); err == nil {
				info.Maximum = &v
			}
		case strings.HasPrefix(part, "len="):
			if v, err := strconv.Atoi(strings.TrimPrefix(part, "len=")); err == nil {
				info.MinLength = intPtr(v)
				info.MaxLength = intPtr(v)
			}
		}
	}

	return info
}

// parseGORMTag parses a gorm struct tag value.
func parseGORMTag(tag string) TagInfo {
	var info TagInfo
	if tag == "" {
		return info
	}
	if tag == "-" || tag == "-:all" {
		info.GORMSkip = true
		return info
	}

	// GORM tags use semicolons as separators.
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		lower := strings.ToLower(part)

		switch {
		case lower == "primarykey" || lower == "primary_key":
			info.PrimaryKey = true
		case lower == "autocreatetime":
			info.AutoCreateTime = true
		case lower == "autoupdatetime":
			info.AutoUpdateTime = true
		case lower == "uniqueindex" || strings.HasPrefix(lower, "uniqueindex:"):
			info.UniqueIndex = true
		case strings.HasPrefix(lower, "size:"):
			if v, err := strconv.Atoi(strings.TrimPrefix(lower, "size:")); err == nil {
				info.GORMSize = intPtr(v)
			}
		case strings.HasPrefix(lower, "default:"):
			val := strings.TrimPrefix(part, "default:")
			val = strings.TrimPrefix(val, "Default:")
			// Remove surrounding quotes.
			val = strings.Trim(val, "'\"")
			info.GORMDefault = &val
		case strings.HasPrefix(lower, "type:"):
			info.GORMType = strings.TrimPrefix(part, "type:")
		}
	}

	return info
}

// parseDocsTag parses a docs struct tag value.
func parseDocsTag(tag string) TagInfo {
	var info TagInfo
	if tag == "" {
		return info
	}

	parts := strings.Split(tag, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)

		switch {
		case part == "deprecated":
			info.Deprecated = true
		case part == "hidden":
			info.Hidden = true
		case strings.HasPrefix(part, "description:"):
			info.Description = strings.TrimPrefix(part, "description:")
		case strings.HasPrefix(part, "example:"):
			info.Example = strings.TrimPrefix(part, "example:")
		case strings.HasPrefix(part, "format:"):
			info.DocsFormat = strings.TrimPrefix(part, "format:")
		case strings.HasPrefix(part, "enum:"):
			enumStr := strings.TrimPrefix(part, "enum:")
			info.DocsEnum = strings.Split(enumStr, "|")
		}
	}

	return info
}

// mergeTags merges parsed tag info from all tag sources into a single TagInfo.
func mergeTags(jsonTag, bindingTag, gormTag, docsTag string) TagInfo {
	name, omitEmpty, jsonSkip := parseJSONTag(jsonTag)
	binding := parseBindingTag(bindingTag)
	gorm := parseGORMTag(gormTag)
	docs := parseDocsTag(docsTag)

	info := TagInfo{
		// JSON
		JSONName:  name,
		OmitEmpty: omitEmpty,
		JSONSkip:  jsonSkip,

		// Binding
		Required:    binding.Required,
		MinLength:   binding.MinLength,
		MaxLength:   binding.MaxLength,
		Minimum:     binding.Minimum,
		Maximum:     binding.Maximum,
		Enum:        binding.Enum,
		Format:      binding.Format,
		Pattern:     binding.Pattern,
		BindingSkip: binding.BindingSkip,

		// GORM
		PrimaryKey:     gorm.PrimaryKey,
		AutoCreateTime: gorm.AutoCreateTime,
		AutoUpdateTime: gorm.AutoUpdateTime,
		GORMSize:       gorm.GORMSize,
		GORMDefault:    gorm.GORMDefault,
		UniqueIndex:    gorm.UniqueIndex,
		GORMSkip:       gorm.GORMSkip,
		GORMType:       gorm.GORMType,

		// Docs
		Description: docs.Description,
		Example:     docs.Example,
		Deprecated:  docs.Deprecated,
		Hidden:      docs.Hidden,
		DocsFormat:  docs.DocsFormat,
		DocsEnum:    docs.DocsEnum,
	}

	// Docs format overrides binding format.
	if info.DocsFormat != "" {
		info.Format = info.DocsFormat
	}

	// Docs enum overrides binding enum.
	if len(info.DocsEnum) > 0 {
		info.Enum = info.DocsEnum
	}

	return info
}

func intPtr(v int) *int {
	return &v
}
