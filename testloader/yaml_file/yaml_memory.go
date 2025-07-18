package yaml_file

import (
	"sort"

	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/testloader"
)

var _ testloader.LoaderInterface = (*YamlInMemoryLoader)(nil)

type YamlInMemoryLoader struct {
	files      map[string]string
	filterFunc func(fileName string) bool
}

func NewInMemoryLoader(files map[string]string) *YamlInMemoryLoader {
	return &YamlInMemoryLoader{
		files: files,
	}
}

func (l *YamlInMemoryLoader) Load() ([]models.TestInterface, error) {
	var tests []models.TestInterface

	keys := make([]string, 0, len(l.files))
	for k := range l.files {
		if l.filterFunc != nil && !l.filterFunc(k) {
			continue
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, relpath := range keys {
		moreTests, err := parseTestDefinitionContent(relpath, []byte(l.files[relpath]))
		if err != nil {
			return nil, err
		}

		for i := range moreTests {
			tests = append(tests, moreTests[i])
		}
	}

	return tests, nil
}

func (l *YamlInMemoryLoader) SetFilter(filterFunc func(fileName string) bool) {
	l.filterFunc = filterFunc
}
