package yaml_file

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/testloader"
)

type FileParseFun func(filePath string, content []byte) ([]*TestDefinition, error)

type LoaderOpts struct {
	CustomFileParse FileParseFun
	TemplateFuncs   template.FuncMap
}

type yamlFileLoader struct {
	testsLocation string
	opts          LoaderOpts
	filterFunc    func(fileName string) bool
}

func NewLoader(testsLocation string) testloader.LoaderInterface {
	return NewFileLoader(testsLocation, nil)
}

func NewFileLoader(testsLocation string, opts *LoaderOpts) testloader.LoaderInterface {
	l := &yamlFileLoader{
		testsLocation: testsLocation,
	}
	if opts != nil {
		l.opts = *opts
	}
	if l.opts.CustomFileParse == nil {
		l.opts.CustomFileParse = DefaultFileParse
	}
	return l
}

func (l *yamlFileLoader) Load() ([]models.TestInterface, error) {
	_, err := os.Stat(l.testsLocation)
	if err != nil && os.IsNotExist(err) {
		return nil, fmt.Errorf("file or directory with tests '%s' does not exist", l.testsLocation)
	}

	var tests []models.TestInterface
	err = filepath.WalkDir(l.testsLocation, func(relpath string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !isYmlFile(relpath) || !l.fitsFilter(relpath) {
			return err
		}

		moreTests, err := parseTestDefinitionFile(&l.opts, relpath)
		if err != nil {
			return err
		}

		tests = append(tests, moreTests...)
		return nil
	})

	return tests, err
}

func (l *yamlFileLoader) SetFilter(filterFunc func(fileName string) bool) {
	l.filterFunc = filterFunc
}

func (l *yamlFileLoader) fitsFilter(fileName string) bool {
	return l.filterFunc == nil || l.filterFunc(fileName)
}

func isYmlFile(name string) bool {
	return strings.HasSuffix(name, ".yaml") || strings.HasSuffix(name, ".yml")
}
