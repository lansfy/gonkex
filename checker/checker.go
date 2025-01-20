package checker

import (
	"errors"

	"github.com/lansfy/gonkex/models"
)

var (
	ErrTestSkipped = errors.New("test was skipped")
	ErrTestBroken  = errors.New("test was broken")
)

type CheckerInterface interface {
	Check(models.TestInterface, *models.Result) ([]error, error)
}

type ExtendedCheckerInterface interface {
	BeforeTest(models.TestInterface) error
	Check(models.TestInterface, *models.Result) ([]error, error)
}
