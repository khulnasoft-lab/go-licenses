package html

import (
	"bytes"
	"testing"

	"github.com/khulnasoft/go-licenses/golicenses"
	"github.com/stretchr/testify/assert"
)

func TestHTMLPresenter_Present(t *testing.T) {
	results := make(chan golicenses.LicenseResult)
	var outputBuffer bytes.Buffer

	p := NewPresenter(results)

	go func() {
		defer close(results)
		results <- golicenses.LicenseResult{
			Library: "lib1",
			License: "MIT",
		}
		results <- golicenses.LicenseResult{
			Library: "lib2",
			License: "Apache-2.0",
		}
	}()

	err := p.Present(&outputBuffer)
	assert.NoError(t, err, "Present should not return an error")

	expectedOutput := "<html><head><title>License Report</title></head><body><h1>License Report</h1><ul>" +
		"<li><strong>lib1</strong>: <code>MIT</code></li>" +
		"<li><strong>lib2</strong>: <code>Apache-2.0</code></li>" +
		"</ul></body></html>"

	assert.Equal(t, expectedOutput, outputBuffer.String(), "Output should match expected HTML format")
}
