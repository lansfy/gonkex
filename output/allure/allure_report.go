package allure

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/lansfy/gonkex/models"
)

type Output struct {
	reportLocation string
	allure         Allure
}

func NewOutput(suiteName, reportLocation string) (*Output, error) {
	resultsDir, err := filepath.Abs(reportLocation)
	if err != nil {
		return nil, err
	}
	err = os.Mkdir(resultsDir, 0o777)
	if err != nil {
		return nil, err
	}

	a := Allure{
		Suites:    nil,
		TargetDir: resultsDir,
	}
	a.StartSuite(suiteName, time.Now())

	return &Output{
		reportLocation: reportLocation,
		allure:         a,
	}, nil
}

func (o *Output) Process(t models.TestInterface, result *models.Result) error {
	description := t.GetDescription()
	if description == "" {
		description = "No description"
	}

	testCase := o.allure.StartCase(t.GetName(), time.Now())
	testCase.SetDescription(description)
	testCase.AddLabel("story", result.Path)

	o.allure.AddAttachment("Request", fmt.Sprintf("Query: %s\n Body: %s", result.Query, result.RequestBody), "txt")
	o.allure.AddAttachment("Response", fmt.Sprintf("Body: %s", result.ResponseBody), "txt")

	for i, dbresult := range result.DatabaseResult {
		if dbresult.Query != "" {
			o.allure.AddAttachment(fmt.Sprintf("Db Query #%d", i+1), fmt.Sprintf("SQL string: %s", dbresult.Query), "txt")
			o.allure.AddAttachment(fmt.Sprintf("Db Response #%d", i+1), fmt.Sprintf("Response: %s", dbresult.Response), "txt")
		}
	}

	status, err := getAllureStatus(result)
	o.allure.EndCase(status, err, time.Now())
	return nil
}

func (o *Output) Finalize() error {
	return o.allure.EndSuite(time.Now())
}

func notRunnedStatus(status models.Status) bool {
	return status == models.StatusBroken || status == models.StatusSkipped
}

func getAllureStatus(r *models.Result) (string, error) {
	testStatus := r.Test.GetStatus()
	if testStatus != models.StatusNone && notRunnedStatus(testStatus) {
		return string(testStatus), nil
	}

	if len(r.Errors) == 0 {
		return "passed", nil
	}

	errText := ""
	for _, err := range r.Errors {
		errText += err.Error() + "\n"
	}

	return "failed", errors.New(errText)
}
