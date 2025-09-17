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
	podsNamespace     string
	podsContext       string
	podsAllNamespaces bool
)

// podsRootCmd represents the kube-pods command
var podsRootCmd = &cobra.Command{
	Use:   "kube-pods",
	Short: "List pods",
	Long: `kube-pods lists pods in your Kubernetes cluster with a clean table output.

It is similar to 'kubectl get pods' but adds colored status, IP, node and image versions columns.`,
	RunE: runPods,
}

// runPods executes the logic to list pods
func runPods(cmd *cobra.Command, args []string) error {
	client, err := k8s.NewClient("", podsContext)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	targetNamespace := podsNamespace
	if targetNamespace == "" {
		// Get current namespace from kubeconfig when --namespace is missing
		ns, err := k8s.GetCurrentNamespace(podsContext)
		if err != nil {
			return fmt.Errorf("failed to get current namespace: %w", err)
		}
		targetNamespace = ns
	}

	// If --all-namespaces, list across all namespaces
	if podsAllNamespaces {
		targetNamespace = ""
	}

	pods, err := client.Clientset.CoreV1().Pods(targetNamespace).List(client.Context, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list pods: %w", err)
	}

	// Prepare table data
	var headers []string
	if podsAllNamespaces {
		headers = []string{"NAMESPACE", "NAME", "READY", "STATUS", "IP", "NODE", "IMAGE-VERSIONS", "RESTARTS", "AGE"}
	} else {
		headers = []string{"NAME", "READY", "STATUS", "IP", "NODE", "IMAGE-VERSIONS", "RESTARTS", "AGE"}
	}

	var rows [][]string
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
		ip := pod.Status.PodIP
		statusColored := colorStatus(string(pod.Status.Phase))
		node := pod.Spec.NodeName
		// Aggregate image versions from containers (including initContainers)
		versionSet := map[string]struct{}{}
		for _, c := range pod.Spec.Containers {
			versionSet[extractImageVersion(c.Image)] = struct{}{}
		}
		for _, c := range pod.Spec.InitContainers {
			versionSet[extractImageVersion(c.Image)] = struct{}{}
		}
		versions := make([]string, 0, len(versionSet))
		for v := range versionSet {
			versions = append(versions, v)
		}
		versionsStr := strings.Join(versions, ",")
		versionsStr = utils.TruncateString(versionsStr, 60)

		if podsAllNamespaces {
			rows = append(rows, []string{
				pod.Namespace,
				pod.Name,
				fmt.Sprintf("%d/%d", ready, total),
				statusColored,
				ip,
				node,
				versionsStr,
				fmt.Sprintf("%d", restarts),
				utils.FormatAge(age),
			})
		} else {
			rows = append(rows, []string{
				pod.Name,
				fmt.Sprintf("%d/%d", ready, total),
				statusColored,
				ip,
				node,
				versionsStr,
				fmt.Sprintf("%d", restarts),
				utils.FormatAge(age),
			})
		}
	}

	renderTable(headers, rows)
	return nil
}

// init initializes flags for kube-pods command
func init() {
	// Define flags
	podsRootCmd.Flags().StringVarP(&podsNamespace, "namespace", "n", "", "Kubernetes namespace to use")
	podsRootCmd.Flags().StringVarP(&podsContext, "context", "c", "", "Kubernetes context to use")
	podsRootCmd.Flags().BoolVarP(&podsAllNamespaces, "all-namespaces", "A", false, "Show pods from all namespaces")

	// Bind flags with viper
	viper.BindPFlag("namespace", podsRootCmd.Flags().Lookup("namespace"))
	viper.BindPFlag("context", podsRootCmd.Flags().Lookup("context"))
}

// renderTable prints an ASCII table with simple borders
// headers: column headers, rows: row data
func renderTable(headers []string, rows [][]string) {
	widths := make([]int, len(headers))
	// Calculate width based on content (excluding ANSI color codes)
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

// extractImageVersion extracts the version part (tag or shortened digest) from image name
// Examples:
// - nginx:1.25 -> 1.25
// - gcr.io/app/backend@sha256:abcd... -> sha256:abcd
// - busybox -> latest
func extractImageVersion(image string) string {
	// Prefer tag if available
	if i := strings.LastIndex(image, ":"); i != -1 && i > strings.LastIndex(image, "/") {
		return image[i+1:]
	}
	// If digest exists
	if i := strings.Index(image, "@sha256:"); i != -1 {
		digest := image[i+1:] // sha256:...
		if len(digest) > 17 { // sha256: + 12 hex = 19, shorten a bit
			return digest[:17]
		}
		return digest
	}
	return "latest"
}

// colorStatus colors STATUS text by phase for easy identification
// - Running: green
// - Pending: yellow
// - Succeeded: light blue
// - Failed: red
// - Unknown: gray
func colorStatus(phase string) string {
	const (
		reset  = "\033[0m"
		green  = "\033[32m"
		yellow = "\033[33m"
		red    = "\033[31m"
		blue   = "\033[36m"
		gray   = "\033[90m"
	)
	switch phase {
	case "Running":
		return green + phase + reset
	case "Pending":
		return yellow + phase + reset
	case "Failed":
		return red + phase + reset
	case "Succeeded":
		return blue + phase + reset
	default:
		return gray + phase + reset
	}
}

// main is the entry point of kube-pods
func main() {
	if err := podsRootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
