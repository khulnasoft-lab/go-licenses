package cmd

import "github.com/khulnasoft/go-licenses/golicenses"

// getLibrariesFromResults extracts library names from a slice of LicenseResult.
func getLibrariesFromResults(results []golicenses.LicenseResult) []string {
	libraries := make([]string, len(results))
	for i, r := range results {
		libraries[i] = r.Library
	}
	return libraries
}
