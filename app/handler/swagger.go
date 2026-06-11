package handler

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

// SwaggerHandler serves OpenAPI spec as interactive Swagger UI.
// Endpoint: GET /docs
type SwaggerHandler struct {
	specPath string
}

func NewSwaggerHandler(specPath string) *SwaggerHandler {
	return &SwaggerHandler{specPath: specPath}
}

// ServeSwaggerUI renders a self-contained Swagger UI page that loads
// the OpenAPI spec from a JSON endpoint. No external CDN dependencies.
func (h *SwaggerHandler) ServeSwaggerUI(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, swaggerHTML)
}

// ServeSpec returns the raw OpenAPI YAML spec.
func (h *SwaggerHandler) ServeSpec(c *gin.Context) {
	data, err := os.ReadFile(h.specPath)
	if err != nil {
		// Try relative path from working dir
		absPath, _ := filepath.Abs(h.specPath)
		data, err = os.ReadFile(absPath)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "openapi.yaml not found"})
			return
		}
	}
	c.Header("Content-Type", "text/yaml; charset=utf-8")
	c.Data(http.StatusOK, "text/yaml", data)
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>YuteBlockchain API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css">
  <style>
    html { box-sizing: border-box; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #0f172a; }
    .topbar { display: none !important; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: '/docs/spec',
      dom_id: '#swagger-ui',
      deepLinking: true,
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIBundle.SwaggerUIStandalonePreset
      ],
      layout: 'BaseLayout',
      defaultModelsExpandDepth: -1,
      docExpansion: 'list',
      filter: true,
      syntaxHighlight: { activate: true, theme: 'monokai' }
    });
  </script>
</body>
</html>`
