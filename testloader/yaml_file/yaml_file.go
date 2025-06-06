package yaml_file

import (
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/testloader"
)

var _ testloader.LoaderInterface = (*YamlInMemoryLoader)(nil)

type YamlFileLoader struct {
	testsLocation string
	fileFilter    string
}

func NewLoader(testsLocation string) *YamlFileLoader {
	return &YamlFileLoader{
		testsLocation: testsLocation,
	}
}

func (l *YamlFileLoader) Load() ([]models.TestInterface, error) {
	var tests []models.TestInterface
	err := filepath.WalkDir(l.testsLocation, func(relpath string, d fs.DirEntry, err error) error {
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
			tests = append(tests, &moreTests[i])
		}

		return nil
	})

	return tests, err
}

func (l *YamlFileLoader) SetFileFilter(f string) {
	l.fileFilter = f
}

func (l *YamlFileLoader) fitsFilter(fileName string) bool {
	if l.fileFilter == "" {
		return true
	}

	return strings.Contains(fileName, l.fileFilter)
}

func isYmlFile(name string) bool {
	return strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}
