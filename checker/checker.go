package checker

import (
	"errors"

	"github.com/lansfy/gonkex/models"
)

var (
	// ErrTestSkipped is returned when a test is marked with "skipped" status in the YAML file
	ErrTestSkipped = errors.New("test was skipped")
	// ErrTestBroken is returned when a test is marked with "broken" status in the YAML file
	ErrTestBroken = errors.New("test was broken")
)

// CheckerInterface defines the basic interface for test result verification
// Implementations of this interface can verify if test results match expected outputs
// Custom checkers implementing this interface should be registered using the AddCheckers method
type CheckerInterface interface {
	// Check compares the actual test result against the expected result
	// Returns a list of validation errors and/or a critical error that stopped the test
	Check(models.TestInterface, *models.Result) ([]error, error)
}

// ExtendedCheckerInterface extends CheckerInterface with pre-test functionality
// This allows for test preparation before execution (for example, make some customizations before running the test)
// Custom checkers implementing this interface should be registered using the AddCheckers method
type ExtendedCheckerInterface interface {
	// BeforeTest runs before test execution to perform setup or validation
	// If BeforeTest returns an error, the test will not run and will be marked as failed
	BeforeTest(models.TestInterface) error

	// Check compares the actual test result against the expected result
	// Returns a list of validation errors and/or a critical error that stopped the test
	Check(models.TestInterface, *models.Result) ([]error, error)
}
