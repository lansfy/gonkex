package testloader

import (
	"github.com/lansfy/gonkex/models"
)

type LoaderInterface interface {
	Load() ([]models.TestInterface, error)
}
