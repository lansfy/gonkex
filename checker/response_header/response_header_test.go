package response_header

import (
	"errors"
	"sort"
	"testing"

	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/testloader/yaml_file"

	"github.com/fatih/color"
	"github.com/stretchr/testify/assert"
)

func TestCheckShouldMatchSubset(t *testing.T) {
	color.NoColor = true
	test := &yaml_file.Test{
		ResponseHeaders: map[int]map[string]string{
			200: {
				"content-type": "application/json",
				"ACCEPT":       "text/html",
			},
		},
	}

	result := &models.Result{
		ResponseStatusCode: 200,
		ResponseHeaders: map[string][]string{
			"Content-Type": {
				"application/json",
			},
			"Accept": {
				// uts enough for expected value to match only one entry of the actual values slice
				"application/json",
				"text/html",
			},
		},
	}

	checker := NewChecker()
	errs, err := checker.Check(test, result)

	assert.NoError(t, err, "Check must not result with an error")
	assert.Empty(t, errs, "Check must succeed")
}

func TestCheckWhenNotMatchedShouldReturnError(t *testing.T) {
	color.NoColor = true
	test := &yaml_file.Test{
		ResponseHeaders: map[int]map[string]string{
			200: {
				"content-type": "application/json",
				"accept":       "text/html",
			},
		},
	}

	result := &models.Result{
		ResponseStatusCode: 200,
		ResponseHeaders: map[string][]string{
			// no header "Content-Type" in response
			"Accept": {
				"application/json",
			},
		},
	}

	checker := NewChecker()
	errs, err := checker.Check(test, result)

	sort.Slice(errs, func(i, j int) bool {
		return errs[i].Error() < errs[j].Error()
	})

	assert.NoError(t, err, "Check must not result with an error")
	assert.Equal(
		t,
		errs,
		[]error{
			errors.New("response does not include expected header Content-Type"),
			errors.New("response header Accept value does not match:\n     expected: text/html\n       actual: application/json"),
		},
	)
}
