package runner

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"text/template"

	"github.com/lansfy/gonkex/checker"
	"github.com/lansfy/gonkex/checker/response_body"
	"github.com/lansfy/gonkex/checker/response_db"
	"github.com/lansfy/gonkex/checker/response_header"
	"github.com/lansfy/gonkex/mocks"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/output"
	"github.com/lansfy/gonkex/output/terminal"
	"github.com/lansfy/gonkex/storage"
	"github.com/lansfy/gonkex/testloader/yaml_file"
	"github.com/lansfy/gonkex/variables"

	"github.com/joho/godotenv"
)

var DefaultOutput = terminal.NewOutput(nil)

type RunWithTestingOpts struct {
	TestsDir    string
	FixturesDir string
	EnvFilePath string

	Mocks          *mocks.Mocks
	DB             storage.StorageInterface
	MainOutputFunc output.OutputInterface
	Outputs        []output.OutputInterface
	Checkers       []checker.CheckerInterface

	CustomClient       Client
	TemplateReplyFuncs template.FuncMap
}

func registerMocksEnvironment(m *mocks.Mocks) {
	names := m.GetNames()
	for _, n := range names {
		varName := fmt.Sprintf("GONKEX_MOCK_%s", strings.ToUpper(n))
		os.Setenv(varName, m.Service(n).ServerAddr())
	}
}

// RunWithTesting is a helper function the wraps the common Run and provides simple way
// to configure Gonkex by filling the params structure.
func RunWithTesting(t *testing.T, serverURL string, opts *RunWithTestingOpts) {
	if opts.Mocks != nil {
		registerMocksEnvironment(opts.Mocks)
	}

	if opts.EnvFilePath != "" {
		if err := godotenv.Load(opts.EnvFilePath); err != nil {
			t.Fatal(err)
		}
	}

	var proxyURL *url.URL
	if os.Getenv("HTTP_PROXY") != "" {
		httpURL, err := url.Parse(os.Getenv("HTTP_PROXY"))
		if err != nil {
			t.Fatal(err)
		}
		proxyURL = httpURL
	}

	yamlLoader := yaml_file.NewLoader(opts.TestsDir)
	yamlLoader.SetFileFilter(os.Getenv("GONKEX_FILE_FILTER"))

	handler := testingHandler{t}
	runner := New(
		&Config{
			Host:  serverURL,
			Mocks: opts.Mocks,
			MocksLoader: mocks.NewYamlLoader(&mocks.YamlLoaderOpts{
				TemplateReplyFuncs: opts.TemplateReplyFuncs,
			}),
			FixturesDir:  opts.FixturesDir,
			DB:           opts.DB,
			Variables:    variables.New(),
			HTTPProxyURL: proxyURL,
		},
		yamlLoader,
		handler.HandleTest,
	)

	addOutputs(runner, opts)
	addCheckers(runner, opts)

	err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func addOutputs(runner *Runner, opts *RunWithTestingOpts) {
	if opts.MainOutputFunc != nil {
		runner.AddOutput(opts.MainOutputFunc)
	} else {
		runner.AddOutput(DefaultOutput)
	}

	for _, o := range opts.Outputs {
		runner.AddOutput(o)
	}
}

func addCheckers(runner *Runner, opts *RunWithTestingOpts) {
	runner.AddCheckers(response_body.NewChecker())
	runner.AddCheckers(response_header.NewChecker())
	if opts.DB != nil {
		runner.AddCheckers(response_db.NewChecker(opts.DB))
	}
	runner.AddCheckers(opts.Checkers...)
}

type testingHandler struct {
	t *testing.T
}

func (h testingHandler) HandleTest(test models.TestInterface, executeTest testExecutor) error {
	var returnErr error
	h.t.Run(test.GetName(), func(t *testing.T) {
		result, err := executeTest(test)
		if err != nil {
			if errors.Is(err, errTestSkipped) || errors.Is(err, errTestBroken) {
				t.Skip()
			} else {
				returnErr = err
				t.Fatal(err)
			}
		}

		if !result.Passed() {
			t.Fail()
		}
	})

	return returnErr
}
