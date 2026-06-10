package handler

import "net/http"

func (h *Handler) ListRows(w http.ResponseWriter, r *http.Request) {
	// TODO: get table from URL param
	// TODO: parse query params: page, per_page, sort, order, search
	// TODO: get DB connection from session
	// TODO: call driver.CountRows
	// TODO: call driver.SelectRows with limit/offset
	// TODO: return model.RowResult
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) InsertRow(w http.ResponseWriter, r *http.Request) {
	// TODO: parse JSON body into map[string]any
	// TODO: get table from URL param
	// TODO: get DB connection from session
	// TODO: call driver.Insert
	// TODO: return inserted row
	http.Error(w, "not implemented", http.StatusNotImplemented)
}

func (h *Handler) UpdateRow(w http.ResponseWriter, r *http.Request) {
	// TODO: parse JSON body {data: {...}, pk: {...}}
	// TODO: get table from URL param
	// TODO: get DB connection from session
	// TODO: call driver.Update
	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteRow(w http.ResponseWriter, r *http.Request) {
	// TODO: parse JSON body {pk: {...}}
	// TODO: get table from URL param
	// TODO: get DB connection from session
	// TODO: call driver.Delete
	w.WriteHeader(http.StatusOK)
}
