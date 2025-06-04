package cmd

import (
	"fmt"
	"os"

	"github.com/gookit/color"
	"github.com/khulnasoft/go-licenses/golicenses"
	"github.com/khulnasoft/go-licenses/golicenses/presenter"

	"github.com/spf13/cobra"
)

// checkCmd represents the check command
var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "ensure only select licenses or types are used",
	Run: func(cmd *cobra.Command, args []string) {
		err := doCheckCmd(cmd, args)
		if err != nil {
			fmt.Fprintln(os.Stderr, color.Style{color.Red, color.Bold}.Sprint(err.Error()))
			os.Exit(1)
		}
		color.Style{color.Green, color.Bold}.Println("Passed!")
	},
}

var checkFormatFlag string
var checkTemplateFileFlag string

func init() {
	checkCmd.Flags().StringVar(&checkFormatFlag, "format", "text", "Output format: text, csv, json, markdown, html, spdx, template")
	checkCmd.Flags().StringVar(&checkTemplateFileFlag, "template-file", "", "Path to Go template file (used only if --format=template)")
	rootCmd.AddCommand(checkCmd)
}

// TODO: add to check the ability to check for 3rd party notices are in the repo

// doCheckCmd runs the license check logic for the check command.
// Now supports --format and --template-file for output customization.
func doCheckCmd(cmd *cobra.Command, args []string) error {
	// Assign CLI flags to appConfig fields for presenter selection
	appConfig.Format = checkFormatFlag
	appConfig.TemplateFile = checkTemplateFileFlag
	if appConfig.Format != "" {
		appConfig.Output = appConfig.Format
	}

	var err error
	switch {
	case len(appConfig.Permit) > 0:
		_, err = golicenses.NewRules(golicenses.AllowAction, appConfig.Permit, appConfig.IgnorePkg...)
		fmt.Fprintf(os.Stderr, "Allow Rules: %+v\n", appConfig.Permit)
	case len(appConfig.Forbid) > 0:
		_, err = golicenses.NewRules(golicenses.DenyAction, appConfig.Forbid, appConfig.IgnorePkg...)
		fmt.Fprintf(os.Stderr, "Deny Rules: %+v\n", appConfig.Forbid)
	default:
		return fmt.Errorf("no rules configured")
	}
	if err != nil {
		return fmt.Errorf("could not parse rules: %+v", err)
	}

	var paths []string
	if len(args) > 0 {
		paths = args
	} else {
		paths = []string{"."}
	}
	licenseFinder := golicenses.NewLicenseFinder(paths, gitRemotes, 0.9)

	resultStream, err := licenseFinder.Find()
	if err != nil {
		return err
	}

	opt := appConfig.PresenterOpt
	var pres presenter.Presenter
	if int(opt) == 6 { // TemplatePresenter
		if appConfig.Output == "" {
			return fmt.Errorf("--template-file must be provided when --format=template")
		}
		pres = presenter.GetPresenter(opt, resultStream, appConfig.Output)
	} else {
		pres = presenter.GetPresenter(opt, resultStream)
	}
	if pres == nil {
		return fmt.Errorf("invalid presenter for option: %v", opt)
	}
	return pres.Present(os.Stdout)
}
