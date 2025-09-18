package spdx

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/khulnasoft/go-licenses/golicenses"
)

// Presenter outputs license information in SPDX tag-value format.
// https://spdx.github.io/spdx-spec/v2.3/SPDX-tag-value-format/
type Presenter struct {
	results <-chan golicenses.LicenseResult
}

// NewPresenter creates a new SPDX presenter.
func NewPresenter(results <-chan golicenses.LicenseResult) *Presenter {
	return &Presenter{results: results}
}

// Present writes the SPDX report to the given writer.
func (p *Presenter) Present(w io.Writer) error {
	// SPDX Document Creation Information
	fmt.Fprintf(w, "SPDXVersion: SPDX-2.3\n")
	fmt.Fprintf(w, "DataLicense: CC0-1.0\n")
	fmt.Fprintf(w, "SPDXID: SPDXRef-DOCUMENT\n")
	fmt.Fprintf(w, "DocumentName: go-licenses-report\n")               // Can be customized
	fmt.Fprintf(w, "DocumentNamespace: urn:uuid:%s\n", generateUUID()) // Placeholder for actual UUID generation
	fmt.Fprintf(w, "Creator: Tool: go-licenses (github.com/khulnasoft/go-licenses)\n")
	fmt.Fprintf(w, "Created: %s\n", time.Now().UTC().Format("2006-01-02T15:04:05Z"))
	fmt.Fprintf(w, "\n")

	// Packages section
	for res := range p.results {
		fmt.Fprintf(w, "##### Package: %s\n\n", res.Library)
		fmt.Fprintf(w, "PackageName: %s\n", res.Library)
		fmt.Fprintf(w, "SPDXID: SPDXRef-Package-%s\n", sanitizeSPDXID(res.Library))
		// Attempt to get a download location if URL is a VCS URL
		downloadLocation := res.URL
		if downloadLocation == "" {
			downloadLocation = "NOASSERTION"
		} else if downloadLocation != "NOASSERTION" {
			// Prepend "git+" only for likely VCS URLs
			isVCS := strings.Contains(downloadLocation, "github.com") ||
				strings.Contains(downloadLocation, "gitlab.com") ||
				strings.Contains(downloadLocation, "bitbucket.org") ||
				strings.HasSuffix(downloadLocation, ".git")

			if isVCS && !strings.HasPrefix(downloadLocation, "git+") {
				downloadLocation = "git+" + downloadLocation
			}
		}
		fmt.Fprintf(w, "PackageDownloadLocation: %s\n", downloadLocation)
		fmt.Fprintf(w, "FilesAnalyzed: false\n") // We are not analyzing individual files
		// LicenseConcluded: Use the license string directly. For more accuracy, map to SPDX license list IDs.
		// For now, using NOASSERTION if license string is complex or not a simple SPDX ID.
		concludedLicense := res.License
		if !isValidSPDXLicenseID(concludedLicense) {
			concludedLicense = "NOASSERTION" // Or attempt to parse/map common names
		}
		fmt.Fprintf(w, "LicenseConcluded: %s\n", concludedLicense)
		// LicenseDeclared: Same as Concluded for now, as we don't have separate declared vs. found info.
		fmt.Fprintf(w, "LicenseDeclared: %s\n", concludedLicense)
		fmt.Fprintf(w, "PackageLicenseComments: Source path: %s\n", res.Path)
		fmt.Fprintf(w, "PackageCopyrightText: NOASSERTION\n") // Copyright info not available in LicenseResult
		fmt.Fprintf(w, "\n")
	}

	return nil
}

// generateUUID generates a new UUID string.
func generateUUID() string {
	return uuid.NewString()
}

// sanitizeSPDXID replaces characters not allowed in SPDXID strings.
// SPDXID strings must be composed of letters, numbers, ".", and "-".
func sanitizeSPDXID(name string) string {
	r := strings.NewReplacer("/", "-", "@", "-at-", ":", "-col-")
	// Further sanitization might be needed for other special characters.
	return r.Replace(name)
}

// isValidSPDXLicenseID checks if the given string is a (very simplified) valid SPDX license identifier.
// This is a basic check and does not cover the full SPDX license expression syntax or the complete list.
// It's recommended to use a dedicated SPDX library for full compliance.
func isValidSPDXLicenseID(id string) bool {
	lowerID := strings.ToLower(id)

	// Disallow complex expressions for this basic check
	if strings.Contains(lowerID, " and ") || strings.Contains(lowerID, " or ") {
		return false
	}

	// Common simple SPDX license identifiers (case-insensitive)
	// This list is not exhaustive.
	switch lowerID {
	case "mit", "apache-2.0", "mpl-2.0":
		return true
	case "gpl-2.0-only", "gpl-2.0-or-later", "gpl-3.0-only", "gpl-3.0-or-later":
		return true
	case "lgpl-2.0-only", "lgpl-2.0-or-later", "lgpl-2.1-only", "lgpl-2.1-or-later", "lgpl-3.0-only", "lgpl-3.0-or-later":
		return true
	case "bsd-2-clause", "bsd-3-clause", "isc", "unlicense":
		return true
	case "cc0-1.0": // Creative Commons Zero v1.0 Universal
		return true
	default:
		// Handle common pattern for licenses with '+' (e.g., "GPL-2.0+")
		// by checking if the "-or-later" equivalent is in our list.
		if strings.HasSuffix(lowerID, "+") {
			baseLicense := strings.TrimSuffix(lowerID, "+") + "-or-later"
			// Recursive call, but on a transformed string, so it should terminate.
			return isValidSPDXLicenseID(baseLicense)
		}
		return false
	}
}
