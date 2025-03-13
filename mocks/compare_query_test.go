package mocks

import (
	"testing"
)

func TestCompareQuery(t *testing.T) {
	tests := []struct {
		name     string
		expected []string
		actual   []string
	}{
		{
			name:     "simple expected and actual",
			expected: []string{"cake"},
			actual:   []string{"cake"},
		},
		{
			name:     "expected and actual with two values",
			expected: []string{"cake", "tea"},
			actual:   []string{"cake", "tea"},
		},
		{
			name:     "expected and actual with two values and different order",
			expected: []string{"cake", "tea"},
			actual:   []string{"tea", "cake"},
		},
		{
			name:     "expected and actual with same values",
			expected: []string{"tea", "cake", "tea"},
			actual:   []string{"cake", "tea", "tea"},
		},
		{
			name:     "expected and actual with regexp",
			expected: []string{"tea", "$matchRegexp(^c\\w+)"},
			actual:   []string{"cake", "tea"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ok, err := compareQuery(tt.expected, tt.actual)
			if err != nil {
				t.Error(err)
			}
			if !ok {
				t.Errorf("expected and actual queries do not match")
			}
		})
	}
}
