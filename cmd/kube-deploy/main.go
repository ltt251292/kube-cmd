package main

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"kube/pkg/kubernetes/k8s"
	"kube/pkg/shared/utils"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	deployNamespace   string
	deployKubeContext string
)

var deployRootCmd = &cobra.Command{
	Use:   "kube-deploy [deployment] --image <image[:tag]>",
	Short: "Update Deployment image and wait for rollout, or list Deployments",
	Long: `kube-deploy can:

- List Deployments in the current namespace (when no deployment is provided)
- Update image for all containers in a Deployment and wait for rollout to complete

Tips:
- Use --namespace/-n to target a namespace
- Use --context/-c to select kube context`,
	Example: `
  # List deployments in current namespace
  kube-deploy

  # List deployments in namespace my-app
  kube-deploy -n my-app

  # Update image for deployment backend and wait for rollout
  kube-deploy backend --image repo/backend:1.2.3
`,
	Args: cobra.RangeArgs(0, 1),
	RunE: runDeploy,
}

// runDeploy updates image for all containers in the Deployment
func runDeploy(cmd *cobra.Command, args []string) error {
	image, _ := cmd.Flags().GetString("image")

	client, err := k8s.NewClient("", deployKubeContext)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ns := deployNamespace
	if ns == "" {
		if ns, err = k8s.GetCurrentNamespace(deployKubeContext); err != nil {
			return fmt.Errorf("failed to get current namespace: %w", err)
		}
	}

	// If no deployment is provided => list deployments
	if len(args) == 0 {
		return listDeployments(context.Background(), client, ns)
	}

	deploymentName := args[0]

	if strings.TrimSpace(image) == "" {
		return fmt.Errorf("--image is required when specifying a deployment")
	}

	// Get current Deployment
	dep, err := client.Clientset.AppsV1().Deployments(ns).Get(context.Background(), deploymentName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get deployment %s: %w", deploymentName, err)
	}

	// Update image for all containers
	for i := range dep.Spec.Template.Spec.Containers {
		dep.Spec.Template.Spec.Containers[i].Image = image
	}

	// Apply update
	if _, err := client.Clientset.AppsV1().Deployments(ns).Update(context.Background(), dep, metav1.UpdateOptions{}); err != nil {
		return fmt.Errorf("failed to update deployment: %w", err)
	}

	fmt.Printf("Updated deployment %s image to %s. Waiting for rollout...\n", deploymentName, image)

	// Wait for rollout to complete
	if err := waitForDeploymentRollout(context.Background(), client, ns, deploymentName); err != nil {
		return err
	}

	fmt.Println("Rollout completed")
	return nil
}

func init() {
	deployRootCmd.Flags().StringVarP(&deployNamespace, "namespace", "n", "", "Kubernetes namespace to use")
	deployRootCmd.Flags().StringVarP(&deployKubeContext, "context", "c", "", "Kubernetes context to use")
	deployRootCmd.Flags().String("image", "", "Container image to set (e.g. repo/app:tag)")
}

// main is the entry point of kube-deploy
func main() {
	if err := deployRootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// waitForDeploymentRollout waits until deployment available replicas == desired
func waitForDeploymentRollout(ctx context.Context, client *k8s.Client, ns, name string) error {
	// Simple polling with light backoff
	for i := 0; i < 180; i++ { // max ~3 minutes (i * 1s)
		dep, err := client.Clientset.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get deployment during rollout: %w", err)
		}
		if dep.Status.UpdatedReplicas == *dep.Spec.Replicas &&
			dep.Status.ReadyReplicas == *dep.Spec.Replicas &&
			dep.Status.AvailableReplicas == *dep.Spec.Replicas &&
			dep.Status.ObservedGeneration >= dep.Generation {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("timeout waiting for rollout of deployment %s", name)
}

// listDeployments displays a table of Deployments in the namespace
func listDeployments(ctx context.Context, client *k8s.Client, ns string) error {
	list, err := client.Clientset.AppsV1().Deployments(ns).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	headers := []string{"NAME", "READY", "UP-TO-DATE", "AVAILABLE", "AGE"}
	var rows [][]string
	for _, dep := range list.Items {
		desired := int32(1)
		if dep.Spec.Replicas != nil {
			desired = *dep.Spec.Replicas
		}
		ready := dep.Status.ReadyReplicas
		upToDate := dep.Status.UpdatedReplicas
		available := dep.Status.AvailableReplicas
		age := time.Since(dep.CreationTimestamp.Time)

		rows = append(rows, []string{
			dep.Name,
			fmt.Sprintf("%d/%d", ready, desired),
			fmt.Sprintf("%d", upToDate),
			fmt.Sprintf("%d", available),
			utils.FormatAge(age),
		})
	}

	renderTable(headers, rows)
	return nil
}

// Table helpers (similar to other commands)
func renderTable(headers []string, rows [][]string) {
	widths := make([]int, len(headers))
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

func displayWidth(s string) int {
	return len(stripANSI(s))
}

func stripANSI(s string) string {
	ansi := regexp.MustCompile("\\x1b\\[[0-9;]*m")
	return ansi.ReplaceAllString(s, "")
}

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
