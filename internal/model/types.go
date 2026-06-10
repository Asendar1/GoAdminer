package model

type DriverType string

const (
	DriverPostgres DriverType = "postgres"
	DriverSQLite   DriverType = "sqlite"
)

type ConnConfig struct {
	Driver   DriverType `json:"driver"`
	Host     string     `json:"host,omitempty"`
	Port     int        `json:"port,omitempty"`
	User     string     `json:"user,omitempty"`
	Password string    `json:"password,omitempty"`
	Database string     `json:"database,omitempty"`
	FilePath string     `json:"filepath,omitempty"`
	Schema   string     `json:"schema,omitempty"`
	SSLMode  string     `json:"ssl_mode,omitempty"`
}

type TableInfo struct {
	Name    string `json:"name"`
	Schema  string `json:"schema"`
	Type    string `json:"type"`
	RowsEst int64  `json:"rows_estimate"`
}

type ColumnInfo struct {
	Name         string  `json:"name"`
	DataType     string  `json:"data_type"`
	Nullable     bool    `json:"nullable"`
	Default      *string `json:"default"`
	IsPK         bool    `json:"is_pk"`
	IsFK         bool    `json:"is_fk"`
	FKRefTable   *string `json:"fk_ref_table,omitempty"`
	FKRefColumn  *string `json:"fk_ref_column,omitempty"`
	AutoIncr     bool    `json:"auto_increment"`
	MaxLen       *int    `json:"max_length,omitempty"`
	NumericPrec  *int    `json:"numeric_precision,omitempty"`
	NumericScale *int    `json:"numeric_scale,omitempty"`
}

type ForeignKey struct {
	Name       string `json:"name"`
	Column     string `json:"column"`
	RefTable   string `json:"ref_table"`
	RefColumn  string `json:"ref_column"`
	OnDelete   string `json:"on_delete"`
	OnUpdate   string `json:"on_update"`
}

type Index struct {
	Name    string   `json:"name"`
	Columns []string `json:"columns"`
	Unique  bool     `json:"unique"`
	Primary bool     `json:"primary"`
}

type TableSchema struct {
	Columns  []ColumnInfo `json:"columns"`
	PKs      []string     `json:"pks"`
	FKs      []ForeignKey `json:"fks"`
	Indexes  []Index      `json:"indexes"`
}

type RowResult struct {
	Columns []string         `json:"columns"`
	Rows    []map[string]any `json:"rows"`
	Total   int              `json:"total"`
	Page    int              `json:"page"`
	PerPage int              `json:"per_page"`
}

type QueryResult struct {
	Columns  []string         `json:"columns,omitempty"`
	Rows     []map[string]any `json:"rows,omitempty"`
	Affected int64            `json:"affected,omitempty"`
	Error    string           `json:"error,omitempty"`
}

type ConnectRequest struct {
	Driver   string `json:"driver"`
	Host     string `json:"host,omitempty"`
	Port     int    `json:"port,omitempty"`
	User     string `json:"user,omitempty"`
	Password string `json:"password,omitempty"`
	Database string `json:"database,omitempty"`
	FilePath string `json:"filepath,omitempty"`
	SSLMode  string `json:"ssl_mode,omitempty"`
	Schema	 string `json:"schema,omitempty"`
}

type ConnectResponse struct {
	SessionID string `json:"session_id"`
	Driver    string `json:"driver"`
}

type StatusResponse struct {
	Connected bool   `json:"connected"`
	Driver    string `json:"driver,omitempty"`
	Database  string `json:"database,omitempty"`
	Schema    string `json:"schema,omitempty"`
}
