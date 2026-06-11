package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Asendar1/GoAdminer/internal/driver"
	"github.com/Asendar1/GoAdminer/internal/model"
)

func (h *Handler) ListRows(w http.ResponseWriter, r *http.Request) {
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

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	if perPage < 1 || perPage > 500 {
		perPage = 50
	}
	sortCol := r.URL.Query().Get("sort")
	sortOrder := r.URL.Query().Get("order")
	search := r.URL.Query().Get("search")

	columns, err := drv.TableColumns(sess.DB, sess.Schema, table)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	where, args := buildSearchWhere(columns, search, drv)
	order := buildOrderBy(columns, sortCol, sortOrder, drv)

	total, err := drv.CountRows(sess.DB, sess.Schema, table, where, args)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	offset := (page - 1) * perPage
	rows, err := drv.SelectRows(sess.DB, sess.Schema, table, nil, where, args, order, perPage, offset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	colNames := make([]string, len(columns))
	for i, c := range columns {
		colNames[i] = c.Name
	}

	res := model.RowResult{
		Columns: colNames,
		Rows:    rows,
		Total:   total,
		Page:    page,
		PerPage: perPage,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *Handler) InsertRow(w http.ResponseWriter, r *http.Request) {
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

	var data map[string]any
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	row, err := drv.Insert(sess.DB, sess.Schema, table, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(row)
}

func (h *Handler) UpdateRow(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		Data map[string]any `json:"data"`
		PK   map[string]any `json:"pk"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := drv.Update(sess.DB, sess.Schema, table, body.Data, body.PK); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteRow(w http.ResponseWriter, r *http.Request) {
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

	var body struct {
		PK map[string]any `json:"pk"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := drv.Delete(sess.DB, sess.Schema, table, body.PK); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func buildSearchWhere(columns []model.ColumnInfo, search string, drv driver.Driver) (string, []any) {
	if search == "" {
		return "", nil
	}
	parts := make([]string, 0, len(columns))
	args := make([]any, 0, len(columns))
	for i, col := range columns {
		parts = append(parts, fmt.Sprintf("LOWER(CAST(%s AS TEXT)) LIKE LOWER(%s)",
			drv.QuoteIdent(col.Name), drv.Placeholder(i+1)))
		args = append(args, "%"+search+"%")
	}
	return "(" + strings.Join(parts, " OR ") + ")", args
}

func buildOrderBy(columns []model.ColumnInfo, sort, order string, drv driver.Driver) string {
	if sort == "" {
		return ""
	}
	for _, col := range columns {
		if col.Name == sort {
			if order != "desc" {
				order = "asc"
			}
			return fmt.Sprintf("%s %s", drv.QuoteIdent(sort), order)
		}
	}
	return ""
}
