package docs

import (
	"net/http"

	"github.com/mainguyen0112/fleetcontrol/api/openapi"
)

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

// ServeSpec serves the raw OpenAPI spec as YAML.
func (h *Handler) ServeSpec(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.Write(openapi.Spec)
}

// ServeUI serves a minimal Swagger UI page pointed at /openapi.yaml.
// Swagger UI assets are loaded from a CDN rather than vendored, to keep
// the binary and repo small — acceptable for an internal/dev-facing
// docs page, not for an air-gapped deployment.
func (h *Handler) ServeUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(swaggerUIHTML))
}

const swaggerUIHTML = `<!DOCTYPE html>
<html>
<head>
  <title>FleetControl API Docs</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = () => {
      window.ui = SwaggerUIBundle({
        url: "/openapi.yaml",
        dom_id: "#swagger-ui",
      });
    };
  </script>
</body>
</html>`
