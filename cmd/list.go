package cmd

import (
	"fmt"
	"os"

	"github.com/khulnasoft/go-licenses/golicenses"
	"github.com/khulnasoft/go-licenses/golicenses/presenter"
	"github.com/spf13/cobra"
)

var gitRemotes []string

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all discovered licenses for a project (including dependencies)",
	Run: func(cmd *cobra.Command, args []string) {
		err := doListCmd(cmd, args)
		if err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
	},
}

var listFormatFlag string
var listTemplateFileFlag string

func init() {
	listCmd.Flags().StringArrayVar(&gitRemotes, "git-remote", []string{"origin", "upstream"}, "Remote Git repositories to try")
	listCmd.Flags().StringVar(&listFormatFlag, "format", "text", "Output format: text, csv, json, markdown, html, spdx, template")
	listCmd.Flags().StringVar(&listTemplateFileFlag, "template-file", "", "Path to Go template file (used only if --format=template)")
	rootCmd.AddCommand(listCmd)
}

func doListCmd(cmd *cobra.Command, args []string) error {
	// Assign CLI flags to appConfig fields for presenter selection
	appConfig.Format = listFormatFlag
	appConfig.TemplateFile = listTemplateFileFlag
	if appConfig.Format != "" {
		appConfig.Output = appConfig.Format
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
