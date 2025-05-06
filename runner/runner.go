package runner

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/lansfy/gonkex/checker"
	"github.com/lansfy/gonkex/checker/response_body"
	"github.com/lansfy/gonkex/checker/response_db"
	"github.com/lansfy/gonkex/checker/response_header"
	"github.com/lansfy/gonkex/cmd_runner"
	"github.com/lansfy/gonkex/colorize"
	"github.com/lansfy/gonkex/endpoint"
	"github.com/lansfy/gonkex/mocks"
	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/output"
	"github.com/lansfy/gonkex/storage"
	"github.com/lansfy/gonkex/testloader"
	"github.com/lansfy/gonkex/variables"
)

// OnFailPolicy defines the policy to follow when a test fails.
type OnFailPolicy string

const (
	// PolicySkipFile skips the current test file if test fails (default policy).
	PolicySkipFile OnFailPolicy = "file"
	// PolicyStop stops all test execution on failure.
	PolicyStop OnFailPolicy = "stop"
	// PolicyContinue continues running tests despite failures.
	PolicyContinue OnFailPolicy = "continue"
)

// HTTPClient defines an interface for making HTTP requests.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// RunnerOpts holds configuration options for the test runner.
type RunnerOpts struct {
	Host            string                                         // The base URL of the server being tested.
	FixturesDir     string                                         // Directory containing fixture files.
	DB              storage.StorageInterface                       // Database interface for test storage.
	Mocks           *mocks.Mocks                                   // Mock implementations for external dependencies.
	MocksLoader     mocks.Loader                                   // Loader for mock configurations.
	Variables       *variables.Variables                           // Variables used in test execution.
	CustomClient    HTTPClient                                     // Custom HTTP client for making requests.
	HTTPProxyURL    *url.URL                                       // Proxy URL for HTTP requests.
	HelperEndpoints endpoint.EndpointMap                           // Map of helper endpoints for testing.
	TestHandler     func(models.TestInterface, TestExecutor) error // Handler function for executing tests.
	OnFailPolicy    OnFailPolicy                                   // Policy defining what happens when a step fails.
}

type TestExecutor func(models.TestInterface) (*models.Result, error)

// Runner orchestrates test execution, output handling, and validation.
type Runner struct {
	loader   testloader.LoaderInterface
	output   []output.OutputInterface
	checkers checkersList
	config   RunnerOpts
}

// New creates a new test runner with the given loader and options.
func New(loader testloader.LoaderInterface, opts *RunnerOpts) *Runner {
	r := &Runner{
		loader: loader,
	}
	if opts != nil {
		r.config = *opts
	}
	if r.config.OnFailPolicy == "" {
		r.config.OnFailPolicy = PolicySkipFile
	}

	if r.config.CustomClient == nil {
		r.config.CustomClient = newClient(r.config.HTTPProxyURL)
	}
	if r.config.TestHandler == nil {
		r.config.TestHandler = r.defaultTestHandler
	}

	r.AddCheckers(response_body.NewChecker())
	r.AddCheckers(response_header.NewChecker())
	if r.config.DB != nil {
		r.AddCheckers(response_db.NewChecker(r.config.DB))
	}
	return r
}

// AddOutput adds one or more output handlers to the runner.
func (r *Runner) AddOutput(o ...output.OutputInterface) {
	r.output = append(r.output, o...)
}

// AddCheckers adds one or more checkers to validate test results.
func (r *Runner) AddCheckers(c ...checker.CheckerInterface) {
	r.checkers.AddCheckers(c...)
}

