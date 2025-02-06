package allure

import (
	"bytes"
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

func NewOutput(suiteName, reportLocation string) *Output {
	resultsDir, _ := filepath.Abs(reportLocation)
	_ = os.Mkdir(resultsDir, 0o777)
	a := Allure{
		Suites:    nil,
		TargetDir: resultsDir,
	}
	a.StartSuite(suiteName, time.Now())

	return &Output{
		reportLocation: reportLocation,
		allure:         a,
	}
}

func (o *Output) Process(t models.TestInterface, result *models.Result) error {
	testCase := o.allure.StartCase(t.GetName(), time.Now())
	testCase.SetDescriptionOrDefaultValue(t.GetDescription(), "No description")
	testCase.AddLabel("story", result.Path)

	o.allure.AddAttachment(
		*bytes.NewBufferString("Request"),
		*bytes.NewBufferString(fmt.Sprintf(`Query: %s\n Body: %s`, result.Query, result.RequestBody)),
		"txt")
	o.allure.AddAttachment(
		*bytes.NewBufferString("Response"),
		*bytes.NewBufferString(fmt.Sprintf(`Body: %s`, result.ResponseBody)),
		"txt")

	for i, dbresult := range result.DatabaseResult {
		if dbresult.Query != "" {
			o.allure.AddAttachment(
				*bytes.NewBufferString(fmt.Sprintf("Db Query #%d", i+1)),
				*bytes.NewBufferString(fmt.Sprintf(`SQL string: %s`, dbresult.Query)),
				"txt")
			o.allure.AddAttachment(
				*bytes.NewBufferString(fmt.Sprintf("Db Response #%d", i+1)),
				*bytes.NewBufferString(fmt.Sprintf(`Response: %s`, dbresult.Response)),
				"txt")
		}
	}

	status, err := getAllureStatus(result)
	o.allure.EndCase(status, err, time.Now())

	return nil
}

func (o *Output) Finalize() {
	_ = o.allure.EndSuite(time.Now())
}

func notRunnedStatus(status models.Status) bool {
	switch status {
	case models.StatusBroken, models.StatusSkipped:
		return true
	default:
		return false
	}
}

func getAllureStatus(r *models.Result) (string, error) {
	testStatus := r.Test.GetStatus()
	if testStatus != models.StatusNone && notRunnedStatus(testStatus) {
		return string(testStatus), nil
	}

	var (
		status     = "passed"
		testErrors []error
	)

	if len(r.Errors) != 0 {
		status = "failed"
		testErrors = r.Errors
	}

	if len(testErrors) != 0 {
		errText := ""
		for _, err := range testErrors {
			errText = errText + err.Error() + "\n"
		}

		return status, errors.New(errText)
	}

	return status, nil
}
