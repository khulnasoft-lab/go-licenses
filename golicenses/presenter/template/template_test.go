package template

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/khulnasoft/go-licenses/golicenses"
	"github.com/stretchr/testify/assert"
)

func TestTemplatePresenter_Present(t *testing.T) {
	results := make(chan golicenses.LicenseResult)
	var outputBuffer bytes.Buffer

	templateFile, err := filepath.Abs("./testdata/test_template.tmpl")
	assert.NoError(t, err, "Should be able to get absolute path for template file")

	p, err := NewPresenter(results, templateFile)
	assert.NoError(t, err, "NewPresenter should not return an error with a valid template file")
	assert.NotNil(t, p, "Presenter should not be nil")

	go func() {
		defer close(results)
		results <- golicenses.LicenseResult{
			Library: "tmpl_lib1",
			License: "MIT-T",
		}
		results <- golicenses.LicenseResult{
			Library: "tmpl_lib2",
			License: "Apache-2.0-T",
		}
	}()

	err = p.Present(&outputBuffer)
	assert.NoError(t, err, "Present should not return an error")

	expectedOutput := "Library: tmpl_lib1, License: MIT-T\n" +
		"Library: tmpl_lib2, License: Apache-2.0-T\n"

	assert.Equal(t, expectedOutput, outputBuffer.String(), "Output should match expected format from template")
}

func TestTemplatePresenter_NewPresenter_FileNotFound(t *testing.T) {
	results := make(chan golicenses.LicenseResult)
	// Close the channel immediately as it won't be used for this error case
	close(results)

	_, err := NewPresenter(results, "./testdata/non_existent_template.tmpl")
	assert.Error(t, err, "NewPresenter should return an error if template file is not found")
}

func TestTemplatePresenter_NewPresenter_InvalidTemplate(t *testing.T) {
	results := make(chan golicenses.LicenseResult)
	// Close the channel immediately
	close(results)

	// Create a temporary invalid template file
	invalidTemplateFile, err := filepath.Abs("./testdata/invalid_template.tmpl")
	assert.NoError(t, err)

	// Write some invalid template content
	// For example, an unclosed action
	content := []byte("{{ if .Library }")
	f, _ := os.Create(invalidTemplateFile)
	_, _ = f.Write(content)
	_ = f.Close()
	//defer os.Remove(invalidTemplateFile) // Clean up

	_, err = NewPresenter(results, invalidTemplateFile)
	assert.Error(t, err, "NewPresenter should return an error for an invalid template")

	// Clean up the invalid template file manually if defer is not used or test panics
	_ = os.Remove(invalidTemplateFile)
}