// Run executes the test suite, processing each test and handling failures according to policy.
func (r *Runner) Run() error {
	tests, err := r.loader.Load()
	if err != nil {
		return err
	}

	hasFocused := checkHasFocused(tests)

	var wasError bool
	var errs []error
	for _, t := range tests {
		if t.FirstTestInFile() {
			wasError = false
		}

		switch t.GetStatus() {
		case models.StatusFocus:
			t.SetStatus(models.StatusNone)
		case models.StatusBroken:
			// do nothing
		default:
			if hasFocused || wasError {
				t.SetStatus(models.StatusSkipped)
			}
		}

		err := r.config.TestHandler(t, r.executeTest)
		if err != nil {
			err = colorize.NewEntityError("test %s error", t.GetName()).SetSubError(err)
			if hasFocused || r.config.OnFailPolicy == PolicyStop {
				return err
			}
			errs = append(errs, err)
			if r.config.OnFailPolicy == PolicySkipFile {
				wasError = true
			}
		}
	}

	switch len(errs) {
	case 0:
		return nil
	case 1:
		return errs[0]
	default:
		return errors.New("some steps failed")
	}
}

func makeServiceRequest(config *RunnerOpts, v models.TestInterface) (*models.Result, error) {
	req, err := NewRequest(config.Host, v)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	if strings.HasPrefix(req.URL.Path, endpoint.Prefix) {
		resp, err = endpoint.SelectEndpoint(config.Mocks, config.HelperEndpoints, req.URL.Path, req) //nolint:bodyclose // false positive
	} else {
		resp, err = config.CustomClient.Do(req) //nolint:bodyclose // false positive
	}
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)

	_ = resp.Body.Close()

	if err != nil {
		return nil, err
	}

	result := &models.Result{
		Path:                req.URL.Path,
		Query:               req.URL.RawQuery,
		RequestBody:         actualRequestBody(req),
		ResponseBody:        string(body),
		ResponseContentType: resp.Header.Get("Content-Type"),
		ResponseStatusCode:  resp.StatusCode,
		ResponseStatus:      resp.Status,
		ResponseHeaders:     resp.Header,
		Test:                v,
	}

	// support for Trailer headers: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Trailer
	for name, value := range resp.Trailer {
		result.ResponseHeaders[name] = value
	}

	return result, nil
}

