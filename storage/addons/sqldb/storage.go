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

func NewStorage(dbType SQLType, db *sql.DB, opts *StorageOpts) *Storage {
	return &Storage{
		dbType: dbType,
		db:     db,
		opts:   opts,
	}
}

func (l *Storage) GetType() string {
	return string(l.dbType)
}

const virtualFileName = "fake.yml"

func (l *Storage) LoadFixtures(location string, names []string) error {
	if err := l.checkDbType(); err != nil {
		return err
	}

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

	loader, err := testfixtures.New(createFixtureParams(l.dbType, l.db, vfs)...)
	if err != nil {
		return err
	}
	return loader.Load()
}

func (l *Storage) ExecuteQuery(query string) ([]json.RawMessage, error) {
	if err := l.checkDbType(); err != nil {
		return nil, err
	}
	return ExecuteQuery(l.dbType, l.db, query)
}

func (l *Storage) checkDbType() error {
	switch l.dbType {
	case PostgreSQL, MySQL, Sqlite, TimescaleDB, MariaDB, SQLServer, ClickHouse:
		return nil
	default:
		return fmt.Errorf("unknown db type %q", l.dbType)
	}
}

// createFixtureParams allows to redefine parameters for test purpose
var createFixtureParams = func(dbType SQLType, db *sql.DB, fs fstest.MapFS) []func(*testfixtures.Loader) error {
	return []func(*testfixtures.Loader) error{
		testfixtures.Database(db),
		testfixtures.Dialect(string(dbType)),
		testfixtures.FS(fs),
		testfixtures.FilesMultiTables(virtualFileName),
		testfixtures.DangerousSkipTestDatabaseCheck(),
		testfixtures.SkipTableChecksumComputation(),
		testfixtures.ResetSequencesTo(1),
	}
}
