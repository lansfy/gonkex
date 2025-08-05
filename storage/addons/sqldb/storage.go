package sqldb

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"testing/fstest"

	"github.com/lansfy/gonkex/storage/addons/fixtures"
	"github.com/lansfy/gonkex/storage/addons/sqldb/testfixtures"
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

func (l *Storage) GetType() string {
	return string(l.dbType)
}

func (l *Storage) LoadFixtures(location string, names []string) error {
	const virtualFileName = "fake.yml"

	opts := &fixtures.LoadDataOpts{
		AllowedTypes: []string{"tables"},
		CustomActions: map[string]func(string) string{
			"eval": func(value string) string {
				return "RAW=" + value
			},
		},
	}

	coll, err := fixtures.LoadData(fixtures.CreateFileLoader(location), names, opts)
	if err != nil {
		return fmt.Errorf("load fixtures: %w", err)
	}

	data, err := fixtures.DumpCollection(coll, false)
	if err != nil {
		return fmt.Errorf("generate global fixtures: %w", err)
	}

	vfs := fstest.MapFS{
		virtualFileName: &fstest.MapFile{
			Data: data,
		},
	}

	loader, err := testfixtures.New(
		testfixtures.Database(l.db),
		testfixtures.Dialect(string(l.dbType)),
		testfixtures.FS(vfs),
		testfixtures.FilesMultiTables(virtualFileName),
		testfixtures.DangerousSkipTestDatabaseCheck(),
		testfixtures.SkipTableChecksumComputation(),
		testfixtures.ResetSequencesTo(1),
	)
	if err != nil {
		return err
	}
	return loader.Load()
}

func (l *Storage) ExecuteQuery(query string) ([]json.RawMessage, error) {
	return ExecuteQuery(l.dbType, l.db, query)
}
