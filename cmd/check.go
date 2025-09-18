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
var checkStrictFlag bool
var checkSummaryFlag bool

func init() {
	checkCmd.Flags().StringVar(&checkFormatFlag, "format", "text", "Output format: text, csv, json, markdown, html, spdx, template")
	checkCmd.Flags().StringVar(&checkTemplateFileFlag, "template-file", "", "Path to Go template file (used only if --format=template)")
	checkCmd.Flags().BoolVar(&checkStrictFlag, "strict", false, "Fail on unknown or missing licenses")
	checkCmd.Flags().BoolVar(&checkSummaryFlag, "summary", false, "Print only a summary of license types found")
	rootCmd.AddCommand(checkCmd)
}

// TODO: add to check the ability to check for 3rd party notices are in the repo

// doCheckCmd runs the license check logic for the check command.
// Now supports --format and --template-file for output customization.
func doCheckCmd(cmd *cobra.Command, args []string) error {
	// Assign CLI flags to appConfig fields
	appConfig.Format = checkFormatFlag
	appConfig.TemplateFile = checkTemplateFileFlag
	appConfig.Strict = checkStrictFlag
	appConfig.Summary = checkSummaryFlag
	if appConfig.Format != "" {
		appConfig.Output = appConfig.Format // Ensure Output is set for presenter.ParseOption in config.Build()
	}
	// Re-run Build to ensure PresenterOpt is set based on Format if it was changed by a flag
	if err := appConfig.Build(); err != nil {
		return fmt.Errorf("config error: %w", err)
	}

	ruleAction := appConfig.Action()
	rulePatterns := appConfig.Patterns()

	if ruleAction == golicenses.UnknownAction {
		return fmt.Errorf("no rules configured (permit or forbid must be set in .golicenses.yaml or via flags not yet implemented)")
	}

	rules, err := golicenses.NewRules(ruleAction, rulePatterns, appConfig.IgnorePkg...)
	if err != nil {
		return fmt.Errorf("could not parse rules: %w", err)
	}

	var paths []string
	if len(args) > 0 {
		paths = args
	} else {
		paths = []string{"."}
	}
	licenseFinder := golicenses.NewLicenseFinder(paths, gitRemotes, 0.9)

	rawResultsChan, err := licenseFinder.Find()
	if err != nil {
		return err
	}

	// Collect results
	var collectedResults []golicenses.LicenseResult
	var unknownLicenseLibraries []string
	licenseSummary := make(map[string]int)

	for res := range rawResultsChan {
		collectedResults = append(collectedResults, res)
		licenseKey := res.License
		if licenseKey == "" {
			licenseKey = "Unknown"
		}
		licenseSummary[licenseKey]++

		if appConfig.Strict && (res.License == "" || res.License == "Unknown") {
			unknownLicenseLibraries = append(unknownLicenseLibraries, res.Library)
		}
	}

	if appConfig.Strict && len(unknownLicenseLibraries) > 0 {
		return fmt.Errorf("strict mode: found unknown/missing licenses for libraries: %v", unknownLicenseLibraries)
	}

	// Evaluate rules against all collected results
	allowed, violations, err := rules.Evaluate(collectedResults...)
	if err != nil {
		return fmt.Errorf("error evaluating rules: %w", err)
	}

	if appConfig.Summary {
		fmt.Println("License Summary:")
		for lic, count := range licenseSummary {
			fmt.Printf("  %s: %d\n", lic, count)
		}
		// If not allowed (rules violated), return an error to indicate failure
		if !allowed {
			return fmt.Errorf("license rule violations detected (summary mode). Problematic licenses for: %v", getLibrariesFromResults(violations))
		}
		return nil // Summary printed, and no rule violations or strict failures
	}

	// If not summary mode, proceed with the standard presenter
	// Create a new channel from the collected (and potentially filtered for violations by presenter) results
	resultStreamForPresenter := make(chan golicenses.LicenseResult, len(collectedResults))
	go func() {
		defer close(resultStreamForPresenter)
		// Presenters like 'text' often show only violations. If rules passed, it might show nothing.
		// For 'check', we want to show violations if rules failed, or all if rules passed (depending on presenter).
		// The current presenters (e.g., text) are designed to show violations for 'check'.
		// If allowed is false, we send violations. Otherwise, some presenters might expect all results.
		// However, the 'text' presenter specifically filters for violations when used with 'check'.
		// Let's stick to sending all results and let the presenter decide, or refine presenter later.
		// For now, to ensure 'text' presenter shows violations, we pass 'violations' if not allowed.
		// This is a bit of a hack; presenters should ideally be more aware of 'check' context.
		resultsToPresent := collectedResults
		if !allowed && appConfig.PresenterOpt == presenter.TextPresenter { // Special handling for text presenter to show violations
			resultsToPresent = violations
		}
		for _, res := range resultsToPresent {
			resultStreamForPresenter <- res
		}
	}()

	opt := appConfig.PresenterOpt
	var pres presenter.Presenter
	if opt == presenter.TemplatePresenter {
		if appConfig.TemplateFile == "" {
			return fmt.Errorf("--template-file must be provided when --format=template")
		}
		pres = presenter.GetPresenter(opt, resultStreamForPresenter, appConfig.TemplateFile)
	} else {
		pres = presenter.GetPresenter(opt, resultStreamForPresenter)
	}

	if pres == nil {
		return fmt.Errorf("invalid presenter for option: %v", opt)
	}

	if err := pres.Present(os.Stdout); err != nil {
		return fmt.Errorf("error during presentation: %w", err)
	}

	// If rules were violated, return an error to ensure check command fails
	if !allowed {
		return fmt.Errorf("license rule violations detected. Problematic licenses for: %v", getLibrariesFromResults(violations))
	}

	return nil
}
