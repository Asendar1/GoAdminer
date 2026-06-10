package handler

import "net/http"

func (h *Handler) ListTables(w http.ResponseWriter, r *http.Request) {
	// TODO: get schema from query param (?schema=public)
	// TODO: get DB connection from session
	// TODO: call driver.ListTables
	// TODO: return JSON
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) TableSchema(w http.ResponseWriter, r *http.Request) {
	// TODO: get table name from URL param
	// TODO: get DB connection from session
	// TODO: call driver.TableColumns, PrimaryKeys, ForeignKeys, Indexes
	// TODO: return JSON
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
