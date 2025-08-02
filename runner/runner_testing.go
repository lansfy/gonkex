package runner

import (
	"errors"
	"net/url"
	"os"
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
)

// MockEnvironmentPrefix is the default prefix used by gonkex for registering mock service environment variables.
const MockEnvironmentPrefix = "GONKEX_MOCK_"

type RunWithTestingOpts struct {
	// TestsDir is the directory where test definitions or files are located.
	TestsDir string
	// FixturesDir is the directory where database or data fixtures are stored for loading during tests.
	FixturesDir string
	// EnvFilePath is the path to the environment configuration file used during tests.
	EnvFilePath string
	// Variables holds test execution variables that can be referenced
	// and modified during test runs for dynamic test behavior.
	Variables variables.Variables

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
	// EndpointsPrefix consists common prefix for all HelperEndpoints.
	EndpointsPrefix string
	// TemplateReplyFuncs contains a set of template functions for processing or customizing replies in tests.
	TemplateReplyFuncs template.FuncMap
	// OnFailPolicy defining what happens when a some step of test fails.
	OnFailPolicy OnFailPolicy
}

// RunWithTesting is a helper function that wraps the common Run function and provides a simple way
// to configure Gonkex by populating the RunWithTestingOpts structure.
// t: The testing object used for managing test cases.
// serverURL: The URL of the server being tested.
// opts: The configuration options for the test run.
func RunWithTesting(t *testing.T, serverURL string, opts *RunWithTestingOpts) {
	if opts.Mocks != nil {
		if err := opts.Mocks.RegisterEnvironmentVariables(MockEnvironmentPrefix); err != nil {
			t.Fatal(err)
		}
	}

	if opts.EnvFilePath != "" {
		if err := RegisterEnvironmentVariables(opts.EnvFilePath, false); err != nil {
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
	if os.Getenv("GONKEX_FILE_FILTER") != "" {
		setStringFilter(yamlLoader, os.Getenv("GONKEX_FILE_FILTER"))
	}

	if allureDirFlag == "" {
		allureDirFlag = os.Getenv("GONKEX_ALLURE_DIR")
	}

	handler := &testingHandler{
		t: t,
	}
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
			Variables:       opts.Variables,
			HTTPProxyURL:    proxyURL,
			EndpointsPrefix: opts.EndpointsPrefix,
			HelperEndpoints: opts.HelperEndpoints,
			TestHandler:     handler.HandleTest,
			OnFailPolicy:    opts.OnFailPolicy,
		},
	)

	addOutputs(runner, opts)
	runner.AddCheckers(opts.Checkers...)

	err := runner.Run()
	if err != nil {
		t.Fatal(err)
	}
	if !handler.executed {
		t.Skip("no tests to run: none found or all were filtered out")
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
	t        *testing.T
	executed bool
}

func (h *testingHandler) HandleTest(test models.TestInterface, executor TestExecutor) (bool, error) {
	h.executed = true
	var critical bool
	var returnErr error
	h.t.Run(test.GetName(), func(t *testing.T) {
		result, err := executor(test)
		if err != nil {
			if isTestWasSkipped(err) {
				t.Skip()
			} else {
				returnErr = err
				critical = true
				t.Fatal(err)
			}
		}

		if !result.Passed() {
			returnErr = errors.New("failed")
			t.Fail()
		}
	})

	return critical, returnErr
}
