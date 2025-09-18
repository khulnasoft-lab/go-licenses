// Copyright 2019 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package licenses

import (
	"context"
	"fmt"
	"go/build"
	"net/url"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/golang/glog"
	"github.com/google/licenseclassifier" // Added this import
	"golang.org/x/tools/go/packages"
)

var (
	// TODO(RJPercival): Support replacing "master" with Go Module version
	repoPathPrefixes = map[string]string{
		"github.com":    "blob/master/",
		"bitbucket.org": "src/master/",
	}
)

// Library is a collection of packages covered by the same license file.
type Library struct {
	// LicensePath is the path of the file containing the library's license.
	LicensePath string
	// Packages contains import paths for Go packages in this library.
	// It may not be the complete set of all packages in the library.
	Packages []string
}

// PackagesError aggregates all Packages[].Errors into a single error.
type PackagesError struct {
	pkgs []*packages.Package
}

func (e PackagesError) Error() string {
	var str strings.Builder
	str.WriteString(fmt.Sprintf("errors for %q:", e.pkgs))
	packages.Visit(e.pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			str.WriteString(fmt.Sprintf("\n%s: %s", pkg.PkgPath, err))
		}
	})
	return str.String()
}

// Libraries returns the collection of libraries used by this package, directly or transitively.
// A library is a collection of one or more packages covered by the same license file.
// Packages not covered by a license will be returned as individual libraries.
// Standard library packages will be ignored.
func Libraries(ctx context.Context, classifier Classifier, importPaths ...string) ([]*Library, error) {
	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedImports | packages.NeedDeps | packages.NeedFiles | packages.NeedName,
	}

	rootPkgs, err := packages.Load(cfg, importPaths...)
	if err != nil {
		return nil, err
	}

	pkgs := map[string]*packages.Package{}
	pkgsByLicense := make(map[string][]*packages.Package)
	errorOccurred := false
	packages.Visit(rootPkgs, func(p *packages.Package) bool {
		if len(p.Errors) > 0 {
			errorOccurred = true
			return false
		}
		if isStdLib(p) {
			// No license requirements for the Go standard library.
			return false
		}
		if len(p.OtherFiles) > 0 {
			glog.Warningf("%q contains non-Go code that can't be inspected for further dependencies:\n%s", p.PkgPath, strings.Join(p.OtherFiles, "\n"))
		}
		var pkgDir string
		switch {
		case len(p.GoFiles) > 0:
			pkgDir = filepath.Dir(p.GoFiles[0])
		case len(p.CompiledGoFiles) > 0:
			pkgDir = filepath.Dir(p.CompiledGoFiles[0])
		case len(p.OtherFiles) > 0:
			pkgDir = filepath.Dir(p.OtherFiles[0])
		default:
			// This package is empty - nothing to do.
			return true
		}
		licensePath, err := Find(pkgDir, classifier)
		if err != nil {
			glog.Errorf("Failed to find license for %s: %v", p.PkgPath, err)
		}
		pkgs[p.PkgPath] = p
		pkgsByLicense[licensePath] = append(pkgsByLicense[licensePath], p)
		return true
	}, nil)
	if errorOccurred {
		return nil, PackagesError{
			pkgs: rootPkgs,
		}
	}

	libraries := make([]*Library, 0)
	for licensePath, pkgs := range pkgsByLicense {
		if licensePath == "" {
			// No license for these packages - return each one as a separate library.
			for _, p := range pkgs {
				libraries = append(libraries, &Library{
					Packages: []string{p.PkgPath},
				})
			}
			continue
		}
		lib := &Library{
			LicensePath: licensePath,
		}
		for _, pkg := range pkgs {
			lib.Packages = append(lib.Packages, pkg.PkgPath)
		}
		libraries = append(libraries, lib)
	}
	return libraries, nil
}

// Name is the common prefix of the import paths for all of the packages in this library.
func (l *Library) Name() string {
	return commonAncestor(l.Packages)
}

func commonAncestor(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	if len(paths) == 1 {
		return paths[0]
	}
	sort.Strings(paths)
	min, max := paths[0], paths[len(paths)-1]
	lastSlashIndex := 0
	for i := 0; i < len(min) && i < len(max); i++ {
		if min[i] != max[i] {
			return min[:lastSlashIndex]
		}
		if min[i] == '/' {
			lastSlashIndex = i
		}
	}
	return min
}

func (l *Library) String() string {
	return l.Name()
}

