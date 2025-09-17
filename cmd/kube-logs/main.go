package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"kube/pkg/kubernetes/k8s"

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

// logsRootCmd represents the kube-logs command
var logsRootCmd = &cobra.Command{
	Use:   "kube-logs [pod-name]",
	Short: "Show pod logs",
	Long: `kube-logs shows logs for a specific pod.

Features:
- Follow logs in real-time (-f)
- Show last N lines (-t, --tail)
- Show logs since seconds ago (--since)
- Select a specific container (-c, --container)
- Include timestamps (--timestamps)

Examples:
  kube-logs my-pod                       # Show logs of a pod
  kube-logs my-pod -f                    # Follow logs in real-time
  kube-logs my-pod -c container-name     # Logs for a specific container`,
	Args: cobra.ExactArgs(1),
	RunE: runLogs,
}

// runLogs executes the logic to display logs
func runLogs(cmd *cobra.Command, args []string) error {
	podName := args[0]

	client, err := k8s.NewClient("", logsKubeContext)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	targetNamespace := logsNamespace
	if targetNamespace == "" {
		// Get current namespace from kubeconfig if no --namespace flag
		ns, err := k8s.GetCurrentNamespace(logsKubeContext)
		if err != nil {
			return fmt.Errorf("failed to get current namespace: %w", err)
		}
		targetNamespace = ns
	}

	// Get pod information to check containers
	pod, err := client.Clientset.CoreV1().Pods(targetNamespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod %s: %w", podName, err)
	}

	// If no container is specified and pod has multiple containers
	if logsContainerName == "" && len(pod.Spec.Containers) > 1 {
		fmt.Println("Pod has multiple containers:")
		for i, container := range pod.Spec.Containers {
			fmt.Printf("  %d. %s\n", i+1, container.Name)
		}
		return fmt.Errorf("please specify container with -c flag")
	}

	// Use first container if not specified
	if logsContainerName == "" {
		logsContainerName = pod.Spec.Containers[0].Name
	}

	// Set up options for logs
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

	// Get logs stream
	req := client.Clientset.CoreV1().Pods(targetNamespace).GetLogs(podName, logOptions)
	stream, err := req.Stream(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get logs stream: %w", err)
	}
	defer stream.Close()

	// Read and display logs
	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading logs: %w", err)
		}

		// Process and display line
		line = strings.TrimSuffix(line, "\n")
		if logsContainerName != "" && len(pod.Spec.Containers) > 1 {
			fmt.Printf("[%s] %s\n", logsContainerName, line)
		} else {
			fmt.Println(line)
		}
	}

	return nil
}

// init initializes configuration for kube-logs command
func init() {
	// Define flags
	logsRootCmd.Flags().StringVarP(&logsNamespace, "namespace", "n", "", "Kubernetes namespace to use")
	logsRootCmd.Flags().StringVarP(&logsKubeContext, "context", "c", "", "Kubernetes context to use")
	logsRootCmd.Flags().BoolVarP(&logsFollow, "follow", "f", true, "Follow logs output (real-time)")
	logsRootCmd.Flags().Int64VarP(&logsTailLines, "tail", "t", 0, "Number of lines to show from the end of the logs")
	logsRootCmd.Flags().Int64Var(&logsSinceSeconds, "since", 0, "Show logs since this many seconds ago")
	logsRootCmd.Flags().StringVar(&logsContainerName, "container", "", "Container name (required if pod has multiple containers)")
	logsRootCmd.Flags().BoolVar(&logsTimestamps, "timestamps", false, "Include timestamps in output")

	// Bind flags with viper
	viper.BindPFlag("namespace", logsRootCmd.Flags().Lookup("namespace"))
	viper.BindPFlag("context", logsRootCmd.Flags().Lookup("context"))
}

// main is the entry point of kube-logs
func main() {
	if err := logsRootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
