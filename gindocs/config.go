package gindocs

// UIType represents the documentation UI to serve.
type UIType int

const (
	// UISwagger serves the Swagger UI (default).
	UISwagger UIType = iota
	// UIScalar serves the Scalar UI.
	UIScalar
)

// AuthType represents the authentication method for "Try It" functionality.
type AuthType int

const (
	// AuthNone disables authentication in the UI.
	AuthNone AuthType = iota
	// AuthBearer enables Bearer token authentication.
	AuthBearer
	// AuthAPIKey enables API key authentication.
	AuthAPIKey
	// AuthBasic enables Basic authentication.
	AuthBasic
)

// Config holds all configuration for Gin Docs.
type Config struct {
	// Prefix is the URL prefix for docs endpoints (default: "/docs").
	Prefix string

	// Title is the API title shown in the docs (default: auto-detect from module name).
	Title string

	// Description is the API description.
	Description string

	// Version is the API version (default: "1.0.0").
	Version string

	// UI selects the documentation UI: UISwagger (default) or UIScalar.
	UI UIType

	// DevMode re-introspects routes on every request when true.
	// Defaults to auto-detection from GIN_MODE.
	DevMode bool

	// ReadOnly disables "Try It" functionality when true.
	ReadOnly bool

	// Auth configures authentication for "Try It" requests.
	Auth AuthConfig

	// Servers lists API server URLs for "Try It" requests.
	Servers []ServerInfo

	// Contact holds API contact information.
	Contact ContactInfo

	// License holds API license information.
	License LicenseInfo

	// Logo is a URL to a custom logo displayed in the UI.
	Logo string

	// ExcludeRoutes is a list of glob patterns for routes to exclude from docs.
	ExcludeRoutes []string

	// ExcludePrefixes is a list of path prefixes for routes to exclude from docs.
	ExcludePrefixes []string

	// Models is a list of GORM model instances to register as schemas.
	Models []interface{}

	// CustomSections adds extra documentation sections rendered as markdown.
	CustomSections []Section

	// CustomCSS is custom CSS injected into the documentation UI.
	CustomCSS string
}

// AuthConfig configures authentication for the "Try It" feature.
type AuthConfig struct {
	// Type is the authentication method.
	Type AuthType

	// Name is the header or query parameter name (for API key auth).
	Name string

	// In specifies where the API key is sent: "header" or "query" (for API key auth).
	In string

	// Scheme is the HTTP auth scheme (default: "bearer" for Bearer auth).
	Scheme string

	// BearerFormat describes the bearer token format (e.g., "JWT").
	BearerFormat string
}

// ServerInfo describes an API server.
type ServerInfo struct {
	// URL is the server URL.
	URL string

	// Description describes this server.
	Description string
}

// ContactInfo holds API contact information.
type ContactInfo struct {
	// Name is the contact name.
	Name string

	// URL is the contact URL.
	URL string

	// Email is the contact email.
	Email string
}

// LicenseInfo holds API license information.
type LicenseInfo struct {
	// Name is the license name (e.g., "MIT").
	Name string

	// URL is the license URL.
	URL string
}

// Section represents a custom documentation section.
type Section struct {
	// Title is the section heading.
	Title string

	// Content is the section body in markdown.
	Content string
}

// defaultConfig returns a Config with sensible defaults applied.
func defaultConfig() Config {
	return Config{
		Prefix:  "/docs",
		Version: "1.0.0",
		UI:      UIScalar,
	}
}

// mergeConfig applies user-provided config values over defaults.
func mergeConfig(configs ...Config) Config {
	cfg := defaultConfig()
	if len(configs) == 0 {
		return cfg
	}

	c := configs[0]

	if c.Prefix != "" {
		cfg.Prefix = c.Prefix
	}
	if c.Title != "" {
		cfg.Title = c.Title
	}
	if c.Description != "" {
		cfg.Description = c.Description
	}
	if c.Version != "" {
		cfg.Version = c.Version
	}
	// Always take the user's UI choice — UISwagger is 0, UIScalar is 1.
	cfg.UI = c.UI
	cfg.DevMode = c.DevMode
	cfg.ReadOnly = c.ReadOnly
	if c.Auth.Type != AuthNone {
		cfg.Auth = c.Auth
	}
	if len(c.Servers) > 0 {
		cfg.Servers = c.Servers
	}
	if c.Contact != (ContactInfo{}) {
		cfg.Contact = c.Contact
	}
	if c.License != (LicenseInfo{}) {
		cfg.License = c.License
	}
	if c.Logo != "" {
		cfg.Logo = c.Logo
	}
	if len(c.ExcludeRoutes) > 0 {
		cfg.ExcludeRoutes = c.ExcludeRoutes
	}
	if len(c.ExcludePrefixes) > 0 {
		cfg.ExcludePrefixes = c.ExcludePrefixes
	}
	if len(c.Models) > 0 {
		cfg.Models = c.Models
	}
	if len(c.CustomSections) > 0 {
		cfg.CustomSections = c.CustomSections
	}
	if c.CustomCSS != "" {
		cfg.CustomCSS = c.CustomCSS
	}

	return cfg
}
