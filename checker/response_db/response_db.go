package response_db

import (
	"encoding/json"
	"fmt"

	"github.com/lansfy/gonkex/checker"
	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/compare"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/storage"

	"github.com/kylelemons/godebug/pretty"
)

type ResponseDbChecker struct {
	db storage.StorageInterface
}

func NewChecker(db storage.StorageInterface) checker.CheckerInterface {
	return &ResponseDbChecker{
		db: db,
	}
}

func (c *ResponseDbChecker) Check(t models.TestInterface, result *models.Result) ([]error, error) {
	var errors []error
	for _, dbCheck := range t.GetDatabaseChecks() {
		errs, err := c.check(t.GetName(), dbCheck, result)
		if err != nil {
			return nil, err
		}
		errors = append(errors, errs...)
	}

	return errors, nil
}

func (c *ResponseDbChecker) check(
	testName string,
	t models.DatabaseCheck,
	result *models.Result,
) ([]error, error) {
	var errors []error
	// check expected db query exist
	if t.DbQueryString() == "" {
		return nil, fmt.Errorf("DB query not found for test \"%s\"", testName)
	}

	// check expected response exist
	if t.DbResponseJson() == nil {
		return nil, fmt.Errorf("expected DB response not found for test \"%s\"", testName)
	}

	// get DB response
	actualDbResponse, err := newQuery(t.DbQueryString(), c.db)
	if err != nil {
		return nil, err
	}

	result.DatabaseResult = append(
		result.DatabaseResult,
		models.DatabaseResult{Query: t.DbQueryString(), Response: actualDbResponse},
	)

	// compare responses length
	err = compareDbResponseLength(t.DbResponseJson(), actualDbResponse, t.DbQueryString())
	if err != nil {
		return append(errors, err), nil
	}
	// compare responses as json lists
	expectedItems, err := toJSONArray(t.DbResponseJson(), "expected", testName)
	if err != nil {
		return nil, err
	}
	actualItems, err := toJSONArray(actualDbResponse, "actual", testName)
	if err != nil {
		return nil, err
	}

	cmpOptions := t.GetComparisonParams()

	errs := compare.Compare(expectedItems, actualItems, compare.Params{
		IgnoreValues:         cmpOptions.IgnoreValuesChecking(),
		IgnoreArraysOrdering: cmpOptions.IgnoreArraysOrdering(),
		DisallowExtraFields:  cmpOptions.DisallowExtraFields(),
	})

	errors = append(errors, errs...)

	return errors, nil
}

func toJSONArray(items []string, qual, testName string) ([]interface{}, error) {
	itemJSONs := make([]interface{}, 0, len(items))
	for i, row := range items {
		var itemJSON interface{}
		if err := json.Unmarshal([]byte(row), &itemJSON); err != nil {
			return nil, fmt.Errorf(
				"invalid JSON in the %s DB response for test %s:\n row #%d:\n %s\n error:\n%s",
				qual,
				testName,
				i,
				row,
				err.Error(),
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

	return colorize.NewError(
		colorize.None("quantity of items in database do not match (-expected: "),
		colorize.Cyan(len(expected)),
		colorize.None(" +actual: "),
		colorize.Cyan(len(actual)),
		colorize.None(")\n     test query:\n"),
		colorize.Cyan(query),
		colorize.None("\n    result diff:\n"),
		colorize.Cyan(pretty.Compare(expected, actual)),
	)
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
