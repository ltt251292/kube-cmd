package main

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// switchNamespaceRootCmd represents the kube-switch-namespace command
var switchNamespaceRootCmd = &cobra.Command{
	Use:   "kube-switch-namespace [namespace-name]",
	Short: "Switch namespace",
	Long: `kube-switch-namespace allows switching namespace in the current context.
	
If no namespace name is provided, displays current namespace.
	
Examples:
  kube-switch-namespace                  # Display current namespace
  kube-switch-namespace my-app           # Switch to namespace my-app`,
	RunE: runSwitchNamespace,
}

// runSwitchNamespace executes the namespace switching logic
func runSwitchNamespace(cmd *cobra.Command, args []string) error {
	kubeconfig := switchNamespaceGetKubeconfigPath()

	// Load kubeconfig
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	currentContext := config.CurrentContext
	if currentContext == "" {
		return fmt.Errorf("no current context set")
	}

	context, exists := config.Contexts[currentContext]
	if !exists {
		return fmt.Errorf("current context '%s' not found", currentContext)
	}

	// If no argument, display current namespace
	if len(args) == 0 {
		currentNamespace := context.Namespace
		if currentNamespace == "" {
			currentNamespace = "default"
		}
		fmt.Printf("Current namespace: %s\n", currentNamespace)
		return nil
	}

	namespaceName := args[0]

	// Update namespace in context
	context.Namespace = namespaceName
	config.Contexts[currentContext] = context

	// Save configuration
	err = clientcmd.WriteToFile(*config, kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to save kubeconfig: %w", err)
	}

	fmt.Printf("Switched to namespace '%s' in context '%s'\n", namespaceName, currentContext)
	return nil
}

// switchNamespaceGetKubeconfigPath returns path to kubeconfig file
func switchNamespaceGetKubeconfigPath() string {
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

// main is the entry point of kube-switch-namespace
func main() {
	if err := switchNamespaceRootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
