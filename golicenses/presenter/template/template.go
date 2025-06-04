package template

import (
	"io"
	"text/template"

	"github.com/khulnasoft/go-licenses/golicenses"
)

// LicenseResult fields available in templates: Library, URL, Path, License, Type, Errs
// Example: {{ .Library }} {{ .License }}
type Presenter struct {
	results <-chan golicenses.LicenseResult
	tmpl    *template.Template
}

func NewPresenter(results <-chan golicenses.LicenseResult, templatePath string) (*Presenter, error) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return nil, err
	}
	return &Presenter{results: results, tmpl: tmpl}, nil
}

func (p *Presenter) Present(w io.Writer) error {
	var all []golicenses.LicenseResult
	for res := range p.results {
		all = append(all, res)
	}
	return p.tmpl.Execute(w, all)
}
