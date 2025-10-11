package http

import (
	"embed"
	"io"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
)

//go:embed swagger-ui/*
var swaggerUI embed.FS

//go:embed openapi.yaml
var openapi []byte

// RegisterDocsRoutes mounts OpenAPI spec and a small Swagger UI under /docs
func RegisterDocsRoutes(r chi.Router) {
	// serve spec
	r.Get("/docs/openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(openapi)
	})

	// serve swagger UI static files under /docs/
	sub, _ := fs.Sub(swaggerUI, "swagger-ui")
	handler := http.StripPrefix("/docs/", http.FileServer(http.FS(sub)))

	// static assets (css/js) -> /docs/*
	r.Handle("/docs/*", handler)

	// redirect bare /docs to /docs/
	r.Get("/docs", func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, "/docs/", http.StatusTemporaryRedirect)
	})

	// serve index on /docs/ by reading index.html from the embedded FS
	r.Get("/docs/", func(w http.ResponseWriter, req *http.Request) {
		f, err := sub.Open("index.html")
		if err != nil {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		defer f.Close()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = io.Copy(w, f)
	})
}
