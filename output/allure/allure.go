package allure

import (
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/lansfy/gonkex/output/allure/beans"
)

type Allure struct {
	Suites    []*beans.Suite
	TargetDir string
}

func New(suites []*beans.Suite) *Allure {
	return &Allure{Suites: suites, TargetDir: "allure-results"}
}

func (a *Allure) GetCurrentSuite() *beans.Suite {
	return a.Suites[0]
}

func (a *Allure) StartSuite(name string, start time.Time) {
	a.Suites = append(a.Suites, beans.NewSuite(name, start))
}

func (a *Allure) EndSuite(end time.Time) error {
	suite := a.GetCurrentSuite()
	suite.SetEnd(end)
	if suite.HasTests() {
		if _, err := writeSuite(a.TargetDir, suite); err != nil {
			return err
		}
	}
	// remove first/current suite
	a.Suites = a.Suites[1:]

	return nil
}

var (
	currentState = map[*beans.Suite]*beans.TestCase{}
	currentStep  = map[*beans.Suite]*beans.Step{}
)

func (a *Allure) StartCase(testName string, start time.Time) *beans.TestCase {
	test := beans.NewTestCase(testName, start)
	step := beans.NewStep(testName, start)
	suite := a.GetCurrentSuite()
	currentState[suite] = test
	currentStep[suite] = step
	suite.AddTest(test)

	return test
}

func (a *Allure) EndCase(status string, err error, end time.Time) {
	suite := a.GetCurrentSuite()
	test, ok := currentState[suite]
	if ok {
		test.End(status, err, end)
	}
}

func (a *Allure) CreateStep(name string, stepFunc func()) {
	status := "passed"
	a.StartStep(name, time.Now())
	// if test error
	stepFunc()
	// end
	a.EndStep(status, time.Now())
}

func (a *Allure) StartStep(stepName string, start time.Time) {
	var (
		// FIXME: step is overwritten below
		// step  = beans.NewStep(stepName, start)
		suite = a.GetCurrentSuite()
	)
	step := currentStep[suite]
	step.Parent.AddStep(step)
	currentStep[suite] = step
}

func (a *Allure) EndStep(status string, end time.Time) {
	suite := a.GetCurrentSuite()
	currentStep[suite].End(status, end)
	currentStep[suite] = currentStep[suite].Parent
}

func (a *Allure) AddAttachment(attachmentName, content string, typ string) {
	mime := "text/plain"
	ext := "txt"
	name, _ := writeAttachment(a.TargetDir, content, ext)
	currentState[a.GetCurrentSuite()].AddAttachment(beans.NewAttachment(
		attachmentName,
		mime,
		name,
		len(content)))
}

func (a *Allure) PendingCase(testName string, start time.Time) {
	a.StartCase(testName, start)
	a.EndCase("pending", errors.New("test ignored"), start)
}

func writeAttachment(pathDir string, content string, ext string) (string, error) {
	fileName := uuid.New().String() + "-attachment." + ext
	err := os.WriteFile(filepath.Join(pathDir, fileName), []byte(content), 0o644)
	return fileName, err
}

func writeSuite(pathDir string, suite *beans.Suite) (string, error) {
	fileName := uuid.New().String() + "-testsuite.xml"
	b, err := xml.Marshal(suite)
	if err != nil {
		return fileName, err
	}
	err = os.WriteFile(filepath.Join(pathDir, fileName), b, 0o644)
	return fileName, err
}
