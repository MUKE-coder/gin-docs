package gindocs

import (
	"fmt"
	"html/template"
	"strings"
)

// renderScalarHTML generates the full Scalar UI HTML page.
func renderScalarHTML(title, specURL string, cfg Config) string {
	customCSS := ""
	if cfg.CustomCSS != "" {
		customCSS = fmt.Sprintf("<style>%s</style>", cfg.CustomCSS)
	}

	// Build auth configuration for Scalar.
	authJSON := ""
	if cfg.Auth.Type != AuthNone {
		switch cfg.Auth.Type {
		case AuthBearer:
			authJSON = `authentication: { preferredSecurityScheme: "bearerAuth" },`
		case AuthAPIKey:
			authJSON = `authentication: { preferredSecurityScheme: "apiKeyAuth" },`
		case AuthBasic:
			authJSON = `authentication: { preferredSecurityScheme: "basicAuth" },`
		}
	}

	hideModels := ""
	if cfg.ReadOnly {
		hideModels = `hiddenClients: true,`
	}

	// Custom sections rendered below the API reference.
	var customSectionsHTML strings.Builder
	if len(cfg.CustomSections) > 0 {
		customSectionsHTML.WriteString(`<div style="padding:24px 32px;max-width:900px;margin:0 auto;">`)
		for _, section := range cfg.CustomSections {
			customSectionsHTML.WriteString(fmt.Sprintf(
				`<div style="margin-bottom:2rem;"><h2 style="font-size:1.4rem;font-weight:600;margin-bottom:0.5rem;color:#1a1a2e;">%s</h2><div style="white-space:pre-wrap;line-height:1.7;color:#4a4a6a;">%s</div></div>`,
				template.HTMLEscapeString(section.Title),
				template.HTMLEscapeString(section.Content),
			))
		}
		customSectionsHTML.WriteString(`</div>`)
	}

	switcherLink := fmt.Sprintf(`<a href="?ui=swagger" style="color:#fff;background:#49cc90;padding:6px 14px;border-radius:4px;text-decoration:none;font-size:13px;font-weight:600;">Switch to Swagger</a>`)

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>%s</title>
    <style>
        body { margin: 0; }
        #ui-switcher {
            position: fixed; top: 12px; right: 20px; z-index: 10000;
            display: flex; align-items: center; gap: 8px;
        }
    </style>
    %s
</head>
<body>
    <div id="ui-switcher">%s</div>

    <script id="api-reference" data-url="%s"></script>
    <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
    <script>
        document.getElementById('api-reference').dataset.configuration = JSON.stringify({
            theme: '%s',
            %s
            %s
        });
    </script>

    %s
</body>
</html>`,
		template.HTMLEscapeString(title),
		customCSS,
		switcherLink,
		template.HTMLEscapeString(specURL),
		template.JSEscapeString(cfg.ScalarTheme),
		authJSON,
		hideModels,
		customSectionsHTML.String(),
	)
}
