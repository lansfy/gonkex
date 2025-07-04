package testfixtures

import (
	"database/sql"
	"fmt"
	"strings"
)

const (
	paramTypeDollar = iota + 1
	paramTypeQuestion
	paramTypeAtSign
)

type loadFunction func(tx *sql.Tx) error

type helper interface {
	init(*sql.DB) error
	disableReferentialIntegrity(*sql.DB, loadFunction) error
	paramType() int
	databaseName(Queryable) (string, error)
	tableNames(Queryable) ([]string, error)
	isTableModified(Queryable, string) (bool, error)
	computeTablesChecksum(Queryable) error
	quoteKeyword(string) string
	whileInsertOnTable(*sql.Tx, string, func() error) error
	cleanTableQuery(string) string
	buildInsertSQL(q Queryable, tableName string, columns, values []string) (string, error)
}

var (
	_ helper = &clickhouse{}
	_ helper = &spanner{}
	_ helper = &mySQL{}
	_ helper = &postgreSQL{}
	_ helper = &sqlite{}
	_ helper = &sqlserver{}
)

type baseHelper struct{}

func (baseHelper) init(_ *sql.DB) error {
	return nil
}

func (baseHelper) quoteKeyword(str string) string {
	return fmt.Sprintf(`"%s"`, str)
}

func (baseHelper) whileInsertOnTable(_ *sql.Tx, _ string, fn func() error) error {
	return fn()
}

func (baseHelper) isTableModified(_ Queryable, _ string) (bool, error) {
	return true, nil
}

func (baseHelper) computeTablesChecksum(_ Queryable) error {
	return nil
}

func (baseHelper) cleanTableQuery(tableName string) string {
	return fmt.Sprintf("DELETE FROM %s", tableName)
}

func (h baseHelper) buildInsertSQL(_ Queryable, tableName string, columns, values []string) (string, error) {
	return fmt.Sprintf(
		"INSERT INTO %s (%s) VALUES (%s)",
		tableName,
		strings.Join(columns, ", "),
		strings.Join(values, ", "),
	), nil
}
