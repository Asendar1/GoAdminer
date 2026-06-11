package handler

import (
	"encoding/json"
	"net/http"

	"github.com/Asendar1/GoAdminer/internal/model"
)

func (h *Handler) ListTables(w http.ResponseWriter, r *http.Request) {
	schema := r.URL.Query().Get("schema")
	if schema == "" {
		schema = "public"
	}

	sessionID := r.Header.Get("X-Sesson-ID")

	session, ok := h.Sessions.Get(sessionID)
	if !ok {
		http.Error(w, "invalid session ID", http.StatusBadRequest)
		return
	}

	tables, err := h.Drivers[session.Driver]().ListTables(session.DB, schema)
	if err != nil {
		http.Error(w, "failed to list tables", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tables)
}

func (h *Handler) TableSchema(w http.ResponseWriter, r *http.Request) {
	table := r.URL.Query().Get("table")
	if table == "" {
		http.Error(w, "missing table name", http.StatusBadRequest)
		return
	}

	sessionID := r.Header.Get("X-Session-ID")
	session, ok := h.Sessions.Get(sessionID)
	if !ok {
		http.Error(w, "invalid session ID", http.StatusBadRequest)
		return
	}

	var rtn model.TableSchema
	driver := h.Drivers[session.Driver]()
	db := session.DB

	cols, err := driver.TableColumns(db, session.Schema, table)
	if err != nil {
		http.Error(w, "Failed to get table schema", http.StatusInternalServerError)
		return
	}

	PKs, err := driver.PrimaryKeys(db, session.Schema, table)
	if err != nil {
		http.Error(w, "Failed to get primary keys", http.StatusInternalServerError)
		return
	}

	FKs, err := driver.ForeignKeys(db, session.Schema, table)
	if err != nil {
		http.Error(w, "Failed to get foreign keys", http.StatusInternalServerError)
		return
	}

	idxs, err := driver.Indexes(db, session.Schema, table)
	if err != nil {
		http.Error(w, "Failed to get indexes", http.StatusInternalServerError)
		return
	}

	rtn.Columns = cols
	rtn.PKs = PKs
	rtn.FKs = FKs
	rtn.Indexes = idxs

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(rtn)

}
