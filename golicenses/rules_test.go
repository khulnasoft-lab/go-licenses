package golicenses

import (
	"github.com/go-test/deep"
	"testing"
)

// TestRules_Evaluate tests the Rules evaluation logic with various allow/deny patterns and edge cases.
func TestRules_Evaluate(t *testing.T) {
	tests := []struct {
		name       string
		act        Action
		patterns   []string
		against    []LicenseResult
		ignore     []string
		expected   bool
		failedHits []LicenseResult
	}{
		{
			name:     "go case",
			act:      AllowAction,
			patterns: []string{"MIT-0"},
			against: []LicenseResult{
				{
					Library: "lib1",
					License: "MIT-0",
				},
			},
			expected: true,
		},
		{
			name:     "multiple allow patterns",
			act:      AllowAction,
			patterns: []string{"MIT-0", "BSD.*"},
			against: []LicenseResult{
				{
					Library: "lib1",
					License: "MIT-0",
				},
				{
					Library: "lib2",
					License: "BSD",
				},
				{
					Library: "lib3",
					License: "WTFPL",
				},
			},
			ignore:   []string{"lib3"},
			expected: true,
		},
		{
			name:     "allow fails eval",
			act:      AllowAction,
			patterns: []string{"MIT.*"},
			against: []LicenseResult{
				{
					Library: "lib1",
					License: "MIT-0",
				},
				{
					Library: "lib2",
					License: "BSD",
				},
			},
			expected: false,
			failedHits: []LicenseResult{
				{
					Library: "lib2",
					License: "BSD",
				},
			},
		},
		{
			name:     "deny fails eval",
			act:      DenyAction,
			patterns: []string{"MIT.*"},
			against: []LicenseResult{
				{
					Library: "lib1",
					License: "MIT-0",
				},
				{
					Library: "lib2",
					License: "BSD",
				},
			},
			expected: false,
			failedHits: []LicenseResult{
				{
					Library: "lib1",
					License: "MIT-0",
				},
			},
		},
		{
			name:     "allow ignore",
			act:      AllowAction,
			patterns: []string{"MIT.*"},
			against: []LicenseResult{
				{
					Library: "lib1",
					License: "MIT-0",
				},
				{
					Library: "lib2",
					License: "BSD",
				},
			},
			expected:   true,
			ignore:     []string{"lib2"},
			failedHits: []LicenseResult{},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			r, err := NewRules(test.act, test.patterns, test.ignore...)
			if err != nil {
				t.Fatalf("failed to make rules: %+v", err)
			}

			actual, failedHits, err := r.Evaluate(test.against...)
			if actual != test.expected {
				t.Errorf("bad evaluation: %v", actual)
			}

			if len(failedHits) != len(test.failedHits) {
				t.Fatalf("bad hint count: %d", len(failedHits))
			}

			for idx, h := range failedHits {
				expected := test.failedHits[idx]
				diffs := deep.Equal(expected, h)
				if len(diffs) > 0 {
					for _, d := range diffs {
						t.Errorf("diff: %+v", d)
					}
				}
			}
		})
		// Additional edge cases
		edgeCases := []struct {
			name     string
			act      Action
			patterns []string
			against  []LicenseResult
			ignore   []string
			expected bool
		}{
			{
				name:     "empty patterns allow",
				act:      AllowAction,
				patterns: []string{},
				against:  []LicenseResult{{Library: "lib1", License: "MIT-0"}},
				expected: false,
			},
			{
				name:     "empty patterns deny",
				act:      DenyAction,
				patterns: []string{},
				against:  []LicenseResult{{Library: "lib1", License: "MIT-0"}},
				expected: true,
			},
			{
				name:     "empty input",
				act:      AllowAction,
				patterns: []string{"MIT.*"},
				against:  []LicenseResult{},
				expected: true,
			},
			{
				name:     "deny all",
				act:      DenyAction,
				patterns: []string{".*"},
				against:  []LicenseResult{{Library: "lib1", License: "MIT-0"}},
				expected: false,
			},
			{
				name:     "allow all",
				act:      AllowAction,
				patterns: []string{".*"},
				against:  []LicenseResult{{Library: "lib1", License: "MIT-0"}},
				expected: true,
			},
			{
				name:     "ignore all",
				act:      AllowAction,
				patterns: []string{"MIT.*"},
				against:  []LicenseResult{{Library: "lib1", License: "BSD"}},
				ignore:   []string{"lib1"},
				expected: true,
			},
		}
		for _, ec := range edgeCases {
			t.Run(ec.name, func(t *testing.T) {
				r, err := NewRules(ec.act, ec.patterns, ec.ignore...)
				if err != nil {
					t.Fatalf("failed to make rules: %+v", err)
				}
				actual, _, err := r.Evaluate(ec.against...)
				if actual != ec.expected {
					t.Errorf("bad evaluation for %s: got %v, want %v", ec.name, actual, ec.expected)
				}
			})
		}
	}
}
