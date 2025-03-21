package runner

import (
	"github.com/lansfy/gonkex/checker"
	"github.com/lansfy/gonkex/models"
)

var _ checker.ExtendedCheckerInterface = (*checkersList)(nil)

type checkersList struct {
	children []checker.CheckerInterface
}

func (l *checkersList) BeforeTest(v models.TestInterface) error {
	for _, child := range l.children {
		ex, ok := child.(checker.ExtendedCheckerInterface)
		if !ok {
			continue
		}
		err := ex.BeforeTest(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (l *checkersList) Check(v models.TestInterface, result *models.Result) ([]error, error) {
	var all []error
	for _, child := range l.children {
		errs, err := child.Check(v, result)
		if err != nil {
			return nil, err
		}
		all = append(all, errs...)
	}
	return all, nil
}

func (l *checkersList) AddCheckers(i ...checker.CheckerInterface) {
	l.children = append(l.children, i...)
}
