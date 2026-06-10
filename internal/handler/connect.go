package handler

import (
	"encoding/json"
	"net/http"

	"goadminer/internal/model"
)

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	// TODO: parse request body into model.ConnectRequest
	// TODO: determine driver type (postgres/sqlite)
	// TODO: build DSN from config
	// TODO: open database connection
	// TODO: create session, return session ID
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	// TODO: get session ID from cookie
	// TODO: look up session in store
	// TODO: ping database, return status
	json.NewEncoder(w).Encode(model.StatusResponse{Connected: false})
}

func (h *Handler) Disconnect(w http.ResponseWriter, r *http.Request) {
	// TODO: get session ID from cookie
	// TODO: delete session from store
	// TODO: clear cookie
	w.WriteHeader(http.StatusOK)
}
