package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/licenseclassifier"
	"github.com/khulnasoft/go-licenses/golicenses"
	"github.com/khulnasoft/go-licenses/golicenses/licenses"
	"github.com/spf13/cobra"
)

var treeFormatFlag string

// treeCmd represents the tree command
var treeCmd = &cobra.Command{
	Use:   "tree [path]",
	Short: "Display dependency tree with licenses",
	Long:  `Display dependency tree with licenses. Path defaults to current directory.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		importPath := "."
		if len(args) > 0 {
			importPath = args[0]
		}

		// GetLicenseDBArchiveFetcher returns ([]byte, error) and is suitable directly for licenseclassifier.ArchiveFunc.
		dbOpt := licenseclassifier.ArchiveFunc(golicenses.GetLicenseDBArchiveFetcher)

		// Use the global appConfig from cmd/init.go.
		confidenceThreshold := appConfig.ConfidenceThreshold
		if confidenceThreshold == 0 { // Default if not set
			confidenceThreshold = 0.9
		}

		treeNodes, err := licenses.BuildDependencyTree(context.Background(), confidenceThreshold, dbOpt, importPath)
		if err != nil {
			return fmt.Errorf("failed to build dependency tree for %s: %w", importPath, err)
		}

		switch treeFormatFlag {
		case "json":
			jsonData, err := json.MarshalIndent(treeNodes, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal tree to JSON: %w", err)
			}
			fmt.Println(string(jsonData))
		case "ascii":
			asciiTree, err := formatTreeASCII(treeNodes)
			if err != nil {
				return fmt.Errorf("failed to format tree to ASCII: %w", err)
			}
			fmt.Print(asciiTree)
		case "dot":
			fmt.Printf("Output format '%s' for dependency tree is not yet implemented.\n", treeFormatFlag)
		default:
			return fmt.Errorf("unsupported tree format: %s. Supported formats are: ascii, json, dot", treeFormatFlag)
		}
		return nil
	},
}

func formatTreeASCII(roots []*licenses.DependencyNode) (string, error) {
	var builder strings.Builder
	// Keep track of visited nodes to prevent infinite loops in case of cyclic dependencies
	// although our BuildDependencyTree should already handle this by returning a DAG.
	// However, for formatting, it's good practice if the input *could* have cycles.
	visited := make(map[string]bool)

	for _, root := range roots {
		// For multiple root packages, they are independent trees, so no special prefix needed based on root index.
		buildASCIILevel(&builder, root, "", true, visited) // Treat each root as if it's the 'last child' for its own line drawing
	}
	return builder.String(), nil
}

// isLastChild is true if this node is the last in its parent's list of dependencies.
func buildASCIILevel(builder *strings.Builder, node *licenses.DependencyNode, prefix string, isLastChild bool, visited map[string]bool) {
	if visited[node.Path] {
		builder.WriteString(fmt.Sprintf("%s%s (%s) ... (cycle detected)\n", prefix, node.Path, node.License))
		return
	}
	visited[node.Path] = true

	licenseStr := ""
	if node.License != "" {
		licenseStr = fmt.Sprintf(" (License: %s)", node.License)
	} else if node.LicensePath != "" {
		licenseStr = fmt.Sprintf(" (License Path: %s)", node.LicensePath) // Fallback if name isn't resolved
	}

	builder.WriteString(fmt.Sprintf("%s%s%s\n", prefix, node.Path, licenseStr))

	for i, dep := range node.Dependencies {
		currentConnector := "├── "
		nextPrefix := prefix

		if isLastChild {
			nextPrefix += "    " // Parent was last, so use spaces for the vertical bar
		} else {
			nextPrefix += "│   " // Parent was not last, so continue the vertical bar
		}

		if i == len(node.Dependencies)-1 {
			currentConnector = "└── "
		}

		buildASCIILevel(builder, dep, nextPrefix+currentConnector, i == len(node.Dependencies)-1, visited)
	}
	// Unmark visited if you want to allow paths to be printed multiple times if they appear in different branches
	// delete(visited, node.Path) // For this tree view, we typically only want to expand a node once.
}

func init() {
	treeCmd.Flags().StringVar(&treeFormatFlag, "format", "ascii", "Output format: ascii, json, dot")
	rootCmd.AddCommand(treeCmd)
}

// getPath determines the target path from arguments or defaults to current directory.
// This is a helper, similar logic might exist in other cmd files.
func getPath(args []string) string {
	if len(args) > 0 {
		return args[0]
	}
	return "."
}
