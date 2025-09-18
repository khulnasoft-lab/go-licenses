package presenter

import "strings"

const (
	UnknownPresenter Option = iota
	CSVPresenter
	JSONPresenter
	TextPresenter
	MarkdownPresenter
	HTMLPresenter
	SPDXPresenter     // Added for SPDX output
	TemplatePresenter // Added for template-based output
)

var optionStr = []string{
	"UnknownPresenter",
	"csv",
	"json",
	"text",
	"markdown",
	"html",
	"spdx",
	"template",
}

var Options = []Option{
	CSVPresenter,
	JSONPresenter,
	TextPresenter,
	MarkdownPresenter,
	HTMLPresenter,
	SPDXPresenter,
	TemplatePresenter,
}

type Option int

func ParseOption(userStr string) Option {
	switch strings.ToLower(userStr) {
	case strings.ToLower(CSVPresenter.String()):
		return CSVPresenter
	case strings.ToLower(JSONPresenter.String()):
		return JSONPresenter
	case strings.ToLower(TextPresenter.String()):
		return TextPresenter
	case strings.ToLower(MarkdownPresenter.String()):
		return MarkdownPresenter
	case strings.ToLower(HTMLPresenter.String()):
		return HTMLPresenter
	case "spdx": // Directly check for "spdx" string
		return SPDXPresenter
	case "template": // Directly check for "template" string
		return TemplatePresenter
	default:
		return UnknownPresenter
	}
}

func (o Option) String() string {
	if int(o) >= len(optionStr) || o < 0 {
		return optionStr[0]
	}

	return optionStr[o]
}
