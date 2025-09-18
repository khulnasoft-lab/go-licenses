package markdown

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
	fmt.Fprintf(w, "# License Report\n\n")
	for res := range p.results {
		fmt.Fprintf(w, "- **%s**: `%s`\n", res.Library, res.License)
	}
	return nil
}
