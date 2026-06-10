package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) ListDatabases(w http.ResponseWriter, r *http.Request) {
	// TODO: get database connection from session
	// TODO: call driver.ListDatabases
	// TODO: return JSON array

	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		http.Error(w, "missing session ID", http.StatusBadRequest)
		return
	}

	session, ok := h.Sessions.Get(sessionID)
	if !ok {
		http.Error(w, "invalid or non-existent session ID", http.StatusBadRequest)
		return
	}

	databases, err := h.Drivers[session.Driver]().ListDatabases(session.DB)
	if err != nil {
		http.Error(w, "failed to list databases: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(databases)
}
