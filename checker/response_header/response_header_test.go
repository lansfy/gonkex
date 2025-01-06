package response_header

import (
	"testing"

	"github.com/lansfy/gonkex/models"
	"github.com/lansfy/gonkex/testloader/yaml_file"

	"github.com/stretchr/testify/require"
)

func TestCheckShouldMatchSubset(t *testing.T) {
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

	require.NoError(t, err, "Check must not result with an error")
	require.Empty(t, errs, "Check must succeed")
}

func TestCheckWhenNotMatchedShouldReturnError(t *testing.T) {
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
	require.NoError(t, err, "Check must not result with an error")

	errText := []string{}
	for _, err = range errs {
		errText = append(errText, err.Error())
	}

	require.ElementsMatch(
		t,
		errText,
		[]string{
			"response does not include expected header 'Content-Type'",
			"response header 'Accept' value does not match:\n     expected: text/html\n       actual: application/json",
		},
	)
}
