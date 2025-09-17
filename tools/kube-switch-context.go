package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/homedir"
)

// switchContextRootCmd đại diện cho kube-switch-context command
var switchContextRootCmd = &cobra.Command{
	Use:   "kube-switch-context [context-name]",
	Short: "Chuyển đổi Kubernetes context",
	Long: `kube-switch-context cho phép chuyển đổi nhanh giữa các Kubernetes contexts.
	
Nếu không có tên context, hiển thị danh sách contexts có sẵn.
	
Ví dụ:
  kube-switch-context                    # Hiển thị danh sách contexts
  kube-switch-context production         # Chuyển sang context production`,
	RunE: runSwitchContext,
}

// runSwitchContext thực thi logic chuyển đổi context
func runSwitchContext(cmd *cobra.Command, args []string) error {
	kubeconfig := switchContextGetKubeconfigPath()

	// Load kubeconfig
	config, err := clientcmd.LoadFromFile(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Nếu không có argument, hiển thị danh sách contexts
	if len(args) == 0 {
		return listContexts(config)
	}

	contextName := args[0]

	// Kiểm tra context có tồn tại không
	if _, exists := config.Contexts[contextName]; !exists {
		return fmt.Errorf("context '%s' not found", contextName)
	}

	// Cập nhật current context
	config.CurrentContext = contextName

	// Lưu cấu hình
	err = clientcmd.WriteToFile(*config, kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to save kubeconfig: %w", err)
	}

	fmt.Printf("Switched to context '%s'\n", contextName)
	return nil
}

// listContexts hiển thị danh sách tất cả contexts
func listContexts(config *api.Config) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "CURRENT\tNAME\tCLUSTER\tUSER\tNAMESPACE")

	for name, context := range config.Contexts {
		current := ""
		if name == config.CurrentContext {
			current = "*"
		}

		namespace := context.Namespace
		if namespace == "" {
			namespace = "default"
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			current,
			name,
			context.Cluster,
			context.AuthInfo,
			namespace,
		)
	}

	w.Flush()
	return nil
}

// switchContextGetKubeconfigPath trả về đường dẫn đến kubeconfig file
func switchContextGetKubeconfigPath() string {
	// Kiểm tra biến môi trường KUBECONFIG
	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		return kubeconfig
	}

	// Sử dụng đường dẫn mặc định
	if home := homedir.HomeDir(); home != "" {
		return filepath.Join(home, ".kube", "config")
	}

	return ""
}

// main là entry point của kube-switch-context
func main() {
	if err := switchContextRootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
