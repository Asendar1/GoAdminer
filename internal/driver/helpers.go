package driver

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/Asendar1/GoAdminer/internal/model"
)

func contains(s, substr string) bool {
	return strings.Contains(strings.ToUpper(s), strings.ToUpper(substr))
}

func joinQuoted(items []string, quote func(string) string) string {
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = quote(item)
	}
	return joinStrings(quoted, ", ")
}

func joinStrings(items []string, sep string) string {
	return strings.Join(items, sep)
}

func queryRows(db *sql.DB, query string, args ...any) ([]map[string]any, error) {
	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var result []map[string]any
	for rows.Next() {
		values := make([]any, len(columns))
		valuePtrs := make([]any, len(columns))
		for i := range columns {
			valuePtrs[i] = &values[i]
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, err
		}
		row := make(map[string]any)
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
	return result, rows.Err()
}

func CoerceTypes(data map[string]any, columns []model.ColumnInfo) {
	if data == nil {
		return
	}
	for _, col := range columns {
		val, ok := data[col.Name]
		if !ok {
			continue
		}
		if val == nil {
			continue
		}
		switch col.DataType {
		case "ARRAY":
			switch v := val.(type) {
			case string:
				list := strings.Split(v, ",")
				res := make([]string, 0, len(list))
				for _, item := range list {
					trimmed := strings.TrimSpace(item)
					if trimmed != "" {
						res = append(res, trimmed)
					}
				}
				if len(res) == 1 && strings.HasPrefix(res[0], "[") {
					var parsed []string
					if json.Unmarshal([]byte(v), &parsed) == nil {
						data[col.Name] = parsed
						break
					}
				}
				data[col.Name] = res
			case []any:
				res := make([]string, 0, len(v))
				for _, item := range v {
					if s, ok := item.(string); ok {
						res = append(res, s)
					}
				}
				data[col.Name] = res
			}
		case "jsonb", "json":
			switch v := val.(type) {
			case map[string]any, []any:
				encoded, err := json.Marshal(v)
				if err == nil {
					data[col.Name] = string(encoded)
				}
			}
		}
	}
}

func scanRow(rows *sql.Rows) (map[string]any, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	if !rows.Next() {
		return nil, sql.ErrNoRows
	}
	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}
	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, err
	}
	row := make(map[string]any)
	for i, col := range columns {
		val := values[i]
		if b, ok := val.([]byte); ok {
			row[col] = string(b)
		} else {
			row[col] = val
		}
	}
	return row, rows.Err()
}
