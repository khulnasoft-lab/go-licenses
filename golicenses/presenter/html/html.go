package html

// LicenseResult fields: Library, URL, Path, License, Type, Errs
// Example: Library (package name), License (license type)
import (
	"fmt"
	"io"

	"github.com/khulnasoft/go-licenses/golicenses"
)

type Presenter struct {
	results <-chan golicenses.LicenseResult
}

func NewPresenter(results <-chan golicenses.LicenseResult) *Presenter {
	return &Presenter{results: results}
}

func (p *Presenter) Present(w io.Writer) error {
	fmt.Fprintf(w, "<html><head><title>License Report</title></head><body><h1>License Report</h1><ul>")
	for res := range p.results {
		fmt.Fprintf(w, "<li><strong>%s</strong>: <code>%s</code></li>", res.Library, res.License)
	}
	fmt.Fprint(w, "</ul></body></html>")
	return nil
}
