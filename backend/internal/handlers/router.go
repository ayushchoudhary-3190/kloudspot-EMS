package handlers

import (
	"net/http"

	"EMS/docs"
	"EMS/internal/config"
	"EMS/internal/service"
)

// NewRouter wires up all routes and global middleware.
func NewRouter(cfg config.Config, empSvc *service.EmployeeService, authSvc *service.AuthService) http.Handler {
	mux := http.NewServeMux()

	emp := NewEmployeeHandler(empSvc)
	authH := NewAuthHandler(authSvc)

	protected := func(h http.HandlerFunc) http.Handler {
		return RequireAuth(cfg.JWTSecret, h)
	}

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	mux.HandleFunc("POST /api/auth/login", authH.Login)

	mux.Handle("GET /api/employees", protected(emp.List))
	mux.Handle("POST /api/employees", protected(emp.Create))
	mux.Handle("GET /api/employees/export", protected(emp.Export))
	mux.Handle("GET /api/employees/{id}", protected(emp.Get))
	mux.Handle("PUT /api/employees/{id}", protected(emp.Update))
	mux.Handle("DELETE /api/employees/{id}", protected(emp.Delete))
	mux.Handle("GET /api/departments", protected(emp.Departments))

	// OpenAPI spec + Swagger UI.
	mux.HandleFunc("GET /openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		_, _ = w.Write(docs.OpenAPISpec)
	})
	mux.HandleFunc("GET /swagger", serveSwaggerUI)

	return Logging(CORS(cfg.CORSOrigins)(mux))
}

func serveSwaggerUI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(swaggerHTML))
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>EMS API - Swagger UI</title>
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
        presets: [SwaggerUIBundle.presets.apis],
      });
    };
  </script>
</body>
</html>`
