package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Asendar1/GoAdminer/internal/model"
)

func (h *Handler) ExecuteQuery(w http.ResponseWriter, r *http.Request) {
	sess, err := h.getSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	var body struct {
		SQL string `json:"sql"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	sqlStr := strings.TrimSpace(body.SQL)
	if sqlStr == "" {
		http.Error(w, "empty SQL query", http.StatusBadRequest)
		return
	}

	var res model.QueryResult

	if returnsRows(sqlStr) {
		rows, err := sess.DB.Query(sqlStr)
		if err != nil {
			res.Error = err.Error()
		} else {
			columns, err := rows.Columns()
			if err != nil {
				res.Error = err.Error()
				rows.Close()
			} else {
				res.Columns = columns
				res.Rows = queryScanAll(rows, columns)
				rows.Close()
				if err := rows.Err(); err != nil {
					res.Error = err.Error()
				}
			}
		}
	} else {
		result, err := sess.DB.Exec(sqlStr)
		if err != nil {
			res.Error = err.Error()
		} else {
			res.Affected, _ = result.RowsAffected()
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func returnsRows(sql string) bool {
	upper := strings.TrimSpace(strings.ToUpper(sql))
	for _, prefix := range []string{"SELECT", "WITH", "EXPLAIN", "SHOW", "DESCRIBE"} {
		if strings.HasPrefix(upper, prefix) {
			return true
		}
	}
	return strings.Contains(upper, " RETURNING")
}

func queryScanAll(rows *sql.Rows, columns []string) []map[string]any {
	var result []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return result
		}
		row := make(map[string]any, len(columns))
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				row[col] = val
			}
		}
		result = append(result, row)
	}
	return result
}
