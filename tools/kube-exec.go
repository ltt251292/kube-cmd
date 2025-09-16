package main

import (
	"context"
	"fmt"
	"os"

	"kube/pkg/k8s"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

var (
	namespace     string
	kubeContext   string
	execContainer string
	execTty       bool
	execStdin     bool
)

// rootCmd đại diện cho kube-exec command
var rootCmd = &cobra.Command{
	Use:   "kube-exec [pod-name] -- [command...]",
	Short: "Thực thi command trong pod",
	Long: `kube-exec cho phép thực thi command bên trong container của pod.
	
Ví dụ:
  kube-exec my-pod -- bash                       # Mở bash shell
  kube-exec my-pod -- ls -la /app                # Thực thi command cụ thể
  kube-exec my-pod -c container-name -- env      # Exec vào container cụ thể`,
	Args: cobra.MinimumNArgs(1),
	RunE: runExec,
}

// runExec thực thi logic exec command
func runExec(cmd *cobra.Command, args []string) error {
	if len(args) < 3 || args[1] != "--" {
		return fmt.Errorf("invalid syntax. Use: kube-exec [pod-name] -- [command...]")
	}

	podName := args[0]
	command := args[2:]

	client, err := k8s.NewClient("", kubeContext)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	targetNamespace := namespace
	if targetNamespace == "" {
		targetNamespace = "default"
	}

	// Lấy thông tin pod để kiểm tra containers
	pod, err := client.Clientset.CoreV1().Pods(targetNamespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod %s: %w", podName, err)
	}

	// Nếu không chỉ định container và pod có nhiều containers
	if execContainer == "" && len(pod.Spec.Containers) > 1 {
		fmt.Println("Pod has multiple containers:")
		for i, container := range pod.Spec.Containers {
			fmt.Printf("  %d. %s\n", i+1, container.Name)
		}
		return fmt.Errorf("please specify container with -c flag")
	}

	// Sử dụng container đầu tiên nếu không chỉ định
	if execContainer == "" {
		execContainer = pod.Spec.Containers[0].Name
	}

	// Thiết lập exec request
	req := client.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(targetNamespace).
		SubResource("exec")

	req.VersionedParams(&corev1.PodExecOptions{
		Container: execContainer,
		Command:   command,
		Stdin:     execStdin,
		Stdout:    true,
		Stderr:    true,
		TTY:       execTty,
	}, scheme.ParameterCodec)

	// Tạo executor
	executor, err := remotecommand.NewSPDYExecutor(client.Config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	// Thực thi command
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:  os.Stdin,
		Stdout: os.Stdout,
		Stderr: os.Stderr,
		Tty:    execTty,
	})
	if err != nil {
		return fmt.Errorf("failed to execute command: %w", err)
	}

	return nil
}

// init khởi tạo cấu hình cho kube-exec command
func init() {
	// Định nghĩa flags
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace to use")
	rootCmd.Flags().StringVarP(&kubeContext, "context", "c", "", "Kubernetes context to use")
	rootCmd.Flags().StringVar(&execContainer, "container", "", "Container name (required if pod has multiple containers)")
	rootCmd.Flags().BoolVarP(&execTty, "tty", "t", true, "Allocate a TTY")
	rootCmd.Flags().BoolVarP(&execStdin, "stdin", "i", true, "Keep STDIN open")

	// Bind flags với viper
	viper.BindPFlag("namespace", rootCmd.Flags().Lookup("namespace"))
	viper.BindPFlag("context", rootCmd.Flags().Lookup("context"))
}

// main là entry point của kube-exec
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
