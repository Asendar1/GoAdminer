package handler

import "net/http"

func (h *Handler) ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	// TODO: parse JSON body {sql: "..."}
	// TODO: detect if it's a SELECT or DML/DDL
	// TODO: for SELECT: db.Query -> return columns + rows
	// TODO: for DML/DDL: db.Exec -> return affected rows
	// TODO: return model.QueryResult
	http.Error(w, "not implemented", http.StatusNotImplemented)
}