func (r *Runner) executeTest(v models.TestInterface) (*models.Result, error) {
	retryPolicy := v.GetRetryPolicy()

	retryCount := retryPolicy.Attempts()
	if retryCount < 0 {
		return nil, colorize.NewEntityError("section %s: attempts count must be non-negative", "retryPolicy")
	}

	successRequired := retryPolicy.SuccessCount()
	if successRequired < 0 {
		return nil, colorize.NewEntityError("section %s: 'successInRow' count must be positive", "retryPolicy")
	}
	if successRequired == 0 {
		successRequired = 1
	}

	for _, o := range r.output {
		if exo, ok := o.(output.ExtendedOutputInterface); ok {
			if err := exo.BeforeTest(v); err != nil {
				return nil, err
			}
		}
	}

	switch v.GetStatus() {
	case models.StatusBroken:
		return nil, checker.ErrTestBroken
	case models.StatusSkipped:
		return nil, checker.ErrTestSkipped
	}

	err := r.checkers.BeforeTest(v)
	if err != nil {
		return nil, err
	}

	r.config.Variables.Merge(v.GetCombinedVariables())

	v = v.Clone()
	if r.config.Variables != nil {
		v.ApplyVariables(r.config.Variables.Substitute)
	}

	if r.config.DB != nil && len(v.Fixtures()) != 0 {
		err = r.config.DB.LoadFixtures(r.config.FixturesDir, v.Fixtures())
		if err != nil {
			return nil, fmt.Errorf("load fixtures %v: %w", v.Fixtures(), err)
		}
	}

	// reset mocks
	if r.config.Mocks != nil {
		// prevent deriving the definition from previous test
		r.config.Mocks.ResetDefinitions()
		r.config.Mocks.ResetRunningContext()
	}

	// load mocks
	if v.ServiceMocks() != nil {
		err = r.config.Mocks.LoadDefinitions(r.config.MocksLoader, v.ServiceMocks())
		if err != nil {
			return nil, err
		}
	}

	// launch script in cmd interface
	if v.BeforeScriptPath() != "" {
		err = cmd_runner.CmdRun(v.BeforeScriptPath(), v.BeforeScriptTimeout())
		if err != nil {
			return nil, err
		}
	}

	// make pause
	pause := v.Pause()
	if pause > 0 {
		_, _ = fmt.Printf("Sleep %s before requests\n", pause)
		time.Sleep(pause)
	}

	retryCheckers := checkersList{}
	if retryCount != 0 {
		retryCheckers.AddCheckers(response_body.NewChecker(), response_header.NewChecker())
	}

	var errs []error
	var result *models.Result
	var successCount int
	for i := 0; i < retryCount+1; i++ {
		if i != 0 {
			time.Sleep(retryPolicy.Delay())
		}

		result, err = makeServiceRequest(&r.config, v)
		if err != nil {
			return nil, err
		}

		errs, err = retryCheckers.Check(v, result)
		if err != nil {
			return nil, err
		}

		if len(errs) != 0 {
			successCount = 0
		} else {
			successCount++
		}
		if successCount >= successRequired {
			break
		}
	}

	if len(errs) == 0 && successCount < successRequired {
		result.Errors = append(result.Errors,
			fmt.Errorf("last run was successful %d times, but %d success at row required", successCount, successRequired),
		)
	}

	// make pause after request
	pause = v.AfterRequestPause()
	if pause > 0 {
		time.Sleep(pause)
	}

	// launch script in cmd interface
	if v.AfterRequestScriptPath() != "" {
		err = cmd_runner.CmdRun(v.AfterRequestScriptPath(), v.AfterRequestScriptTimeout())
		if err != nil {
			return nil, err
		}
	}

	if r.config.Mocks != nil {
		errs := r.config.Mocks.EndRunningContext()
		result.Errors = append(result.Errors, errs...)
	}

	skipCheckers := false
	changed, errs := r.setVariablesFromResponse(v, result)
	if len(errs) != 0 {
		// we should show response in output, so better to add this error as result error
		// and skip all checkers
		result.Errors = append(result.Errors, errs...)
		skipCheckers = true
	}

	if changed {
		// if new variable assigned we will apply them to model
		v.ApplyVariables(r.config.Variables.Substitute)
	}

	if !skipCheckers {
		errs, err = r.checkers.Check(v, result)
		if err != nil {
			return nil, err
		}

		result.Errors = append(result.Errors, errs...)
	}

	for _, o := range r.output {
		err = o.Process(v, result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (r *Runner) setVariablesFromResponse(t models.TestInterface, result *models.Result) (bool, []error) {
	varTemplates, ok := t.GetVariablesToSet(result.ResponseStatusCode)
	if !ok || len(varTemplates) == 0 {
		return false, nil
	}

	vars, errs := response_body.ExtractValues(varTemplates, result)
	if len(errs) != 0 || len(vars) == 0 {
		for idx := range errs {
			errs[idx] = colorize.NewEntityError("section %s", "variables_to_set").SetSubError(errs[idx])
		}
		return false, errs
	}

	r.config.Variables.Merge(vars)
	return true, nil
}

func (r *Runner) defaultTestHandler(t models.TestInterface, f TestExecutor) error {
	result, err := f(t)
	if err != nil {
		if isTestWasSkipped(err) {
			return nil
		}
		return err
	}
	if len(result.Errors) != 0 {
		if len(r.output) != 0 {
			return errors.New("failed")
		}
		return result.Errors[0]
	}
	return nil
}

// checkHasFocused checks if any test has a "focus" status, indicating prioritized execution.
func checkHasFocused(tests []models.TestInterface) bool {
	for _, test := range tests {
		if test.GetStatus() == models.StatusFocus {
			return true
		}
	}

	return false
}

func isTestWasSkipped(err error) bool {
	return err != nil && (errors.Is(err, checker.ErrTestSkipped) || errors.Is(err, checker.ErrTestBroken))
}
