package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"kube/pkg/k8s"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	logsNamespace     string
	logsKubeContext   string
	logsFollow        bool
	logsTailLines     int64
	logsSinceSeconds  int64
	logsContainerName string
	logsTimestamps    bool
)

// logsRootCmd đại diện cho kube-logs command
var logsRootCmd = &cobra.Command{
	Use:   "kube-logs [pod-name]",
	Short: "Hiển thị logs của pod",
	Long: `kube-logs hiển thị logs của một pod cụ thể.
	
Tính năng:
- Follow logs real-time (-f)
- Hiển thị số dòng cuối (-t, --tail)
- Hiển thị logs từ timestamp cụ thể (--since)
- Chọn container cụ thể (-c, --container)
- Timestamps trong output (--timestamps)

Ví dụ:
  kube-logs my-pod                       # Xem logs của pod
  kube-logs my-pod -f                    # Follow logs real-time
  kube-logs my-pod -c container-name     # Logs của container cụ thể`,
	Args: cobra.ExactArgs(1),
	RunE: runLogs,
}

// runLogs thực thi logic hiển thị logs
func runLogs(cmd *cobra.Command, args []string) error {
	podName := args[0]

	client, err := k8s.NewClient("", logsKubeContext)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	targetNamespace := logsNamespace
	if targetNamespace == "" {
		targetNamespace = "default"
	}

	// Lấy thông tin pod để kiểm tra containers
	pod, err := client.Clientset.CoreV1().Pods(targetNamespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod %s: %w", podName, err)
	}

	// Nếu không chỉ định container và pod có nhiều containers
	if logsContainerName == "" && len(pod.Spec.Containers) > 1 {
		fmt.Println("Pod has multiple containers:")
		for i, container := range pod.Spec.Containers {
			fmt.Printf("  %d. %s\n", i+1, container.Name)
		}
		return fmt.Errorf("please specify container with -c flag")
	}

	// Sử dụng container đầu tiên nếu không chỉ định
	if logsContainerName == "" {
		logsContainerName = pod.Spec.Containers[0].Name
	}

	// Thiết lập options cho logs
	logOptions := &corev1.PodLogOptions{
		Container:  logsContainerName,
		Follow:     logsFollow,
		Timestamps: logsTimestamps,
	}

	if logsTailLines > 0 {
		logOptions.TailLines = &logsTailLines
	}

	if logsSinceSeconds > 0 {
		logOptions.SinceSeconds = &logsSinceSeconds
	}

	// Lấy logs stream
	req := client.Clientset.CoreV1().Pods(targetNamespace).GetLogs(podName, logOptions)
	stream, err := req.Stream(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get logs stream: %w", err)
	}
	defer stream.Close()

	// Đọc và hiển thị logs
	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading logs: %w", err)
		}

		// Xử lý và hiển thị line
		line = strings.TrimSuffix(line, "\n")
		if logsContainerName != "" && len(pod.Spec.Containers) > 1 {
			fmt.Printf("[%s] %s\n", logsContainerName, line)
		} else {
			fmt.Println(line)
		}
	}

	return nil
}

// init khởi tạo cấu hình cho kube-logs command
func init() {
	// Định nghĩa flags
	logsRootCmd.Flags().StringVarP(&logsNamespace, "namespace", "n", "", "Kubernetes namespace to use")
	logsRootCmd.Flags().StringVarP(&logsKubeContext, "context", "c", "", "Kubernetes context to use")
	logsRootCmd.Flags().BoolVarP(&logsFollow, "follow", "f", false, "Follow logs output (real-time)")
	logsRootCmd.Flags().Int64VarP(&logsTailLines, "tail", "t", 0, "Number of lines to show from the end of the logs")
	logsRootCmd.Flags().Int64Var(&logsSinceSeconds, "since", 0, "Show logs since this many seconds ago")
	logsRootCmd.Flags().StringVar(&logsContainerName, "container", "", "Container name (required if pod has multiple containers)")
	logsRootCmd.Flags().BoolVar(&logsTimestamps, "timestamps", false, "Include timestamps in output")

	// Bind flags với viper
	viper.BindPFlag("namespace", logsRootCmd.Flags().Lookup("namespace"))
	viper.BindPFlag("context", logsRootCmd.Flags().Lookup("context"))
}

// main là entry point của kube-logs
func main() {
	if err := logsRootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
