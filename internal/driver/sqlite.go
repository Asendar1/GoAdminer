package driver

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Asendar1/GoAdminer/internal/model"
	_ "modernc.org/sqlite"
)

type SQLiteDriver struct{}

func NewSQLite() Driver { return &SQLiteDriver{} }

func (d *SQLiteDriver) DSN(cfg model.ConnConfig) string {
	return cfg.FilePath
}

func (d *SQLiteDriver) Open(dsn string) (*sql.DB, error) {
	return sql.Open("sqlite", dsn)
}

func (d *SQLiteDriver) Close(db *sql.DB) error {
	return db.Close()
}

func (d *SQLiteDriver) QuoteIdent(name string) string {
	return `"` + name + `"`
}

func (d *SQLiteDriver) Placeholder(_ int) string {
	return "?"
}

func (d *SQLiteDriver) ListDatabases(db *sql.DB) ([]string, error) {
	return []string{"main"}, nil
}

func (d *SQLiteDriver) ListTables(db *sql.DB, schema string) ([]model.TableInfo, error) {
	var q string
	var args []any
	if schema == "" || schema == "main" {
		q = `SELECT name FROM sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
	} else {
		q = `SELECT name FROM ` + d.QuoteIdent(schema) + `.sqlite_master WHERE type = 'table' AND name NOT LIKE 'sqlite_%' ORDER BY name`
	}
	rows, err := db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []model.TableInfo
	for rows.Next() {
		var t model.TableInfo
		if err := rows.Scan(&t.Name); err != nil {
			return nil, err
		}
		t.Schema = schema
		if t.Schema == "" {
			t.Schema = "main"
		}
		t.Type = "TABLE"
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (d *SQLiteDriver) TableColumns(db *sql.DB, schema, table string) ([]model.ColumnInfo, error) {
	q := fmt.Sprintf("PRAGMA table_info(%s)", d.QuoteIdent(table))
	if schema != "" && schema != "main" {
		q = fmt.Sprintf("PRAGMA %s.table_info(%s)", d.QuoteIdent(schema), d.QuoteIdent(table))
	}
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []model.ColumnInfo
	for rows.Next() {
		var cid int
		var c model.ColumnInfo
		var nullable int
		var dflt *string
		var pk int
		if err := rows.Scan(&cid, &c.Name, &c.DataType, &nullable, &dflt, &pk); err != nil {
			return nil, err
		}
		c.Nullable = nullable == 1
		c.Default = dflt
		c.IsPK = pk == 1
		c.AutoIncr = c.IsPK && strings.EqualFold(c.DataType, "INTEGER")
		cols = append(cols, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return cols, nil
}

func (d *SQLiteDriver) PrimaryKeys(db *sql.DB, schema, table string) ([]string, error) {
	cols, err := d.TableColumns(db, schema, table)
	if err != nil {
		return nil, err
	}
	var pks []string
	for _, c := range cols {
		if c.IsPK {
			pks = append(pks, c.Name)
		}
	}
	return pks, nil
}

func (d *SQLiteDriver) ForeignKeys(db *sql.DB, schema, table string) ([]model.ForeignKey, error) {
	q := fmt.Sprintf("PRAGMA foreign_key_list(%s)", d.QuoteIdent(table))
	if schema != "" && schema != "main" {
		q = fmt.Sprintf("PRAGMA %s.foreign_key_list(%s)", d.QuoteIdent(schema), d.QuoteIdent(table))
	}
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fks []model.ForeignKey
	for rows.Next() {
		var id, seq int
		var fk model.ForeignKey
		var onUpdate, onDelete string
		var match string
		if err := rows.Scan(&id, &seq, &fk.RefTable, &fk.Column, &fk.RefColumn, &onUpdate, &onDelete, &match); err != nil {
			return nil, err
		}
		fk.OnUpdate = onUpdate
		fk.OnDelete = onDelete
		fks = append(fks, fk)
	}
	return fks, rows.Err()
}

func (d *SQLiteDriver) Indexes(db *sql.DB, schema, table string) ([]model.Index, error) {
	q := fmt.Sprintf("PRAGMA index_list(%s)", d.QuoteIdent(table))
	if schema != "" && schema != "main" {
		q = fmt.Sprintf("PRAGMA %s.index_list(%s)", d.QuoteIdent(schema), d.QuoteIdent(table))
	}
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var indexes []model.Index
	for rows.Next() {
		var seq int
		var idx model.Index
		var unique int
		var origin, partial string
		if err := rows.Scan(&seq, &idx.Name, &unique, &origin, &partial); err != nil {
			return nil, err
		}
		idx.Unique = unique == 1
		indexes = append(indexes, idx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i, idx := range indexes {
		cols, err := d.indexColumns(db, schema, idx.Name)
		if err != nil {
			return nil, err
		}
		indexes[i].Columns = cols
	}
	return indexes, nil
}

func (d *SQLiteDriver) indexColumns(db *sql.DB, schema, indexName string) ([]string, error) {
	var q string
	if schema == "" || schema == "main" {
		q = fmt.Sprintf("PRAGMA index_info(%s)", d.QuoteIdent(indexName))
	} else {
		q = fmt.Sprintf("PRAGMA %s.index_info(%s)", d.QuoteIdent(schema), d.QuoteIdent(indexName))
	}
	rows, err := db.Query(q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var seq int
		var cid int
		var name string
		if err := rows.Scan(&seq, &cid, &name); err != nil {
			return nil, err
		}
		cols = append(cols, name)
	}
	return cols, rows.Err()
}

func (d *SQLiteDriver) CountRows(db *sql.DB, schema, table string, where string, args []any) (int, error) {
	q := fmt.Sprintf("SELECT COUNT(*) FROM %s", d.QuoteIdent(table))
	if where != "" {
		q += " WHERE " + where
	}
	var count int
	if err := db.QueryRow(q, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (d *SQLiteDriver) SelectRows(db *sql.DB, schema, table string, columns []string, where string, args []any, order string, limit, offset int) ([]map[string]any, error) {
	selCols := "*"
	if len(columns) > 0 {
		selCols = joinQuoted(columns, d.QuoteIdent)
	}
	q := fmt.Sprintf("SELECT %s FROM %s", selCols, d.QuoteIdent(table))
	if where != "" {
		q += " WHERE " + where
	}
	if order != "" {
		q += " ORDER BY " + order
	}
	if limit > 0 {
		q += fmt.Sprintf(" LIMIT %d", limit)
	}
	if offset > 0 {
		q += fmt.Sprintf(" OFFSET %d", offset)
	}
	return queryRows(db, q, args...)
}

func (d *SQLiteDriver) Insert(db *sql.DB, schema, table string, data map[string]any) (map[string]any, error) {
	cols := make([]string, 0, len(data))
	vals := make([]any, 0, len(data))
	ph := make([]string, 0, len(data))
	for k, v := range data {
		cols = append(cols, d.QuoteIdent(k))
		vals = append(vals, v)
		ph = append(ph, "?")
	}
	q := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) RETURNING *",
		d.QuoteIdent(table),
		strings.Join(cols, ", "),
		strings.Join(ph, ", "))
	rows, err := db.Query(q, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRow(rows)
}

func (d *SQLiteDriver) Update(db *sql.DB, schema, table string, data map[string]any, pk map[string]any) error {
	setClauses := make([]string, 0, len(data))
	vals := make([]any, 0, len(data))
	for k, v := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = ?", d.QuoteIdent(k)))
		vals = append(vals, v)
	}
	whereClauses := make([]string, 0, len(pk))
	for k, v := range pk {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", d.QuoteIdent(k)))
		vals = append(vals, v)
	}
	q := fmt.Sprintf("UPDATE %s SET %s WHERE %s",
		d.QuoteIdent(table),
		strings.Join(setClauses, ", "),
		strings.Join(whereClauses, " AND "))
	_, err := db.Exec(q, vals...)
	return err
}

func (d *SQLiteDriver) Delete(db *sql.DB, schema, table string, pk map[string]any) error {
	whereClauses := make([]string, 0, len(pk))
	vals := make([]any, 0, len(pk))
	for k, v := range pk {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = ?", d.QuoteIdent(k)))
		vals = append(vals, v)
	}
	q := fmt.Sprintf("DELETE FROM %s WHERE %s",
		d.QuoteIdent(table),
		strings.Join(whereClauses, " AND "))
	_, err := db.Exec(q, vals...)
	return err
}
