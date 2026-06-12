package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Asendar1/GoAdminer/internal/model"
)

func (h *Handler) ListTables(w http.ResponseWriter, r *http.Request) {
	sess, err := h.getSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	drv := h.getDriver(sess.Driver)
	if drv == nil {
		http.Error(w, "unknown driver", http.StatusInternalServerError)
		return
	}

	schema := r.URL.Query().Get("schema")
	if schema == "" {
		schema = sess.Schema
	}

	tables, err := drv.ListTables(sess.DB, schema)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tables)
}

func (h *Handler) TableSchema(w http.ResponseWriter, r *http.Request) {
	sess, err := h.getSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	drv := h.getDriver(sess.Driver)
	if drv == nil {
		http.Error(w, "unknown driver", http.StatusInternalServerError)
		return
	}

	table := chi.URLParam(r, "table")
	if table == "" {
		http.Error(w, "missing table name", http.StatusBadRequest)
		return
	}

	schema := sess.Schema

	columns, err := drv.TableColumns(sess.DB, schema, table)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pks, err := drv.PrimaryKeys(sess.DB, schema, table)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fks, err := drv.ForeignKeys(sess.DB, schema, table)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fkCols := make(map[string]model.ForeignKey, len(fks))
	for _, fk := range fks {
		fkCols[fk.Column] = fk
	}
	for i := range columns {
		if fk, ok := fkCols[columns[i].Name]; ok {
			columns[i].IsFK = true
			columns[i].FKRefTable = &fk.RefTable
			columns[i].FKRefColumn = &fk.RefColumn
		}
	}

	indexes, err := drv.Indexes(sess.DB, schema, table)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	resp := model.TableSchema{
		Columns: columns,
		PKs:     pks,
		FKs:     fks,
		Indexes: indexes,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
