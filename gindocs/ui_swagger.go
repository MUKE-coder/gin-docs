package gindocs

import (
	"fmt"
	"html/template"
	"strings"
)

// swaggerUIVersion is the Swagger UI version loaded from CDN.
const swaggerUIVersion = "5.18.2"

// renderSwaggerHTML generates the full Swagger UI HTML page.
func renderSwaggerHTML(title, specURL string, cfg Config) string {
	readOnlyStr := "false"
	if cfg.ReadOnly {
		readOnlyStr = "true"
	}

	logoHTML := ""
	if cfg.Logo != "" {
		logoHTML = fmt.Sprintf(`<img src="%s" alt="Logo" style="max-height:40px;margin-right:12px;">`, template.HTMLEscapeString(cfg.Logo))
	}

	customCSS := ""
	if cfg.CustomCSS != "" {
		customCSS = fmt.Sprintf("<style>%s</style>", cfg.CustomCSS)
	}

	// Build auth config for Swagger UI.
	authConfigJS := ""
	if cfg.Auth.Type != AuthNone {
		switch cfg.Auth.Type {
		case AuthBearer:
			authConfigJS = `
        requestInterceptor: (req) => {
          const token = window.ui?.getState()?.getIn(["auth", "authorized", "bearerAuth", "value"]);
          if (token) { req.headers["Authorization"] = "Bearer " + token; }
          return req;
        },`
		case AuthAPIKey:
			name := cfg.Auth.Name
			if name == "" {
				name = "X-API-Key"
			}
			authConfigJS = fmt.Sprintf(`
        requestInterceptor: (req) => {
          const key = window.ui?.getState()?.getIn(["auth", "authorized", "apiKeyAuth", "value"]);
          if (key) { req.headers["%s"] = key; }
          return req;
        },`, template.JSEscapeString(name))
		}
	}

	// Build the custom sections markdown if any.
	var customSectionsHTML strings.Builder
	if len(cfg.CustomSections) > 0 {
		customSectionsHTML.WriteString(`<div id="custom-sections" style="padding:20px 40px;max-width:900px;">`)
		for _, section := range cfg.CustomSections {
			customSectionsHTML.WriteString(fmt.Sprintf(
				`<div style="margin-bottom:2rem;"><h2 style="color:#333;border-bottom:2px solid #49cc90;padding-bottom:8px;">%s</h2><div style="white-space:pre-wrap;line-height:1.6;color:#3b4151;">%s</div></div>`,
				template.HTMLEscapeString(section.Title),
				template.HTMLEscapeString(section.Content),
			))
		}
		customSectionsHTML.WriteString(`</div>`)
	}

	switcherLink := `<a href="?ui=scalar" style="color:#fff;background:#6c63ff;padding:6px 14px;border-radius:4px;text-decoration:none;font-size:13px;font-weight:600;">Switch to Scalar</a>`

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@%s/swagger-ui.css">
    <style>
        html { box-sizing: border-box; overflow-y: scroll; }
        *, *:before, *:after { box-sizing: inherit; }
        body { margin: 0; background: #fafafa; }
        .topbar-wrapper { display: flex; align-items: center; }
        .topbar-wrapper .link { display: flex; align-items: center; }
        #ui-switcher {
            position: fixed; top: 12px; right: 20px; z-index: 10000;
            display: flex; align-items: center; gap: 8px;
        }
        .swagger-ui .topbar { background-color: #2d3748; padding: 8px 0; }
        .swagger-ui .topbar .download-url-wrapper { display: none; }
    </style>
    %s
</head>
<body>
    <div id="ui-switcher">%s %s</div>
    <div id="swagger-ui"></div>
    %s

    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@%s/swagger-ui-bundle.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@%s/swagger-ui-standalone-preset.js"></script>
    <script>
    window.onload = function() {
        window.ui = SwaggerUIBundle({
            url: "%s",
            dom_id: '#swagger-ui',
            deepLinking: true,
            presets: [
                SwaggerUIBundle.presets.apis,
                SwaggerUIStandalonePreset
            ],
            plugins: [
                SwaggerUIBundle.plugins.DownloadUrl
            ],
            layout: "StandaloneLayout",
            tryItOutEnabled: !%s,
            %s
        });
    };
    </script>
</body>
</html>`,
		template.HTMLEscapeString(title),
		swaggerUIVersion,
		customCSS,
		logoHTML,
		switcherLink,
		customSectionsHTML.String(),
		swaggerUIVersion,
		swaggerUIVersion,
		template.JSEscapeString(specURL),
		readOnlyStr,
		authConfigJS,
	)
}
