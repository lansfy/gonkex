package runner

import (
	"flag"
)

const filterFlagName = "gonkex-filter"

var filterFlag string

// RegisterFlags registers command-line flags for the Gonkex testing framework.
// Now it registers a "gonkex-filter" flag that allows users to filter which test files
// are executed during a test run.
//
// Usage: in test file add next code
//
//	func init() {
//	    runner.RegisterFlags()
//	}
//
//	// Filter will be automatically applied if -gonkex-filter flag is provided
//
// Command line examples:
//
//	go test -gonkex-filter="mytest.yaml"    // Run only tests in file mytest.yaml
//	go test -gonkex-filter=mytest           // Run all files which has "mytest" in path or name
//
// The filter value is stored in the package-level filterFlag variable and applied
// to the test loader when non-empty, allowing users to run only a subset of tests
// matching the specified criteria.
func RegisterFlags() {
	if flag.Lookup(filterFlagName) == nil {
		flag.StringVar(&filterFlag, filterFlagName, "", "if non-empty, gonkex will use this string as filter.")
	}
}
