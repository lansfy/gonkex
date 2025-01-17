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

type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

type RunnerOpts struct {
	Host            string
	FixturesDir     string
	DB              storage.StorageInterface
	Mocks           *mocks.Mocks
	MocksLoader     mocks.Loader
	Variables       *variables.Variables
	CustomClient    HTTPClient
	HTTPProxyURL    *url.URL
	HelperEndpoints endpoint.EndpointMap
	TestHandler     func(models.TestInterface, TestExecutor) error
}

type TestExecutor func(models.TestInterface) (*models.Result, error)

type Runner struct {
	loader   testloader.LoaderInterface
	output   []output.OutputInterface
	checkers []checker.CheckerInterface
	config   RunnerOpts
}

func New(loader testloader.LoaderInterface, opts *RunnerOpts) *Runner {
	r := &Runner{
		loader: loader,
	}
	if opts != nil {
		r.config = *opts
	}

	if r.config.CustomClient == nil {
		r.config.CustomClient = newClient(r.config.HTTPProxyURL)
	}
	if r.config.TestHandler == nil {
		r.config.TestHandler = defaultTestHandler
	}

	r.AddCheckers(response_body.NewChecker())
	r.AddCheckers(response_header.NewChecker())
	if r.config.DB != nil {
		r.AddCheckers(response_db.NewChecker(r.config.DB))
	}
	return r
}

func (r *Runner) AddOutput(o ...output.OutputInterface) {
	r.output = append(r.output, o...)
}

func (r *Runner) AddCheckers(c ...checker.CheckerInterface) {
	r.checkers = append(r.checkers, c...)
}

func (r *Runner) Run() error {
	tests, err := r.loader.Load()
	if err != nil {
		return err
	}

	hasFocused := checkHasFocused(tests)
	for _, t := range tests {
		// make a copy because go test runner runs tests in separate goroutines
		// and without copy tests will override each other
		test := t
		if hasFocused {
			switch test.GetStatus() {
			case "focus":
				test.SetStatus("")
			case "broken":
				// do nothing
			default:
				test.SetStatus("skipped")
			}
		}

		err := r.config.TestHandler(test, r.executeTestWithRetryPolicy)
		if err != nil {
			return colorize.NewEntityError("test %s error", test.GetName()).SetSubError(err)
		}
	}

	return nil
}

var (
	errTestSkipped = errors.New("test was skipped")
	errTestBroken  = errors.New("test was broken")
)

func (r *Runner) executeTestWithRetryPolicy(v models.TestInterface) (*models.Result, error) {
	var testResult *models.Result
	var err error

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

	successCount := 0
	for i := 0; i < retryCount+1; i++ {
		if i != 0 {
			time.Sleep(retryPolicy.Delay())
		}
		testResult, err = r.executeTest(v)
		if err != nil {
			return nil, err
		}
		if !testResult.Passed() {
			successCount = 0
		} else {
			successCount++
		}
		if successCount >= successRequired {
			break
		}
	}

	if testResult.Passed() && successCount < successRequired {
		testResult.Errors = append(testResult.Errors,
			fmt.Errorf("last run was successful %d times, but %d success at row required", successCount, successRequired),
		)
	}

	for _, o := range r.output {
		if err = o.Process(v, testResult); err != nil {
			return nil, err
		}
	}

	return testResult, nil
}

func (r *Runner) executeTest(v models.TestInterface) (*models.Result, error) {
	if v.GetStatus() != "" {
		if v.GetStatus() == "broken" {
			return &models.Result{Test: v}, errTestBroken
		}

		if v.GetStatus() == "skipped" {
			return &models.Result{Test: v}, errTestSkipped
		}
	}

	r.config.Variables.Load(v.GetCombinedVariables())
	v = r.config.Variables.Apply(v)

	if r.config.DB != nil && len(v.Fixtures()) != 0 {
		err := r.config.DB.LoadFixtures(r.config.FixturesDir, v.Fixtures())
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
		if err := r.config.Mocks.LoadDefinitions(r.config.MocksLoader, v.ServiceMocks()); err != nil {
			return nil, err
		}
	}

	// launch script in cmd interface
	if v.BeforeScriptPath() != "" {
		if err := cmd_runner.CmdRun(v.BeforeScriptPath(), v.BeforeScriptTimeout()); err != nil {
			return nil, err
		}
	}

	// make pause
	pause := v.Pause()
	if pause > 0 {
		fmt.Printf("Sleep %s before requests\n", pause)
		time.Sleep(pause)
	}

	req, err := NewRequest(r.config.Host, v)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	if strings.HasPrefix(req.URL.Path, endpoint.Prefix) {
		resp, err = endpoint.SelectEndpoint(r.config.Mocks, r.config.HelperEndpoints, req.URL.Path, req) //nolint:bodyclose // false positive
	} else {
		resp, err = r.config.CustomClient.Do(req) //nolint:bodyclose // false positive
	}
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)

	_ = resp.Body.Close()

	if err != nil {
		return nil, err
	}

	bodyStr := string(body)

	result := models.Result{
		Path:                req.URL.Path,
		Query:               req.URL.RawQuery,
		RequestBody:         actualRequestBody(req),
		ResponseBody:        bodyStr,
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

	// launch script in cmd interface
	if v.AfterRequestScriptPath() != "" {
		if err := cmd_runner.CmdRun(v.AfterRequestScriptPath(), v.AfterRequestScriptTimeout()); err != nil {
			return nil, err
		}
	}

	if r.config.Mocks != nil {
		errs := r.config.Mocks.EndRunningContext()
		result.Errors = append(result.Errors, errs...)
	}

	if err := r.setVariablesFromResponse(v, result.ResponseContentType, bodyStr, resp.StatusCode); err != nil {
		return nil, err
	}

	r.config.Variables.Load(v.GetCombinedVariables())
	v = r.config.Variables.Apply(v)

	for _, c := range r.checkers {
		errs, err := c.Check(v, &result)
		if err != nil {
			return nil, err
		}
		result.Errors = append(result.Errors, errs...)
	}

	return &result, nil
}

func (r *Runner) setVariablesFromResponse(t models.TestInterface, contentType, body string, statusCode int) error {
	varTemplates := t.GetVariablesToSet()
	if varTemplates == nil {
		return nil
	}

	isJSON := strings.Contains(contentType, "json") && body != ""

	vars, err := variables.FromResponse(varTemplates[statusCode], body, isJSON)
	if err != nil {
		return err
	}

	if vars != nil {
		r.config.Variables.Merge(vars)
	}

	return nil
}

func checkHasFocused(tests []models.TestInterface) bool {
	for _, test := range tests {
		if test.GetStatus() == "focus" {
			return true
		}
	}

	return false
}

func defaultTestHandler(t models.TestInterface, f TestExecutor) error {
	result, err := f(t)
	if err != nil {
		return err
	}
	if len(result.Errors) != 0 {
		return result.Errors[0]
	}
	return nil
}
