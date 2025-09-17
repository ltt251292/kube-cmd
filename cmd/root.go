package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd is the base command when called without subcommands
var rootCmd = &cobra.Command{
	Use:   "kube",
	Short: "Kubernetes CLI helper tools",
	Long: `kube is a collection of small helper tools to work with your Kubernetes clusters.

Available tools:
  kube-pods              List pods
  kube-services          List services  
  kube-switch-context    Switch Kubernetes context
  kube-switch-namespace  Switch namespace
  kube-logs              Show pod logs
  kube-port-forward      Port forward to pods/services
  kube-exec              Execute commands in pods
  kube-deploy            Update Deployment image and wait for rollout
  kube-rollout           Restart or show rollout status for a Deployment

Use tools individually, or install all with 'make install-all'.`,
	RunE: listTools,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

// init initializes configuration for the root command
func init() {
	cobra.OnInitialize(initConfig)

	// Define flags and configuration settings
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kube.yaml)")
	rootCmd.PersistentFlags().StringP("namespace", "n", "", "Kubernetes namespace to use")
	rootCmd.PersistentFlags().StringP("context", "c", "", "Kubernetes context to use")

	// Bind flags with viper
	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("context", rootCmd.PersistentFlags().Lookup("context"))
}

// initConfig reads config file and environment variables if set
func initConfig() {
	if cfgFile != "" {
		// Use config file from flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}

		// Look for config in home directory with name ".kube" (no extension)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kube")
	}

	viper.AutomaticEnv() // read env variables

	// If a config file is found, read it
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// listTools prints the list of available kube-* tools
func listTools(cmd *cobra.Command, args []string) error {
	// Supported tools
	tools := []struct {
		name        string
		description string
	}{
		{"kube-pods", "List pods"},
		{"kube-services", "List services"},
		{"kube-switch-context", "Switch Kubernetes context"},
		{"kube-switch-namespace", "Switch namespace"},
		{"kube-logs", "Show pod logs"},
		{"kube-port-forward", "Port forward to pods/services"},
		{"kube-exec", "Execute commands in pods"},
		{"kube-deploy", "Update image and wait for rollout"},
		{"kube-rollout", "Restart or show rollout status"},
	}

	fmt.Println("Kubernetes CLI Helper Tools")
	fmt.Println("===========================")
	fmt.Println()

	// Check which tools are installed
	fmt.Println("Available tools:")
	for _, tool := range tools {
		status := "❌ Not installed"
		if _, err := exec.LookPath(tool.name); err == nil {
			status = "✅ Installed"
		}
		fmt.Printf("  %-20s %s - %s\n", tool.name, status, tool.description)
	}

	fmt.Println()
	fmt.Println("Installation:")
	fmt.Println("  make install-all    # Install all tools")
	fmt.Println("  make uninstall-all  # Uninstall all tools")
	fmt.Println("  make build-all      # Build all tools locally")
	fmt.Println()
	fmt.Println("Usage examples:")
	fmt.Println("  kube-pods                              # List pods")
	fmt.Println("  kube-services -A                       # List services in all namespaces")
	fmt.Println("  kube-switch-context production         # Switch to production context")
	fmt.Println("  kube-logs my-pod -f                    # Follow logs")
	fmt.Println("  kube-port-forward svc/my-svc 8080:80   # Port forward to service")
	fmt.Println("  kube-exec my-pod -- bash               # Exec into pod")

	return nil
}
