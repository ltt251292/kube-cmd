package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

// switchContextRootCmd represents the kube-switch-context command
var switchContextRootCmd = &cobra.Command{
	Use:   "kube-switch-context [context-name]",
	Short: "Switch Kubernetes context",
	Long: `kube-switch-context allows quick switching between Kubernetes contexts.
	
If no context name is provided, displays list of available contexts.
	
Examples:
  kube-switch-context                    # Display list of contexts
  kube-switch-context production         # Switch to production context`,
	RunE: runSwitchContext,
}

// runSwitchContext executes the context switching logic
func runSwitchContext(cmd *cobra.Command, args []string) error {
	kubeconfig := switchContextGetKubeconfigPath()

	// Load kubeconfig
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// If no argument, display list of contexts
	if len(args) == 0 {
		return listContexts(config)
	}

	contextName := args[0]

	// Check if context exists
	if _, exists := config.Contexts[contextName]; !exists {
		return fmt.Errorf("context '%s' not found", contextName)
	}

	// Update current context
	config.CurrentContext = contextName

	// Save configuration
	err = clientcmd.WriteToFile(*config, kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to save kubeconfig: %w", err)
	}

	fmt.Printf("Switched to context '%s'\n", contextName)
	return nil
}

// listContexts displays list of all contexts
func listContexts(config *api.Config) error {
	headers := []string{"CURRENT", "NAME", "CLUSTER", "USER", "NAMESPACE"}
	var rows [][]string

	for name, context := range config.Contexts {
		current := ""
		if name == config.CurrentContext {
			current = "*"
		}

		namespace := context.Namespace
		if namespace == "" {
			namespace = "default"
		}

		rows = append(rows, []string{current, name, context.Cluster, context.AuthInfo, namespace})
	}

	renderTable(headers, rows)
	return nil
}

// renderTable prints an ASCII table with simple borders
// headers: column headers, rows: row data
func renderTable(headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	for c, h := range headers {
		w := displayWidth(h)
		if w > widths[c] {
			widths[c] = w
		}
	}
	for _, row := range rows {
		for c, cell := range row {
			w := displayWidth(cell)
			if w > widths[c] {
				widths[c] = w
			}
		}
	}

	printSeparator(widths)
	fmt.Println("| " + joinRow(headers, widths) + " |")
	printSeparator(widths)
	for _, row := range rows {
		fmt.Println("| " + joinRow(row, widths) + " |")
	}
	printSeparator(widths)
}

// displayWidth returns display length (excluding ANSI codes)
func displayWidth(s string) int {
	return len(stripANSI(s))
}

// stripANSI removes ANSI color codes for accurate width calculation
func stripANSI(s string) string {
	ansi := regexp.MustCompile("\\x1b\\[[0-9;]*m")
	return ansi.ReplaceAllString(s, "")
}

// joinRow left-aligns each cell and joins with column separator
func joinRow(cols []string, widths []int) string {
	parts := make([]string, len(cols))
	for i, col := range cols {
		pad := widths[i] - displayWidth(col)
		if pad < 0 {
			pad = 0
		}
		parts[i] = col + strings.Repeat(" ", pad)
	}
	return strings.Join(parts, " | ")
}

// printSeparator prints border line based on column widths
func printSeparator(widths []int) {
	b := strings.Builder{}
	b.WriteString("+")
	for i, w := range widths {
		b.WriteString(strings.Repeat("-", w+2))
		if i == len(widths)-1 {
			b.WriteString("+")
		} else {
			b.WriteString("+")
		}
	}
	fmt.Println(b.String())
}

// switchContextGetKubeconfigPath returns path to kubeconfig file
func switchContextGetKubeconfigPath() string {
	// Check KUBECONFIG environment variable
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	// Use default path
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}

	return ""
}

// main is the entry point of kube-switch-context
func main() {
	if err := switchContextRootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
