package handler

import "net/http"

func (h *Handler) ListDatabases(w http.ResponseWriter, r *http.Request) {
	// TODO: get database connection from session
	// TODO: call driver.ListDatabases
	// TODO: return JSON array
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
