package output

import (
	"github.com/lansfy/gonkex/models"
)

// OutputInterface defines a contract for output test results.
// It requires implementing a Process method that accepts a TestInterface
// and a Result, returning an error if processing fails.
type OutputInterface interface {
	Process(models.TestInterface, *models.Result) error
}

// ExtendedOutputInterface extends OutputInterface by adding a BeforeTest method.
// BeforeTest is executed before a test is run, and Process handles the result processing.
type ExtendedOutputInterface interface {
	BeforeTest(models.TestInterface) error
	Process(models.TestInterface, *models.Result) error
}
