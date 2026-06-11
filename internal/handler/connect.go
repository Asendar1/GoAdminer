package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Asendar1/GoAdminer/internal/model"
)

func (h *Handler) Connect(w http.ResponseWriter, r *http.Request) {
	var req model.ConnectRequest
	dec := json.NewDecoder(r.Body)

	// * This Checks for: extra fields and unknown ones.
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid request: " + err.Error(), http.StatusBadRequest)
		return
	}

	_, ok := h.Drivers[req.Driver]
	if !ok {
		http.Error(w, "unsupported driver", http.StatusBadRequest)
		return
	}

	// * Space trimming is handled by the frontend.
	// * FilePath is only used for SQLite.
	cfg := model.ConnConfig{
		Driver: model.DriverType(req.Driver),
		Host: req.Host,
		Port: req.Port,
		User: req.User,
		Password: req.Password,
		Database: req.Database,
		FilePath: req.FilePath,
		Schema: req.Schema,
		SSLMode: req.SSLMode,
	}

	dbStruct := h.Drivers[req.Driver]()
	url := dbStruct.DSN(cfg)

	conn, err := dbStruct.Open(url)
	if err != nil {
		http.Error(w, "failed to connect: " + err.Error(), http.StatusInternalServerError)
		return
	}

	err = conn.Ping()
	if err != nil {
		conn.Close()
		http.Error(w, "failed to connect: " + err.Error(), http.StatusInternalServerError)
		return
	}

	sessionID := h.Sessions.New(conn, req.Driver, req.Database, req.Schema)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.ConnectResponse{SessionID: sessionID, Driver: req.Driver})
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		http.Error(w, "missing session ID", http.StatusBadRequest)
		return
	}

	conn, ok := h.Sessions.Get(sessionID)
	if !ok {
		http.Error(w, "invalid or non-existent session ID", http.StatusBadRequest)
		return
	}

	err := conn.DB.Ping()
	if err != nil {
		http.Error(w, "database connection error: " + err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.StatusResponse{Connected: true})
}

func (h *Handler) Disconnect(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Header.Get("X-Session-ID")
	if sessionID == "" {
		http.Error(w, "missing session ID", http.StatusBadRequest)
		return
	}

	_, ok := h.Sessions.Get(sessionID)
	if !ok {
		http.Error(w, "invalid or non-existent session ID", http.StatusBadRequest)
		return
	}

	h.Sessions.Delete(sessionID)

	// * Frontend deletes the cookie
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(model.StatusResponse{Connected: false})
}
