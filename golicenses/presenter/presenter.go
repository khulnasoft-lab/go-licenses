package presenter

import (
	"io"

	"github.com/khulnasoft/go-licenses/golicenses"
	"github.com/khulnasoft/go-licenses/golicenses/presenter/csv"
	"github.com/khulnasoft/go-licenses/golicenses/presenter/html"
	"github.com/khulnasoft/go-licenses/golicenses/presenter/json"
	"github.com/khulnasoft/go-licenses/golicenses/presenter/markdown"
	"github.com/khulnasoft/go-licenses/golicenses/presenter/spdx" // Placeholder for SPDX presenter
	templatepresenter "github.com/khulnasoft/go-licenses/golicenses/presenter/template"
	"github.com/khulnasoft/go-licenses/golicenses/presenter/text"
)

type Presenter interface {
	Present(io.Writer) error
}

func GetPresenter(option Option, results <-chan golicenses.LicenseResult, templatePath ...string) Presenter {
	switch option {
	case CSVPresenter:
		return csv.NewPresenter(results)
	case JSONPresenter:
		return json.NewPresenter(results)
	case TextPresenter:
		return text.NewPresenter(results)
	case MarkdownPresenter:
		return markdown.NewPresenter(results)
	case HTMLPresenter:
		return html.NewPresenter(results)
	case SPDXPresenter:
		return spdx.NewPresenter(results)
	case TemplatePresenter: // TemplatePresenter, since Option is int and not in optionStr, use explicit value
		if len(templatePath) == 0 {
			return nil
		}
		pres, err := templatepresenter.NewPresenter(results, templatePath[0])
		if err != nil {
			return nil
		}
		return pres
	default:
		return nil
	}
}
