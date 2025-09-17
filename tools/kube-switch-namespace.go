package main

import (
	"fmt"
	"os"

	"path/filepath"

	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// switchNamespaceRootCmd đại diện cho kube-switch-namespace command
var switchNamespaceRootCmd = &cobra.Command{
	Use:   "kube-switch-namespace [namespace-name]",
	Short: "Chuyển đổi namespace",
	Long: `kube-switch-namespace cho phép chuyển đổi namespace trong context hiện tại.
	
Nếu không có tên namespace, hiển thị namespace hiện tại.
	
Ví dụ:
  kube-switch-namespace                  # Hiển thị namespace hiện tại
  kube-switch-namespace my-app           # Chuyển sang namespace my-app`,
	RunE: runSwitchNamespace,
}

// runSwitchNamespace thực thi logic chuyển đổi namespace
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

	// Nếu không có argument, hiển thị namespace hiện tại
	if len(args) == 0 {
		currentNamespace := context.Namespace
		if currentNamespace == "" {
			currentNamespace = "default"
		}
		fmt.Printf("Current namespace: %s\n", currentNamespace)
		return nil
	}

	namespaceName := args[0]

	// Cập nhật namespace trong context
	context.Namespace = namespaceName
	config.Contexts[currentContext] = context

	// Lưu cấu hình
	err = clientcmd.WriteToFile(*config, kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to save kubeconfig: %w", err)
	}

	fmt.Printf("Switched to namespace '%s' in context '%s'\n", namespaceName, currentContext)
	return nil
}

// switchNamespaceGetKubeconfigPath trả về đường dẫn đến kubeconfig file
func switchNamespaceGetKubeconfigPath() string {
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

// main là entry point của kube-switch-namespace
func main() {
	if err := switchNamespaceRootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
