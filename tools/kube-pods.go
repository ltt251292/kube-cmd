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

// rootCmd đại diện cho kube-pods command
var rootCmd = &cobra.Command{
	Use:   "kube-pods",
	Short: "Hiển thị danh sách pods",
	Long: `kube-pods là một tool để xem danh sách pods trong Kubernetes cluster.
	
Tương đương với kubectl get pods nhưng với format đẹp hơn và các tùy chọn thuận tiện.`,
	RunE: runPods,
}

// runPods thực thi logic lấy danh sách pods
func runPods(cmd *cobra.Command, args []string) error {
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

	pods, err := client.Clientset.CoreV1().Pods(targetNamespace).List(client.Context, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	// Hiển thị kết quả dạng bảng
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)

	if allNamespaces {
		fmt.Fprintln(w, "NAMESPACE\tNAME\tREADY\tSTATUS\tRESTARTS\tAGE")
	} else {
		fmt.Fprintln(w, "NAME\tREADY\tSTATUS\tRESTARTS\tAGE")
	}

	for _, pod := range pods.Items {
		ready := 0
		total := len(pod.Spec.Containers)
		for _, status := range pod.Status.ContainerStatuses {
			if status.Ready {
				ready++
			}
		}

		restarts := int32(0)
		for _, status := range pod.Status.ContainerStatuses {
			restarts += status.RestartCount
		}

		age := metav1.Now().Time.Sub(pod.CreationTimestamp.Time)

		if allNamespaces {
			fmt.Fprintf(w, "%s\t%s\t%d/%d\t%s\t%d\t%s\n",
				pod.Namespace,
				pod.Name,
				ready,
				total,
				pod.Status.Phase,
				restarts,
				utils.FormatAge(age),
			)
		} else {
			fmt.Fprintf(w, "%s\t%d/%d\t%s\t%d\t%s\n",
				pod.Name,
				ready,
				total,
				pod.Status.Phase,
				restarts,
				utils.FormatAge(age),
			)
		}
	}

	w.Flush()
	return nil
}

// init khởi tạo cấu hình cho kube-pods command
func init() {
	// Định nghĩa flags
	rootCmd.Flags().StringVarP(&namespace, "namespace", "n", "", "Kubernetes namespace to use")
	rootCmd.Flags().StringVarP(&context, "context", "c", "", "Kubernetes context to use")
	rootCmd.Flags().BoolVarP(&allNamespaces, "all-namespaces", "A", false, "Hiển thị pods từ tất cả namespaces")

	// Bind flags với viper
	viper.BindPFlag("namespace", rootCmd.Flags().Lookup("namespace"))
	viper.BindPFlag("context", rootCmd.Flags().Lookup("context"))
}

// main là entry point của kube-pods
func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
