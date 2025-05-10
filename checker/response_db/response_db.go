package response_db

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/lansfy/gonkex/checker"
	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/compare"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/storage"

	"github.com/kylelemons/godebug/diff"
	"github.com/kylelemons/godebug/pretty"
)

func NewChecker(aliasedDB map[string]storage.StorageInterface) checker.CheckerInterface {
	return &responseDbChecker{
		aliasedDB: aliasedDB,
	}
}

type responseDbChecker struct {
	aliasedDB map[string]storage.StorageInterface
}

func (c *responseDbChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
	var errors []error
	for _, dbCheck := range t.GetDatabaseChecks() {
		// check expected db query exist
		if dbCheck.DbQueryString() == "" {
			return nil, fmt.Errorf("dbQuery not found in the test declaration")
		}

		// check expected response exist
		if dbCheck.DbResponseJson() == nil {
			return nil, fmt.Errorf("dbResponse not found in the test declaration")
		}

		aliases := dbCheck.DbAliases()
		if len(aliases) == 0 {
			aliases = []string{storage.DefaultDBAlias}
		}

		for _, alias := range aliases {
			db, ok := c.aliasedDB[alias]
			if !ok {
				return nil, fmt.Errorf("unknown database alias '%s'", alias)
			}
			if db == nil {
				continue
			}
			errs, err := c.check(alias, db, dbCheck, result)
			if err != nil {
				return nil, err
			}
			errors = append(errors, errs...)
			break
		}
	}

	return errors, nil
}

func (c *responseDbChecker) check(alias string, db storage.StorageInterface,
	t models.DatabaseCheck, result *models.Result) ([]error, error) {
	// get DB response
	actualDbResponse, err := newQuery(t.DbQueryString(), db)
	if err != nil {
		return nil, err
	}

	result.DatabaseResult = append(
		result.DatabaseResult,
		models.DatabaseResult{
			Query: t.DbQueryString(),
			Response: actualDbResponse,
			Alias: alias,
		},
	)

	// compare responses length
	err = compareDbResponseLength(t.DbResponseJson(), actualDbResponse, t.DbQueryString())
	if err != nil {
		return []error{err}, nil
	}
	// compare responses as json lists
	expectedItems, err := toJSONArray(t.DbResponseJson(), "dbResponse in the test declaration")
	if err != nil {
		return nil, err
	}
	actualItems, err := toJSONArray(actualDbResponse, "database response")
	if err != nil {
		return nil, err
	}

	cmpOptions := t.GetComparisonParams()

	return compare.Compare(expectedItems, actualItems, compare.Params{
		IgnoreValues:         cmpOptions.IgnoreValuesChecking(),
		IgnoreArraysOrdering: cmpOptions.IgnoreArraysOrdering(),
		DisallowExtraFields:  cmpOptions.DisallowExtraFields(),
	}), nil
}

func toJSONArray(items []string, qual string) ([]interface{}, error) {
	itemJSONs := make([]interface{}, 0, len(items))
	for i, row := range items {
		var itemJSON interface{}
		if err := json.Unmarshal([]byte(row), &itemJSON); err != nil {
			return nil, fmt.Errorf(
				"invalid JSON in the %s:\n row #%d:\n %s\n error:\n%w",
				qual, i, row, err,
			)
		}
		itemJSONs = append(itemJSONs, itemJSON)
	}

	return itemJSONs, nil
}

func compareDbResponseLength(expected, actual []string, query interface{}) error {
	if len(expected) == len(actual) {
		return nil
	}

	diffCfg := *pretty.DefaultConfig
	diffCfg.Diffable = true
	chunks := diff.DiffChunks(
		strings.Split(diffCfg.Sprint(expected), "\n"),
		strings.Split(diffCfg.Sprint(actual), "\n"),
	)

	tail := []colorize.Part{
		colorize.None("\n\n   query: "),
		colorize.Cyan(query),
		colorize.None("\n   diff (--- expected vs +++ actual):\n"),
	}
	tail = append(tail, colorize.MakeColorDiff(chunks)...)

	return colorize.NewNotEqualError(
		"quantity of %s do not match:",
		"items in database",
		len(expected),
		len(actual),
	).AddParts(tail...)
}

func newQuery(dbQuery string, db storage.StorageInterface) ([]string, error) {
	messages, err := db.ExecuteQuery(dbQuery)
	if err != nil {
		return nil, err
	}

	dbResponse := []string{}
	for _, item := range messages {
		data, err := item.MarshalJSON()
		if err != nil {
			return nil, err
		}

		dbResponse = append(dbResponse, string(data))
	}

	return dbResponse, nil
}
