package driver

import (
	"database/sql"

	"goadminer/internal/model"
)

type Driver interface {
	Open(dsn string) (*sql.DB, error)

	ListDatabases(db *sql.DB) ([]string, error)
	ListTables(db *sql.DB, schema string) ([]model.TableInfo, error)
	TableColumns(db *sql.DB, schema, table string) ([]model.ColumnInfo, error)
	PrimaryKeys(db *sql.DB, schema, table string) ([]string, error)
	ForeignKeys(db *sql.DB, schema, table string) ([]model.ForeignKey, error)
	Indexes(db *sql.DB, schema, table string) ([]model.Index, error)

	CountRows(db *sql.DB, schema, table string, where string) (int, error)
	SelectRows(db *sql.DB, schema, table string, columns []string, where string, order string, limit, offset int) ([]map[string]any, error)
	Insert(db *sql.DB, schema, table string, data map[string]any) (map[string]any, error)
	Update(db *sql.DB, schema, table string, data map[string]any, pk map[string]any) error
	Delete(db *sql.DB, schema, table string, pk map[string]any) error

	QuoteIdent(name string) string
	Placeholder(n int) string
	DSN(cfg model.ConnConfig) string

	Close(db *sql.DB) error
}
