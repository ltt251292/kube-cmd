package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"kube/pkg/k8s"
	"kube/pkg/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	namespace     string
	context       string
	allNamespaces bool
)

// rootCmd đại diện cho kube-services command
var rootCmd = &cobra.Command{
	Use:   "kube-services",
	Short: "Hiển thị danh sách services",
	Long: `kube-services là một tool để xem danh sách services trong Kubernetes cluster.
	
Tương đương với kubectl get services nhưng với format đẹp hơn và các tùy chọn thuận tiện.`,
	RunE: runServices,
}

// runServices thực thi logic lấy danh sách services
func runServices(cmd *cobra.Command, args []string) error {
	client, err := k8s.NewClient("", context)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	targetNamespace := namespace
	if targetNamespace == "" {
		targetNamespace = "default"
	}

	// Nếu --all-namespaces, lấy từ tất cả namespaces
	if allNamespaces {
		targetNamespace = ""
	}

	services, err := client.Clientset.CoreV1().Services(targetNamespace).List(client.Context, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	// Hiển thị kết quả dạng bảng
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	if allNamespaces {
		fmt.Fprintln(w, "NAMESPACE\tNAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE")
	} else {
		fmt.Fprintln(w, "NAME\tTYPE\tCLUSTER-IP\tEXTERNAL-IP\tPORT(S)\tAGE")
	}

	for _, svc := range services.Items {
		externalIP := "<none>"
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			if svc.Status.LoadBalancer.Ingress[0].IP != "" {
				externalIP = svc.Status.LoadBalancer.Ingress[0].IP
			} else if svc.Status.LoadBalancer.Ingress[0].Hostname != "" {
				externalIP = svc.Status.LoadBalancer.Ingress[0].Hostname
			}
		}

		ports := ""
		for i, port := range svc.Spec.Ports {
			if i > 0 {
				ports += ","
			}
			if port.NodePort != 0 {
				ports += fmt.Sprintf("%d:%d/%s", port.Port, port.NodePort, port.Protocol)
			} else {
				ports += fmt.Sprintf("%d/%s", port.Port, port.Protocol)
			}
		}

		age := metav1.Now().Time.Sub(svc.CreationTimestamp.Time)

		if allNamespaces {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				svc.Namespace,
				svc.Name,
				svc.Spec.Type,
				svc.Spec.ClusterIP,
				externalIP,
				ports,
				utils.FormatAge(age),
			)
		} else {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n",
				svc.Name,
				svc.Spec.Type,
				svc.Spec.ClusterIP,
				externalIP,
				ports,
				utils.FormatAge(age),
			)
		}
	}

	w.Flush()
	return nil
}

// init khởi tạo cấu hình cho kube-services command
func init() {
	// Định nghĩa flags
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace to use")
	rootCmd.Flags().StringVarP(&context, "context", "c", "", "Kubernetes context to use")
	rootCmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Hiển thị services từ tất cả namespaces")

	// Bind flags với viper
	viper.BindPFlag("namespace", rootCmd.Flags().Lookup("namespace"))
	viper.BindPFlag("context", rootCmd.Flags().Lookup("context"))
}

// main là entry point của kube-services
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
