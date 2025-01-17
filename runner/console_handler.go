package runner

import (
	"errors"

	"github.com/lansfy/gonkex/checker"
	"github.com/lansfy/gonkex/models"
)

type ConsoleHandler struct {
	totalTests   int
	failedTests  int
	skippedTests int
	brokenTests  int
}

func NewConsoleHandler() *ConsoleHandler {
	return &ConsoleHandler{}
}

func (h *ConsoleHandler) HandleTest(test models.TestInterface, executeTest TestExecutor) error {
	testResult, err := executeTest(test)
	switch {
	case err != nil && errors.Is(err, checker.ErrTestSkipped):
		h.skippedTests++
	case err != nil && errors.Is(err, checker.ErrTestBroken):
		h.brokenTests++
	case err != nil:
		return err
	}

	h.totalTests++
	if !testResult.Passed() {
		h.failedTests++
	}

	return nil
}

func (h *ConsoleHandler) Summary() *models.Summary {
	return &models.Summary{
		Success: h.failedTests == 0,
		Skipped: h.skippedTests,
		Broken:  h.brokenTests,
		Failed:  h.failedTests,
		Total:   h.totalTests,
	}
}
