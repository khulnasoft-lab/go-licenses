package markdown

import (
	"bytes"
	"testing"

	"github.com/khulnasoft/go-licenses/golicenses"
	"github.com/stretchr/testify/assert"
)

func TestMarkdownPresenter_Present(t *testing.T) {
	results := make(chan golicenses.LicenseResult)
	var outputBuffer bytes.Buffer

	p := NewPresenter(results)

	go func() {
		defer close(results)
		results <- golicenses.LicenseResult{
			Library: "library1",
			URL:     "http://example.com/library1",
			Path:    "/path/to/library1",
			License: "MIT",
			Type:    "Permissive",
		}
		results <- golicenses.LicenseResult{
			Library: "library2",
			URL:     "http://example.com/library2",
			Path:    "/path/to/library2",
			License: "Apache-2.0",
			Type:    "Permissive",
			Errs:    nil, // Explicitly nil for clarity
		}
	}()

	err := p.Present(&outputBuffer)
	assert.NoError(t, err, "Present should not return an error")

	expectedOutput := "# License Report\n\n" +
		"- **library1**: `MIT`\n" +
		"- **library2**: `Apache-2.0`\n"

	assert.Equal(t, expectedOutput, outputBuffer.String(), "Output should match expected Markdown format")
}