// FileURL attempts to determine the URL for a file in this library.
// This only works for certain supported package prefixes, such as github.com,
// bitbucket.org and googlesource.com. Prefer GitRepo.FileURL() if possible.
func (l *Library) FileURL(filePath string) (*url.URL, error) {
	relFilePath, err := filepath.Rel(filepath.Dir(l.LicensePath), filePath)
	if err != nil {
		return nil, err
	}
	nameParts := strings.SplitN(l.Name(), "/", 4)
	if len(nameParts) < 3 {
		return nil, fmt.Errorf("cannot determine URL for %q package", l.Name())
	}
	host, user, project := nameParts[0], nameParts[1], nameParts[2]
	pathPrefix, ok := repoPathPrefixes[host]
	if !ok {
		return nil, fmt.Errorf("unsupported package host %q for %q", host, l.Name())
	}
	if len(nameParts) == 4 {
		pathPrefix = path.Join(pathPrefix, nameParts[3])
	}
	return &url.URL{
		Scheme: "https",
		Host:   host,
		Path:   path.Join(user, project, pathPrefix, relFilePath),
	}, nil
}

// isStdLib returns true if this package is part of the Go standard library.
func isStdLib(pkg *packages.Package) bool {
	if len(pkg.GoFiles) == 0 {
		return false
	}
	return strings.HasPrefix(pkg.GoFiles[0], build.Default.GOROOT)
}

// DependencyNode represents a node in the dependency tree.
type DependencyNode struct {
	Path         string            `json:"path"`
	License      string            `json:"license,omitempty"` // Optional: We can populate this later if needed for tree view
	LicensePath  string            `json:"licensePath,omitempty"`
	Dependencies []*DependencyNode `json:"dependencies,omitempty"`
}

// BuildDependencyTree constructs a dependency tree for the given import paths.
// It returns the root nodes of the dependency trees (for each importPath provided).
func BuildDependencyTree(ctx context.Context, confidenceThreshold float64, dbOption licenseclassifier.OptionFunc, importPaths ...string) ([]*DependencyNode, error) {
	classifier, err := NewClassifier(confidenceThreshold, dbOption)
	if err != nil {
		return nil, fmt.Errorf("failed to create classifier for BuildDependencyTree: %w", err)
	}

	cfg := &packages.Config{
		Context: ctx,
		Mode:    packages.NeedImports | packages.NeedDeps | packages.NeedFiles | packages.NeedName,
	}

	rootPkgs, err := packages.Load(cfg, importPaths...)
	if err != nil {
		return nil, err
	}

	if packages.PrintErrors(rootPkgs) > 0 {
		return nil, PackagesError{pkgs: rootPkgs} // Assuming PackagesError is suitable
	}

	// visited keeps track of processed packages to avoid cycles and redundant work.
	visited := make(map[string]*DependencyNode)
	var resultRoots []*DependencyNode

	var buildNode func(pkg *packages.Package) *DependencyNode
	buildNode = func(pkg *packages.Package) *DependencyNode {
		if node, ok := visited[pkg.PkgPath]; ok {
			return node // Already processed or currently processing (cycle)
		}

		if isStdLib(pkg) {
			return nil // Skip standard library packages
		}

		node := &DependencyNode{Path: pkg.PkgPath}
		visited[pkg.PkgPath] = node // Mark as visited early to handle cycles

		// Attempt to find license for this package node (optional for basic tree)
		var pkgDir string
		switch {
		case len(pkg.GoFiles) > 0:
			pkgDir = filepath.Dir(pkg.GoFiles[0])
		case len(pkg.CompiledGoFiles) > 0:
			pkgDir = filepath.Dir(pkg.CompiledGoFiles[0])
		case len(pkg.OtherFiles) > 0:
			pkgDir = filepath.Dir(pkg.OtherFiles[0])
		}
		if pkgDir != "" {
			licensePath, findLicErr := Find(pkgDir, classifier) // Find still needs a classifier instance
			if findLicErr == nil && licensePath != "" {
				node.LicensePath = licensePath
				licenseName, _, identifyErr := classifier.Identify(licensePath)
				if identifyErr == nil {
					node.License = licenseName
				}
			}
		}

		for _, impPkg := range pkg.Imports {
			if depNode := buildNode(impPkg); depNode != nil {
				node.Dependencies = append(node.Dependencies, depNode)
			}
		}
		return node
	}

	for _, rootPkg := range rootPkgs {
		if rootNode := buildNode(rootPkg); rootNode != nil {
			resultRoots = append(resultRoots, rootNode)
		}
	}

	return resultRoots, nil
}
