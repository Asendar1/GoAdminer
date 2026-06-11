package server

import (
	"embed"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Asendar1/GoAdminer/internal/handler"
)

func New(h *handler.Handler, webFS *embed.FS, devMode bool) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(corsMiddleware)

	r.Route("/api", func(r chi.Router) {
		r.Post("/connect", h.Connect)
		r.Get("/status", h.Status)
		r.Post("/disconnect", h.Disconnect)

		r.Get("/databases", h.ListDatabases)

		r.Get("/tables", h.ListTables)
		r.Get("/tables/{table}/schema", h.TableSchema)

		r.Get("/tables/{table}/rows", h.ListRows)
		r.Post("/tables/{table}/rows", h.InsertRow)
		r.Put("/tables/{table}/rows", h.UpdateRow)
		r.Delete("/tables/{table}/rows", h.DeleteRow)

		r.Post("/query", h.ExecuteQuery)
	})

	if devMode {
		fsys := http.FileServer(http.Dir("web"))
		r.Handle("/*", fsys)
	} else if webFS != nil {
		sub, err := fs.Sub(*webFS, "web")
		if err != nil {
			panic(err)
		}
		r.Handle("/*", http.FileServer(http.FS(sub)))
	}

	return r
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-ID")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}


