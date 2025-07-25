package yaml_file

import (
	"sort"

	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/testloader"
)

type YamlInMemoryLoader struct {
	files      map[string]string
	opts       LoaderOpts
	filterFunc func(fileName string) bool
}

func NewInMemoryLoader(files map[string]string, opts *LoaderOpts) testloader.LoaderInterface {
	l := &YamlInMemoryLoader{
		files: files,
	}
	if opts != nil {
		l.opts = *opts
	}
	if l.opts.CustomFileRead == nil {
		l.opts.CustomFileRead = DefaultFileRead
	}
	return l
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
		moreTests, err := parseTestDefinitionContent(l.opts.CustomFileRead,
			relpath, []byte(l.files[relpath]))
		if err != nil {
			return nil, err
		}

		tests = append(tests, moreTests...)
	}

	return tests, nil
}

func (l *YamlInMemoryLoader) SetFilter(filterFunc func(fileName string) bool) {
	l.filterFunc = filterFunc
}
