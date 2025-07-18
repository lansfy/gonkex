package yaml_file

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/testloader"
)

var _ testloader.LoaderInterface = (*YamlInMemoryLoader)(nil)

type YamlFileLoader struct {
	testsLocation string
	filterFunc    func(fileName string) bool
}

func NewLoader(testsLocation string) *YamlFileLoader {
	return &YamlFileLoader{
		testsLocation: testsLocation,
	}
}

func (l *YamlFileLoader) Load() ([]models.TestInterface, error) {
	_, err := os.Stat(l.testsLocation)
	if err != nil && os.IsNotExist(err) {
		return nil, fmt.Errorf("file or directory with tests '%s' does not exist", l.testsLocation)
	}

	var tests []models.TestInterface
	err = filepath.WalkDir(l.testsLocation, func(relpath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !isYmlFile(relpath) || !l.fitsFilter(relpath) {
			return nil
		}

		moreTests, err := parseTestDefinitionFile(relpath)
		if err != nil {
			return err
		}

		for i := range moreTests {
			tests = append(tests, moreTests[i])
		}

		return nil
	})

	return tests, err
}

func (l *YamlFileLoader) SetFilter(filterFunc func(fileName string) bool) {
	l.filterFunc = filterFunc
}

func (l *YamlFileLoader) fitsFilter(fileName string) bool {
	return l.filterFunc == nil || l.filterFunc(fileName)
}

func isYmlFile(name string) bool {
	return strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}
