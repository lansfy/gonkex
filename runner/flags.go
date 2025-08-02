package runner

import (
	"flag"
)

const filterFlagName = "gonkex-filter"
const allureDirFlagName = "gonkex-allure-dir"

var filterFlag string
var allureDirFlag string

// RegisterFlags registers command-line flags for the Gonkex testing framework.
// Now it registers:
// * "gonkex-filter" flag that allows users to filter which test files are executed during a test run.
// * "gonkex-allure-dir" flag which enable allure report and set folder for execurion's result.
//
// Usage: in test file add next code
//
//	func init() {
//	    runner.RegisterFlags()
//	}
//
// Option will be automatically applied if some flag is provided.
//
// Command line examples:
//
//	go test -gonkex-filter=mytest.yaml      // Run only tests in file mytest.yaml
//	go test -gonkex-filter=mytest           // Run all files which has "mytest" in path or name
//	go test -gonkex-allure-dir=testresult   // Generate allure report after tests in "testresult" folder
//
// The filter value is stored in the package-level filterFlag variable and applied
// to the test loader when non-empty, allowing users to run only a subset of tests
// matching the specified criteria.
func RegisterFlags() {
	if flag.Lookup(filterFlagName) == nil {
		flag.StringVar(&filterFlag, filterFlagName, "", "if non-empty, gonkex will use this string as filter.")
	}
	if flag.Lookup(allureDirFlagName) == nil {
		flag.StringVar(&allureDirFlag, allureDirFlagName, "", "if non-empty, gonkex will create allure report in specified folder.")
	}
}
