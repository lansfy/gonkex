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
	"github.com/lansfy/gonkex/endpoint"
	"github.com/lansfy/gonkex/mocks"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/output"
	"github.com/lansfy/gonkex/output/terminal"
	"github.com/lansfy/gonkex/storage"
	"github.com/lansfy/gonkex/testloader/yaml_file"
	"github.com/lansfy/gonkex/variables"

	"github.com/joho/godotenv"
)

type RunWithTestingOpts struct {
	// TestsDir is the directory where test definitions or files are located.
	TestsDir string
	// FixturesDir is the directory where database or data fixtures are stored for loading during tests.
	FixturesDir string
	// EnvFilePath is the path to the environment configuration file used during tests.
	EnvFilePath string

	// Mocks contains the mock implementations for dependencies to be used during testing.
	Mocks *mocks.Mocks
	// DB is the storage interface used for interacting with the database during tests.
	DB storage.StorageInterface
	// MainOutputFunc is the primary output handler for the testing run.
	MainOutputFunc output.OutputInterface
	// Outputs is a collection of additional output handlers to process test results.
	Outputs []output.OutputInterface
	// Checkers is a list of custom checker interfaces for validating test results or conditions.
	Checkers []checker.CheckerInterface

	// CustomClient is a custom HTTP client used for making requests to the server during tests.
	CustomClient HTTPClient
	// HelperEndpoints is a map of helper endpoints available for facilitating tests.
	HelperEndpoints endpoint.EndpointMap
	// TemplateReplyFuncs contains a set of template functions for processing or customizing replies in tests.
	TemplateReplyFuncs template.FuncMap
}

// RunWithTesting is a helper function that wraps the common Run function and provides a simple way
// to configure Gonkex by populating the RunWithTestingOpts structure.
// t: The testing object used for managing test cases.
// serverURL: The URL of the server being tested.
// opts: The configuration options for the test run.
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

	handler := &testingHandler{t}
	runner := New(
		yamlLoader,
		&RunnerOpts{
			Host:  serverURL,
			Mocks: opts.Mocks,
			MocksLoader: mocks.NewYamlLoader(&mocks.YamlLoaderOpts{
				TemplateReplyFuncs: opts.TemplateReplyFuncs,
			}),
			FixturesDir:     opts.FixturesDir,
			DB:              opts.DB,
			Variables:       variables.New(),
			HTTPProxyURL:    proxyURL,
			HelperEndpoints: opts.HelperEndpoints,
			TestHandler:     handler.HandleTest,
		},
	)

	addOutputs(runner, opts)
	runner.AddCheckers(opts.Checkers...)

	err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}
}

func registerMocksEnvironment(m *mocks.Mocks) {
	names := m.GetNames()
	for _, n := range names {
		varName := fmt.Sprintf("GONKEX_MOCK_%s", strings.ToUpper(n))
		os.Setenv(varName, m.Service(n).ServerAddr())
	}
}

func addOutputs(runner *Runner, opts *RunWithTestingOpts) {
	if opts.MainOutputFunc != nil {
		runner.AddOutput(opts.MainOutputFunc)
	} else {
		runner.AddOutput(terminal.NewOutput(nil))
	}

	for _, o := range opts.Outputs {
		runner.AddOutput(o)
	}
}

type testingHandler struct {
	t *testing.T
}

func (h *testingHandler) HandleTest(test models.TestInterface, executor TestExecutor) error {
	var returnErr error
	h.t.Run(test.GetName(), func(t *testing.T) {
		result, err := executor(test)
		if err != nil {
			if errors.Is(err, checker.ErrTestSkipped) || errors.Is(err, checker.ErrTestBroken) {
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
