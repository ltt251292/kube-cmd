package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"kube/pkg/kubernetes/k8s"
	"kube/pkg/shared/utils"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	servicesNamespace     string
	servicesContext       string
	servicesAllNamespaces bool
)

// servicesRootCmd represents the kube-services command
var servicesRootCmd = &cobra.Command{
	Use:   "kube-services",
	Short: "List services",
	Long:  `kube-services lists services in your Kubernetes cluster with a clean table output.`,
	RunE:  runServices,
}

// runServices executes the logic to list services
func runServices(cmd *cobra.Command, args []string) error {
	client, err := k8s.NewClient("", servicesContext)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	targetNamespace := servicesNamespace
	if targetNamespace == "" {
		// Get current namespace from kubeconfig if no --namespace flag
		ns, err := k8s.GetCurrentNamespace(servicesContext)
		if err != nil {
			return fmt.Errorf("failed to get current namespace: %w", err)
		}
		targetNamespace = ns
	}

	// If --all-namespaces, get from all namespaces
	if servicesAllNamespaces {
		targetNamespace = ""
	}

	services, err := client.Clientset.CoreV1().Services(targetNamespace).List(client.Context, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list services: %w", err)
	}

	// Prepare table data
	var headers []string
	if servicesAllNamespaces {
		headers = []string{"NAMESPACE", "NAME", "TYPE", "CLUSTER-IP", "EXTERNAL-IP", "PORT(S)", "AGE"}
	} else {
		headers = []string{"NAME", "TYPE", "CLUSTER-IP", "EXTERNAL-IP", "PORT(S)", "AGE"}
	}

	var rows [][]string
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

		if servicesAllNamespaces {
			rows = append(rows, []string{
				svc.Namespace,
				svc.Name,
				string(svc.Spec.Type),
				svc.Spec.ClusterIP,
				externalIP,
				ports,
				utils.FormatAge(age),
			})
		} else {
			rows = append(rows, []string{
				svc.Name,
				string(svc.Spec.Type),
				svc.Spec.ClusterIP,
				externalIP,
				ports,
				utils.FormatAge(age),
			})
		}
	}

	renderTable(headers, rows)
	return nil
}

// init initializes flags for kube-services command
func init() {
	// Define flags
	servicesRootCmd.Flags().StringVarP(&servicesNamespace, "namespace", "n", "", "Kubernetes namespace to use")
	servicesRootCmd.Flags().StringVarP(&servicesContext, "context", "c", "", "Kubernetes context to use")
	servicesRootCmd.Flags().BoolVarP(&servicesAllNamespaces, "all-namespaces", "A", false, "Show services from all namespaces")

	// Bind flags with viper
	viper.BindPFlag("namespace", servicesRootCmd.Flags().Lookup("namespace"))
	viper.BindPFlag("context", servicesRootCmd.Flags().Lookup("context"))
}

// renderTable prints an ASCII table with simple borders
// headers: column headers, rows: row data
func renderTable(headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	// Calculate width based on content (excluding ANSI color codes if any)
	for c, h := range headers {
		w := displayWidth(h)
		if w > widths[c] {
			widths[c] = w
		}
	}
	for _, row := range rows {
		for c, cell := range row {
			w := displayWidth(cell)
			if w > widths[c] {
				widths[c] = w
			}
		}
	}

	printSeparator(widths)
	fmt.Println("| " + joinRow(headers, widths) + " |")
	printSeparator(widths)
	for _, row := range rows {
		fmt.Println("| " + joinRow(row, widths) + " |")
	}
	printSeparator(widths)
}

// displayWidth returns display length (excluding ANSI codes)
func displayWidth(s string) int {
	return len(stripANSI(s))
}

// stripANSI removes ANSI color codes for accurate width calculation
func stripANSI(s string) string {
	ansi := regexp.MustCompile("\\x1b\\[[0-9;]*m")
	return ansi.ReplaceAllString(s, "")
}

// joinRow left-aligns each cell and joins with column separator
func joinRow(cols []string, widths []int) string {
	parts := make([]string, len(cols))
	for i, col := range cols {
		pad := widths[i] - displayWidth(col)
		if pad < 0 {
			pad = 0
		}
		parts[i] = col + strings.Repeat(" ", pad)
	}
	return strings.Join(parts, " | ")
}

// printSeparator prints border line based on column widths
func printSeparator(widths []int) {
	b := strings.Builder{}
	b.WriteString("+")
	for i, w := range widths {
		b.WriteString(strings.Repeat("-", w+2))
		if i == len(widths)-1 {
			b.WriteString("+")
		} else {
			b.WriteString("+")
		}
	}
	fmt.Println(b.String())
}

// main is the entry point of kube-services
func main() {
	if err := servicesRootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
