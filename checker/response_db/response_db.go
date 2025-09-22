package response_db

import (
	"encoding/json"
	"fmt"

	"github.com/lansfy/gonkex/checker"
	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/compare"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/storage"
)

func NewChecker(db storage.StorageInterface) checker.CheckerInterface {
	return &responseDbChecker{
		db: db,
	}
}

type responseDbChecker struct {
	db storage.StorageInterface
}

func (c *responseDbChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
	var errors []error
	for idx, dbCheck := range t.GetDatabaseChecks() {
		path := fmt.Sprintf("$.dbChecks[%d]", idx)
		errs, err := c.check(path, dbCheck, result)
		if err != nil {
			return nil, err
		}
		errors = append(errors, errs...)
	}

	return errors, nil
}

func (c *responseDbChecker) check(path string, t models.DatabaseCheck, result *models.Result) ([]error, error) {
	if t.DbQueryString() == "" {
		return nil, createDefinitionError(path, colorize.NewEntityError("%s key required", "dbQuery"))
	}

	if t.DbResponseJson() == nil {
		return nil, createDefinitionError(path, colorize.NewEntityError("%s key required", "dbResponse"))
	}

	// parse expected DB response
	expectedItems, err := unmarshalArray(path, t.DbResponseJson())
	if err != nil {
		return nil, err
	}

	// get real DB response
	actualItems, err := makeQuery(path, c.db, t.DbQueryString())
	if err != nil {
		result.DatabaseResult = append(result.DatabaseResult,
			models.DatabaseResult{Query: t.DbQueryString(), Response: []string{}},
		)
		return []error{err}, nil
	}

	result.DatabaseResult = append(result.DatabaseResult,
		models.DatabaseResult{Query: t.DbQueryString(), Response: toStringArray(actualItems)},
	)

	if len(expectedItems) != len(actualItems) {
		return []error{createDifferentLengthError(path, expectedItems, actualItems)}, nil
	}

	cmpOptions := t.GetComparisonParams()

	errs := compare.Compare(expectedItems, actualItems, compare.Params{
		IgnoreValues:         cmpOptions.IgnoreValuesChecking(),
		IgnoreArraysOrdering: cmpOptions.IgnoreArraysOrdering(),
		DisallowExtraFields:  cmpOptions.DisallowExtraFields(),
	})

	for idx := range errs {
		errs[idx] = colorize.NewEntityError("database check for %s", path+".dbResponse").WithSubError(errs[idx])
	}

	return errs, nil
}

func toStringArray(src []interface{}) []string {
	result := make([]string, len(src))
	for idx := range src {
		data, _ := json.Marshal(src[idx])
		result[idx] = string(data)
	}
	return result
}

func makeQuery(path string, db storage.StorageInterface, dbQuery string) ([]interface{}, error) {
	rawMessages, err := db.ExecuteQuery(dbQuery)
	if err != nil {
		return nil, colorize.NewEntityError("failed %s", "database check").WithSubError(
			colorize.NewPathError(path, fmt.Errorf("execute request '%s': %w", dbQuery, err)),
		)
	}

	response := make([]interface{}, len(rawMessages))
	for idx := range rawMessages {
		err := json.Unmarshal(rawMessages[idx], &response[idx])
		if err != nil {
			return nil, err
		}
	}

	return response, nil
}

func unmarshalArray(path string, items []string) ([]interface{}, error) {
	itemJSONs := make([]interface{}, 0, len(items))
	for idx, row := range items {
		var itemJSON interface{}
		if err := json.Unmarshal([]byte(row), &itemJSON); err != nil {
			return nil, createDefinitionError(fmt.Sprintf("%s.dbResponse[%d]", path, idx),
				fmt.Errorf("invalid JSON: %w", err))
		}
		itemJSONs = append(itemJSONs, itemJSON)
	}

	return itemJSONs, nil
}

func sprintWithSingleQuotes(items []interface{}) []string {
	result := []string{"["}
	for _, item := range toStringArray(items) {
		line := " '" + item + "',"
		result = append(result, line)
	}
	result = append(result, "]")
	return result
}

func createDifferentLengthError(path string, expected, actual []interface{}) error {
	tail := colorize.MakeColorDiff(
		"\n\n   diff (--- expected vs +++ actual):\n",
		sprintWithSingleQuotes(expected),
		sprintWithSingleQuotes(actual),
	)

	return colorize.NewPathError(path, colorize.NewEntityNotEqualError(
		"quantity of %s does not match:",
		"items in database",
		len(expected),
		len(actual),
	).WithPostfix(tail))
}

func createDefinitionError(path string, err error) error {
	return colorize.NewEntityError("load definition for %s", "database check").
		WithSubError(colorize.NewPathError(path, err))
}
