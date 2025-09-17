package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"kube/pkg/kubernetes/k8s"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/portforward"
	"k8s.io/client-go/transport/spdy"
)

var (
	portForwardNamespace   string
	portForwardKubeContext string
)

var portForwardRootCmd = &cobra.Command{
	Use:   "kube-port-forward [pod-name|svc/<service-name>] [local-port]:[remote-port]",
	Short: "Port-forward a local port to a pod (or service)",
	Long: `kube-port-forward creates a tunnel from a local port to a pod in the cluster.
    
You can target a pod directly or a service via svc/<service-name>.
When targeting a service, the tool will select a backing pod from the Endpoints of that service.

Port format: [local-port]:[remote-port]
If only one port is provided, it will be used for both local and remote.

Examples:
  kube-port-forward my-pod 8080:80         # Forward local 8080 -> pod 80
  kube-port-forward svc/my-service 3000    # Forward local 3000 -> service 3000`,
	Args: cobra.ExactArgs(2),
	RunE: runPortForward,
}

// runPortForward executes port-forward logic
func runPortForward(cmd *cobra.Command, args []string) error {
	target := args[0]
	portSpec := args[1]

	// Parse port specification
	localPort, remotePort, err := parsePortSpec(portSpec)
	if err != nil {
		return fmt.Errorf("invalid port specification '%s': %w", portSpec, err)
	}

	client, err := k8s.NewClient("", portForwardKubeContext)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	targetNamespace := portForwardNamespace
	if targetNamespace == "" {
		// Get current namespace from kubeconfig if no --namespace flag
		ns, err := k8s.GetCurrentNamespace(portForwardKubeContext)
		if err != nil {
			return fmt.Errorf("failed to get current namespace: %w", err)
		}
		targetNamespace = ns
	}

	// Resolve target pod: direct pod or svc/<name>
	podName, err := resolveTargetPod(context.Background(), client, targetNamespace, target)
	if err != nil {
		return err
	}

	// Create URL for port-forward request
	url := client.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(targetNamespace).
		Name(podName).
		SubResource("portforward").URL()

	// Create SPDY transport
	transport, upgrader, err := spdy.RoundTripperFor(client.Config)
	if err != nil {
		return fmt.Errorf("failed to create SPDY transport: %w", err)
	}

	// Create dialer
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)

	// Setup stop and ready channels
	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{})

	// Setup signal handling
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalCh
		fmt.Println("\nStopping port forward...")
		close(stopCh)
	}()

	// Create port forwarder
	ports := []string{fmt.Sprintf("%d:%d", localPort, remotePort)}
	pf, err := portforward.New(dialer, ports, stopCh, readyCh, os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to create port forwarder: %w", err)
	}

	// Start port-forwarding in a goroutine
	go func() {
		if err := pf.ForwardPorts(); err != nil {
			fmt.Fprintf(os.Stderr, "Port forwarding error: %v\n", err)
		}
	}()

	// Wait for ready signal
	<-readyCh

	fmt.Printf("Forwarding from 127.0.0.1:%d -> %s:%d\n", localPort, podName, remotePort)
	fmt.Printf("Press Ctrl+C to stop\n")

	// Wait for stop signal
	<-stopCh

	return nil
}

// parsePortSpec parses port specification
// Supported formats: port, local:remote
func parsePortSpec(spec string) (int, int, error) {
	parts := strings.Split(spec, ":")

	switch len(parts) {
	case 1:
		// Single port: use for both local and remote
		port, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid port number: %s", parts[0])
		}
		return port, port, nil

	case 2:
		// local:remote format
		localPort, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid local port number: %s", parts[0])
		}

		remotePort, err := strconv.Atoi(parts[1])
		if err != nil {
			return 0, 0, fmt.Errorf("invalid remote port number: %s", parts[1])
		}

		return localPort, remotePort, nil

	default:
		return 0, 0, fmt.Errorf("invalid port specification format")
	}
}

// resolveTargetPod resolves the target to a pod name.
// Supports: "<pod-name>" or "svc/<service-name>" / "service/<service-name>"
func resolveTargetPod(ctx context.Context, client *k8s.Client, namespace string, target string) (string, error) {
	lower := strings.ToLower(target)
	if strings.HasPrefix(lower, "svc/") || strings.HasPrefix(lower, "service/") {
		parts := strings.SplitN(target, "/", 2)
		if len(parts) != 2 || parts[1] == "" {
			return "", fmt.Errorf("invalid service target, expected svc/<name>")
		}
		svcName := parts[1]

		eps, err := client.Clientset.CoreV1().Endpoints(namespace).Get(ctx, svcName, metav1.GetOptions{})
		if err != nil {
			return "", fmt.Errorf("failed to get endpoints for service %s: %w", svcName, err)
		}
		for _, subset := range eps.Subsets {
			for _, addr := range subset.Addresses {
				if addr.TargetRef != nil && addr.TargetRef.Kind == "Pod" && addr.TargetRef.Name != "" {
					return addr.TargetRef.Name, nil
				}
			}
		}
		return "", fmt.Errorf("no backing pod found for service %s", svcName)
	}

	// Default: treat target as pod name; validate existence.
	if _, err := client.Clientset.CoreV1().Pods(namespace).Get(ctx, target, metav1.GetOptions{}); err != nil {
		return "", fmt.Errorf("failed to get pod %s: %w", target, err)
	}
	return target, nil
}

// init initializes configuration for kube-port-forward command
func init() {
	// Define flags
	portForwardRootCmd.Flags().StringVarP(&portForwardNamespace, "namespace", "n", "", "Kubernetes namespace to use")
	portForwardRootCmd.Flags().StringVarP(&portForwardKubeContext, "context", "c", "", "Kubernetes context to use")

	// Bind flags with viper
	viper.BindPFlag("namespace", portForwardRootCmd.Flags().Lookup("namespace"))
	viper.BindPFlag("context", portForwardRootCmd.Flags().Lookup("context"))
}

// main is the entry point of kube-port-forward
func main() {
	if err := portForwardRootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
