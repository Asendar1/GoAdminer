package driver

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/Asendar1/GoAdminer/internal/model"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresDriver struct{}

func NewPostgres() Driver { return &PostgresDriver{} }

func (d *PostgresDriver) DSN(cfg model.ConnConfig) string {
	host := cfg.Host
	if host == "" {
		host = "localhost"
	}
	port := cfg.Port
	if port == 0 {
		port = 5432
	}
	ssl := cfg.SSLMode
	if ssl == "" {
		ssl = "disable"
	}
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, host, port, cfg.Database, ssl)
}

func (d *PostgresDriver) Open(dsn string) (*sql.DB, error) {
	return sql.Open("pgx", dsn)
}

func (d *PostgresDriver) Close(db *sql.DB) error {
	return db.Close()
}

func (d *PostgresDriver) QuoteIdent(name string) string {
	return `"` + name + `"`
}

func (d *PostgresDriver) Placeholder(n int) string {
	return fmt.Sprintf("$%d", n)
}

func (d *PostgresDriver) ListDatabases(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT datname FROM pg_database WHERE datistemplate = false ORDER BY datname")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var dbs []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		dbs = append(dbs, name)
	}
	return dbs, rows.Err()
}

func (d *PostgresDriver) ListTables(db *sql.DB, schema string) ([]model.TableInfo, error) {
	if schema == "" {
		schema = "public"
	}
	rows, err := db.Query(`
		SELECT table_name, table_schema, table_type
		FROM information_schema.tables
		WHERE table_schema = $1
		ORDER BY table_name`, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []model.TableInfo
	for rows.Next() {
		var t model.TableInfo
		if err := rows.Scan(&t.Name, &t.Schema, &t.Type); err != nil {
			return nil, err
		}
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (d *PostgresDriver) TableColumns(db *sql.DB, schema, table string) ([]model.ColumnInfo, error) {
	if schema == "" {
		schema = "public"
	}
	rows, err := db.Query(`
		SELECT
			c.column_name,
			c.data_type,
			c.is_nullable,
			c.column_default,
			c.character_maximum_length,
			c.numeric_precision,
			c.numeric_scale,
			c.is_identity
		FROM information_schema.columns c
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cols []model.ColumnInfo
	for rows.Next() {
		var c model.ColumnInfo
		var nullable string
		var isIdentity string
		if err := rows.Scan(&c.Name, &c.DataType, &nullable, &c.Default,
				&c.MaxLen, &c.NumericPrec, &c.NumericScale, &isIdentity); err != nil {
			return nil, err
		}
		c.Nullable = nullable == "YES"
		c.AutoIncr = isIdentity == "YES" || (c.Default != nil && strings.Contains(*c.Default, "nextval("))
		cols = append(cols, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	pks, err := d.PrimaryKeys(db, schema, table)
	if err != nil {
		return nil, err
	}
	pkSet := make(map[string]bool, len(pks))
	for _, pk := range pks {
		pkSet[pk] = true
	}
	for i := range cols {
		if pkSet[cols[i].Name] {
			cols[i].IsPK = true
		}
	}
	return cols, nil
}

func (d *PostgresDriver) PrimaryKeys(db *sql.DB, schema, table string) ([]string, error) {
	if schema == "" {
		schema = "public"
	}
	rows, err := db.Query(`
		SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		WHERE tc.constraint_type = 'PRIMARY KEY'
			AND tc.table_schema = $1
			AND tc.table_name = $2
		ORDER BY kcu.ordinal_position`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var pks []string
	for rows.Next() {
		var pk string
		if err := rows.Scan(&pk); err != nil {
			return nil, err
		}
		pks = append(pks, pk)
	}
	return pks, rows.Err()
}

func (d *PostgresDriver) ForeignKeys(db *sql.DB, schema, table string) ([]model.ForeignKey, error) {
	if schema == "" {
		schema = "public"
	}
	rows, err := db.Query(`
		SELECT
			tc.constraint_name,
			kcu.column_name,
			ccu.table_name AS ref_table,
			ccu.column_name AS ref_column,
			rc.update_rule,
			rc.delete_rule
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_name = kcu.constraint_name
			AND tc.table_schema = kcu.table_schema
		JOIN information_schema.constraint_column_usage ccu
			ON ccu.constraint_name = tc.constraint_name
			AND ccu.table_schema = tc.table_schema
		JOIN information_schema.referential_constraints rc
			ON rc.constraint_name = tc.constraint_name
			AND rc.constraint_schema = tc.table_schema
		WHERE tc.constraint_type = 'FOREIGN KEY'
			AND tc.table_schema = $1
			AND tc.table_name = $2`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var fks []model.ForeignKey
	for rows.Next() {
		var fk model.ForeignKey
		if err := rows.Scan(&fk.Name, &fk.Column, &fk.RefTable, &fk.RefColumn, &fk.OnUpdate, &fk.OnDelete); err != nil {
			return nil, err
		}
		fks = append(fks, fk)
	}
	return fks, rows.Err()
}

func (d *PostgresDriver) Indexes(db *sql.DB, schema, table string) ([]model.Index, error) {
	if schema == "" {
		schema = "public"
	}
	rows, err := db.Query(`
		SELECT
			i.indexname,
			i.indexdef
		FROM pg_indexes i
		WHERE i.schemaname = $1 AND i.tablename = $2
		ORDER BY i.indexname`, schema, table)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var indexes []model.Index
	for rows.Next() {
		var idx model.Index
		var def string
		if err := rows.Scan(&idx.Name, &def); err != nil {
			return nil, err
		}
		idx.Unique = contains(def, "UNIQUE")
		idx.Primary = contains(def, "PRIMARY KEY")
		indexes = append(indexes, idx)
	}
	return indexes, rows.Err()
}

func (d *PostgresDriver) CountRows(db *sql.DB, schema, table string, where string, args []any) (int, error) {
	q := fmt.Sprintf("SELECT COUNT(*) FROM %s.%s", d.QuoteIdent(schema), d.QuoteIdent(table))
	if where != "" {
		q += " WHERE " + where
	}
	var count int
	if err := db.QueryRow(q, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (d *PostgresDriver) SelectRows(db *sql.DB, schema, table string, columns []string, where string, args []any, order string, limit, offset int) ([]map[string]any, error) {
	selCols := "*"
	if len(columns) > 0 {
		selCols = joinQuoted(columns, d.QuoteIdent)
	}
	q := fmt.Sprintf("SELECT %s FROM %s.%s", selCols, d.QuoteIdent(schema), d.QuoteIdent(table))
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

func (d *PostgresDriver) Insert(db *sql.DB, schema, table string, data map[string]any) (map[string]any, error) {
	cols := make([]string, 0, len(data))
	vals := make([]any, 0, len(data))
	i := 1
	placeholders := make([]string, 0, len(data))
	for k, v := range data {
		cols = append(cols, d.QuoteIdent(k))
		vals = append(vals, v)
		placeholders = append(placeholders, d.Placeholder(i))
		i++
	}
	q := fmt.Sprintf("INSERT INTO %s.%s (%s) VALUES (%s) RETURNING *",
		d.QuoteIdent(schema), d.QuoteIdent(table),
		joinQuoted(cols, func(s string) string { return s }),
		joinStrings(placeholders, ", "))
	rows, err := db.Query(q, vals...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanRow(rows)
}

func (d *PostgresDriver) Update(db *sql.DB, schema, table string, data map[string]any, pk map[string]any) error {
	setClauses := make([]string, 0, len(data))
	vals := make([]any, 0, len(data))
	i := 1
	for k, v := range data {
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", d.QuoteIdent(k), d.Placeholder(i)))
		vals = append(vals, v)
		i++
	}
	whereClauses := make([]string, 0, len(pk))
	for k, v := range pk {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = %s", d.QuoteIdent(k), d.Placeholder(i)))
		vals = append(vals, v)
		i++
	}
	q := fmt.Sprintf("UPDATE %s.%s SET %s WHERE %s",
		d.QuoteIdent(schema), d.QuoteIdent(table),
		joinStrings(setClauses, ", "),
		joinStrings(whereClauses, " AND "))
	_, err := db.Exec(q, vals...)
	return err
}

func (d *PostgresDriver) Delete(db *sql.DB, schema, table string, pk map[string]any) error {
	whereClauses := make([]string, 0, len(pk))
	vals := make([]any, 0, len(pk))
	i := 1
	for k, v := range pk {
		whereClauses = append(whereClauses, fmt.Sprintf("%s = %s", d.QuoteIdent(k), d.Placeholder(i)))
		vals = append(vals, v)
		i++
	}
	q := fmt.Sprintf("DELETE FROM %s.%s WHERE %s",
		d.QuoteIdent(schema), d.QuoteIdent(table),
		joinStrings(whereClauses, " AND "))
	_, err := db.Exec(q, vals...)
	return err
}
