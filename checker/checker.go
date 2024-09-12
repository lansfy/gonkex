package checker

import "github.com/lansfy/gonkex/models"

type CheckerInterface interface {
	Check(models.TestInterface, *models.Result) ([]error, error)
}
