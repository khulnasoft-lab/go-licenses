package spdx

import (
	"bytes"
	"regexp"
	"testing"

	"github.com/khulnasoft/go-licenses/golicenses"
	"github.com/stretchr/testify/assert"
)

func TestSPDXPresenter_Present(t *testing.T) {
	results := make(chan golicenses.LicenseResult)
	var outputBuffer bytes.Buffer

	p := NewPresenter(results)

	go func() {
		defer close(results)
		results <- golicenses.LicenseResult{
			Library: "github.com/owner/repo1",
			URL:     "https://github.com/owner/repo1",
			License: "MIT",
			Path:    "/path/to/repo1",
		}
		results <- golicenses.LicenseResult{
			Library: "gitlab.com/another/project2",
			URL:     "https://gitlab.com/another/project2.git",
			License: "Apache-2.0",
			Path:    "/path/to/project2",
		}
		results <- golicenses.LicenseResult{
			Library: "my-custom-lib@v1.2.3",
			URL:     "https://example.com/my-custom-lib", // Non-VCS URL
			License: "BSD-3-Clause-Invalid",              // To test NOASSERTION
			Path:    "/path/to/custom",
		}
	}()

	err := p.Present(&outputBuffer)
	assert.NoError(t, err, "Present should not return an error")

	output := outputBuffer.String()

	// Document Creation Information
	assert.Contains(t, output, "SPDXVersion: SPDX-2.3")
	assert.Contains(t, output, "DataLicense: CC0-1.0")
	assert.Contains(t, output, "SPDXID: SPDXRef-DOCUMENT")
	assert.Contains(t, output, "DocumentName: go-licenses-report")
	assert.Regexp(t, regexp.MustCompile(`DocumentNamespace: urn:uuid:[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`), output)
	assert.Contains(t, output, "Creator: Tool: go-licenses (github.com/khulnasoft/go-licenses)")
	assert.Regexp(t, regexp.MustCompile(`Created: \d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z`), output)

	// Package 1: github.com/owner/repo1
	assert.Contains(t, output, "PackageName: github.com/owner/repo1")
	assert.Contains(t, output, "SPDXID: SPDXRef-Package-github.com-owner-repo1")
	assert.Contains(t, output, "PackageDownloadLocation: git+https://github.com/owner/repo1")
	assert.Contains(t, output, "LicenseConcluded: MIT")
	assert.Contains(t, output, "LicenseDeclared: MIT")
	assert.Contains(t, output, "PackageLicenseComments: Source path: /path/to/repo1")

	// Package 2: gitlab.com/another/project2
	assert.Contains(t, output, "PackageName: gitlab.com/another/project2")
	assert.Contains(t, output, "SPDXID: SPDXRef-Package-gitlab.com-another-project2")
	assert.Contains(t, output, "PackageDownloadLocation: git+https://gitlab.com/another/project2.git")
	assert.Contains(t, output, "LicenseConcluded: Apache-2.0")
	assert.Contains(t, output, "LicenseDeclared: Apache-2.0")

	// Package 3: my-custom-lib@v1.2.3 (testing NOASSERTION for invalid license)
	assert.Contains(t, output, "PackageName: my-custom-lib@v1.2.3")
	assert.Contains(t, output, "SPDXID: SPDXRef-Package-my-custom-lib-at-v1.2.3")
	assert.Contains(t, output, "PackageDownloadLocation: https://example.com/my-custom-lib") // Non-VCS URL remains as is
	assert.Contains(t, output, "LicenseConcluded: NOASSERTION")
	assert.Contains(t, output, "LicenseDeclared: NOASSERTION")
}

func TestSanitizeSPDXID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "abc", "abc"},
		{"with slash", "github.com/foo/bar", "github.com-foo-bar"},
		{"with at", "package@version", "package-at-version"},
		{"with colon", "example:123", "example-col-123"},
		{"complex", "github.com/user/project@v1.2.3:sub", "github.com-user-project-at-v1.2.3-col-sub"},
		{"no changes", "valid-id.123", "valid-id.123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, sanitizeSPDXID(tt.input))
		})
	}
}

func TestIsValidSPDXLicenseID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{"valid MIT lower", "mit", true},
		{"valid MIT upper", "MIT", true},
		{"valid Apache-2.0", "Apache-2.0", true},
		{"valid GPL-2.0+", "GPL-2.0+", true},
		{"valid GPL-2.0-or-later lower", "gpl-2.0-or-later", true},
		{"invalid simple", "MyLicense", false},
		{"invalid with space", "Apache 2.0", false},
		{"invalid with AND", "MIT AND GPL-2.0", false},
		{"invalid with OR", "MIT OR Apache-2.0", false},
		{"valid bsd-2-clause", "bsd-2-clause", true},
		{"valid CC0-1.0", "CC0-1.0", true},
		{"invalid suffix", "GPL-2.0-only-foo", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, isValidSPDXLicenseID(tt.input))
		})
	}
}
