package output

import (
	"github.com/lansfy/gonkex/models"
)

type OutputInterface interface {
	Process(models.TestInterface, *models.Result) error
}
