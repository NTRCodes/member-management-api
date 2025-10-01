package httpx

import (
	"fmt"
	"net/http"
)

// RegisterRedoc sets up the Redoc documentation endpoint
func RegisterRedoc(mux *http.ServeMux) {
	// Serve swagger.json at a different path to avoid route conflicts
	mux.HandleFunc("/api-spec.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		http.ServeFile(w, r, "/app/docs/swagger.json")
	})

	// Redirect /docs (without trailing slash) to /redoc
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redoc", http.StatusMovedPermanently)
	})

	// Redirect /docs/ UI to /redoc (less specific)
	mux.HandleFunc("/docs/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/redoc", http.StatusMovedPermanently)
	})

	// Main documentation endpoint
	mux.HandleFunc("/redoc", func(w http.ResponseWriter, r *http.Request) {
		html := `<!DOCTYPE html>
<html>
<head>
    <title>Organization API Documentation</title>
    <meta charset="utf-8"/>
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://fonts.googleapis.com/css?family=Montserrat:300,400,700|Roboto:300,400,700" rel="stylesheet">
    <style>
        body { margin: 0; padding: 0; }
    </style>
</head>
<body>
    <redoc spec-url='/api-spec.json'></redoc>
    <script src="https://cdn.jsdelivr.net/npm/redoc@2.1.3/bundles/redoc.standalone.js"></script>
</body>
</html>`
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	})
}
