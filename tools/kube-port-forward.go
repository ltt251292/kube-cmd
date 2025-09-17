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

	"kube/pkg/k8s"

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

// portForwardRootCmd đại diện cho kube-port-forward command
var portForwardRootCmd = &cobra.Command{
	Use:   "kube-port-forward [pod-name] [local-port]:[remote-port]",
	Short: "Forward một local port đến port của pod",
	Long: `kube-port-forward tạo tunnel từ local port đến port của pod trong cluster.
	
Format port: [local-port]:[remote-port]
Nếu chỉ có một port, sẽ sử dụng cùng port cho cả local và remote.

Ví dụ:
  kube-port-forward my-pod 8080:80       # Forward local 8080 -> pod 80
  kube-port-forward my-pod 3000          # Forward local 3000 -> pod 3000`,
	Args: cobra.ExactArgs(2),
	RunE: runPortForward,
}

// runPortForward thực thi logic port forwarding
func runPortForward(cmd *cobra.Command, args []string) error {
	podName := args[0]
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
		targetNamespace = "default"
	}

	// Kiểm tra pod có tồn tại không
	_, err = client.Clientset.CoreV1().Pods(targetNamespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get pod %s: %w", podName, err)
	}

	// Tạo URL cho port forward request
	url := client.Clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Namespace(targetNamespace).
		Name(podName).
		SubResource("portforward").URL()

	// Tạo SPDY transport
	transport, upgrader, err := spdy.RoundTripperFor(client.Config)
	if err != nil {
		return fmt.Errorf("failed to create SPDY transport: %w", err)
	}

	// Tạo dialer
	dialer := spdy.NewDialer(upgrader, &http.Client{Transport: transport}, "POST", url)

	// Thiết lập stop channel và ready channel
	stopCh := make(chan struct{}, 1)
	readyCh := make(chan struct{})

	// Thiết lập signal handling
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-signalCh
		fmt.Println("\nStopping port forward...")
		close(stopCh)
	}()

	// Tạo port forwarder
	ports := []string{fmt.Sprintf("%d:%d", localPort, remotePort)}
	pf, err := portforward.New(dialer, ports, stopCh, readyCh, os.Stdout, os.Stderr)
	if err != nil {
		return fmt.Errorf("failed to create port forwarder: %w", err)
	}

	// Bắt đầu port forwarding trong goroutine
	go func() {
		if err := pf.ForwardPorts(); err != nil {
			fmt.Fprintf(os.Stderr, "Port forwarding error: %v\n", err)
		}
	}()

	// Đợi ready signal
	<-readyCh

	fmt.Printf("Forwarding from 127.0.0.1:%d -> %s:%d\n", localPort, podName, remotePort)
	fmt.Printf("Press Ctrl+C to stop\n")

	// Đợi stop signal
	<-stopCh

	return nil
}

// parsePortSpec phân tích port specification
// Hỗ trợ format: port, local:remote
func parsePortSpec(spec string) (int, int, error) {
	parts := strings.Split(spec, ":")

	switch len(parts) {
	case 1:
		// Chỉ có một port, sử dụng cho cả local và remote
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

// init khởi tạo cấu hình cho kube-port-forward command
func init() {
	// Định nghĩa flags
	portForwardRootCmd.Flags().StringVarP(&portForwardNamespace, "namespace", "n", "", "Kubernetes namespace to use")
	portForwardRootCmd.Flags().StringVarP(&portForwardKubeContext, "context", "c", "", "Kubernetes context to use")

	// Bind flags với viper
	viper.BindPFlag("namespace", portForwardRootCmd.Flags().Lookup("namespace"))
	viper.BindPFlag("context", portForwardRootCmd.Flags().Lookup("context"))
}

// main là entry point của kube-port-forward
func main() {
	if err := portForwardRootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
