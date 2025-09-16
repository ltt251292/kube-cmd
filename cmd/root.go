package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd đại diện cho base command khi được gọi mà không có subcommands
var rootCmd = &cobra.Command{
	Use:   "kube",
	Short: "Kubernetes CLI helper tools",
	Long: `kube là một collection các tools để làm việc với Kubernetes cluster.

Các tools có sẵn:
  kube-pods              Xem danh sách pods
  kube-services          Xem danh sách services  
  kube-switch-context    Chuyển đổi Kubernetes context
  kube-switch-namespace  Chuyển đổi namespace
  kube-logs              Xem logs của pods
  kube-port-forward      Port forwarding đến pods
  kube-exec              Thực thi commands trong pods

Sử dụng các tools riêng lẻ hoặc cài đặt tất cả với 'make install-all'.`,
	RunE: listTools,
}

// Execute thêm tất cả child commands vào root command và set flags một cách phù hợp.
// Đây là function được gọi bởi main.main(). Nó chỉ cần xảy ra một lần với rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

// init khởi tạo cấu hình cho root command
func init() {
	cobra.OnInitialize(initConfig)

	// Định nghĩa flags và configuration settings
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kube.yaml)")
	rootCmd.PersistentFlags().StringP("namespace", "n", "", "Kubernetes namespace to use")
	rootCmd.PersistentFlags().StringP("context", "c", "", "Kubernetes context to use")

	// Bind flags với viper
	viper.BindPFlag("namespace", rootCmd.PersistentFlags().Lookup("namespace"))
	viper.BindPFlag("context", rootCmd.PersistentFlags().Lookup("context"))
}

// initConfig đọc config file và ENV variables nếu được set
func initConfig() {
	if cfgFile != "" {
		// Sử dụng config file từ flag
		viper.SetConfigFile(cfgFile)
	} else {
		// Tìm home directory
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting home directory: %v\n", err)
			os.Exit(1)
		}

		// Tìm config trong home directory với tên ".kube" (không có extension)
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".kube")
	}

	viper.AutomaticEnv() // đọc các biến môi trường khớp với KUBE_*

	// Nếu tìm thấy config file, đọc nó
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}

// listTools hiển thị danh sách các kube-* tools có sẵn
func listTools(cmd *cobra.Command, args []string) error {
	// Danh sách tools được hỗ trợ
	tools := []struct {
		name        string
		description string
	}{
		{"kube-pods", "Xem danh sách pods"},
		{"kube-services", "Xem danh sách services"},
		{"kube-switch-context", "Chuyển đổi Kubernetes context"},
		{"kube-switch-namespace", "Chuyển đổi namespace"},
		{"kube-logs", "Xem logs của pods"},
		{"kube-port-forward", "Port forwarding đến pods"},
		{"kube-exec", "Thực thi commands trong pods"},
	}

	fmt.Println("Kubernetes CLI Helper Tools")
	fmt.Println("===========================")
	fmt.Println()

	// Kiểm tra tools nào đã được cài đặt
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
	fmt.Println("  kube-port-forward my-pod 8080:80       # Port forward")
	fmt.Println("  kube-exec my-pod -- bash               # Exec into pod")

	return nil
}
