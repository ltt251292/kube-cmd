package main

import (
	"context"
	"fmt"
	"os"

	"kube/pkg/kubernetes/k8s"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

var (
	execNamespace   string
	execKubeContext string
	execContainer   string
	execTty         bool
	execStdin       bool
)

// execRootCmd represents the kube-exec command
var execRootCmd = &cobra.Command{
	Use:   "kube-exec [pod-name] -- [command...]",
	Short: "Execute command in pod",
	Long: `kube-exec allows executing commands inside a pod's container.
	
Examples:
  kube-exec my-pod -- bash                       # Open bash shell
  kube-exec my-pod -- ls -la /app                # Execute specific command
  kube-exec my-pod -c container-name -- env      # Exec into specific container`,
	Args: cobra.MinimumNArgs(1),
	RunE: runExec,
}

// runExec executes the exec command logic
// It uses Cobra's ArgsLenAtDash to split arguments at "--":
//   - args[:dash] are positional args before "--" (expects: pod name)
//   - args[dash+1:] are the command and its arguments to run inside the pod
func runExec(cmd *cobra.Command, args []string) error {
	dashIndex := cmd.ArgsLenAtDash()

	if dashIndex == -1 {
		return fmt.Errorf("invalid syntax. Use: kube-exec [pod-name] -- [command...]")
	}
	if dashIndex < 1 {
		return fmt.Errorf("pod name is required before --")
	}

	podName := args[0]
	command := args[dashIndex:]

	client, err := k8s.NewClient("", execKubeContext)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	targetNamespace := execNamespace
	if targetNamespace == "" {
		// Get current namespace from kubeconfig if no --namespace flag
		ns, err := k8s.GetCurrentNamespace(execKubeContext)
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
	if execContainer == "" && len(pod.Spec.Containers) > 1 {
		fmt.Println("Pod has multiple containers:")
		for i, container := range pod.Spec.Containers {
			fmt.Printf("  %d. %s\n", i+1, container.Name)
		}
		return fmt.Errorf("please specify container with -c flag")
	}

	// Use first container if not specified
	if execContainer == "" {
		execContainer = pod.Spec.Containers[0].Name
	}

	// Set up exec request
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

	// Create executor
	executor, err := remotecommand.NewSPDYExecutor(client.Config, "POST", req.URL())
	if err != nil {
		return fmt.Errorf("failed to create executor: %w", err)
	}

	// Execute command
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

// init initializes configuration for kube-exec command
func init() {
	// Define flags
	execRootCmd.Flags().StringVarP(&execNamespace, "namespace", "n", "", "Kubernetes namespace to use")
	execRootCmd.Flags().StringVarP(&execKubeContext, "context", "c", "", "Kubernetes context to use")
	execRootCmd.Flags().StringVar(&execContainer, "container", "", "Container name (required if pod has multiple containers)")
	execRootCmd.Flags().BoolVarP(&execTty, "tty", "t", true, "Allocate a TTY")
	execRootCmd.Flags().BoolVarP(&execStdin, "stdin", "i", true, "Keep STDIN open")

	// Bind flags with viper
	viper.BindPFlag("namespace", execRootCmd.Flags().Lookup("namespace"))
	viper.BindPFlag("context", execRootCmd.Flags().Lookup("context"))
}

// main is the entry point of kube-exec
func main() {
	if err := execRootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
