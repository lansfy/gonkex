package sqldb

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type SQLType string

const (
	PostgreSQL  SQLType = "postgresql"
	MySQL       SQLType = "mysql"
	Sqlite      SQLType = "sqlite"
	TimescaleDB SQLType = "timescaledb"
	MariaDB     SQLType = "mariadb"
	SQLServer   SQLType = "sqlserver"
	ClickHouse  SQLType = "clickhouse"
)

type Storage struct {
	dbType SQLType
	db     *sql.DB
	opts   *StorageOpts
}

type StorageOpts struct {
}

func NewStorage(dbType SQLType, db *sql.DB, opts *StorageOpts) (*Storage, error) {
	switch dbType {
	case PostgreSQL, MySQL, Sqlite, TimescaleDB, MariaDB, SQLServer, ClickHouse:
	default:
		return nil, fmt.Errorf("unknown db type %q", dbType)
	}
	return &Storage{
		dbType: dbType,
		db:     db,
		opts:   opts,
	}, nil
}

func (l *Storage) GetName() string {
	return l.GetType()
}

func (l *Storage) GetType() string {
	return string(l.dbType)
}

func (l *Storage) LoadFixtures(location string, names []string) error {
	return LoadFixtures(l.dbType, l.db, location, names)
}

func (l *Storage) ExecuteQuery(query string) ([]json.RawMessage, error) {
	return ExecuteQuery(l.dbType, l.db, query)
}
