package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"kube/pkg/kubernetes/k8s"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	rolloutNamespace   string
	rolloutKubeContext string
)

var rolloutRootCmd = &cobra.Command{
	Use:   "kube-rollout <deployment> [--restart]",
	Short: "Show rollout status or restart a Deployment",
	Long: `kube-rollout can:

- Restart a Deployment by touching the restartedAt annotation
- Wait for rollout to complete, or just print current status once

Tips:
- Use --namespace/-n to target a namespace
- Use --context/-c to select kube context`,
	Example: `
  # Show rollout status once
  kube-rollout backend -n my-ns --restart=false

  # Restart a deployment then wait for rollout to complete
  kube-rollout backend -n my-ns --restart
`,
	Args: cobra.ExactArgs(1),
	RunE: runRollout,
}

func runRollout(cmd *cobra.Command, args []string) error {
	deploymentName := args[0]
	doRestart, _ := cmd.Flags().GetBool("restart")

	client, err := k8s.NewClient("", rolloutKubeContext)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	ns := rolloutNamespace
	if ns == "" {
		if ns, err = k8s.GetCurrentNamespace(rolloutKubeContext); err != nil {
			return fmt.Errorf("failed to get current namespace: %w", err)
		}
	}

	if doRestart {
		// Restart by touching annotation to trigger a new rollout
		dep, err := client.Clientset.AppsV1().Deployments(ns).Get(context.Background(), deploymentName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get deployment %s: %w", deploymentName, err)
		}
		if dep.Spec.Template.ObjectMeta.Annotations == nil {
			dep.Spec.Template.ObjectMeta.Annotations = map[string]string{}
		}
		dep.Spec.Template.ObjectMeta.Annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339)
		if _, err := client.Clientset.AppsV1().Deployments(ns).Update(context.Background(), dep, metav1.UpdateOptions{}); err != nil {
			return fmt.Errorf("failed to update deployment: %w", err)
		}
		fmt.Println("Deployment restarted. Waiting for rollout...")
	}

	// Wait for rollout to complete or just print current status
	for i := 0; i < 180; i++ {
		dep, err := client.Clientset.AppsV1().Deployments(ns).Get(context.Background(), deploymentName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get deployment: %w", err)
		}
		fmt.Printf("ObservedGeneration=%d/%d Updated=%d Ready=%d Available=%d Desired=%d\n",
			dep.Status.ObservedGeneration, dep.Generation,
			dep.Status.UpdatedReplicas,
			dep.Status.ReadyReplicas,
			dep.Status.AvailableReplicas,
			*dep.Spec.Replicas,
		)
		if dep.Status.UpdatedReplicas == *dep.Spec.Replicas &&
			dep.Status.ReadyReplicas == *dep.Spec.Replicas &&
			dep.Status.AvailableReplicas == *dep.Spec.Replicas &&
			dep.Status.ObservedGeneration >= dep.Generation {
			fmt.Println("Rollout is complete")
			return nil
		}
		time.Sleep(1 * time.Second)
		if !doRestart { // status-only: print once and exit
			return nil
		}
	}
	return fmt.Errorf("timeout waiting for rollout of deployment %s", deploymentName)
}

func init() {
	rolloutRootCmd.Flags().StringVarP(&rolloutNamespace, "namespace", "n", "", "Kubernetes namespace to use")
	rolloutRootCmd.Flags().StringVarP(&rolloutKubeContext, "context", "c", "", "Kubernetes context to use")
	rolloutRootCmd.Flags().BoolVar(&rolloutRestart, "restart", true, "Restart the deployment before waiting for rollout")
}

var rolloutRestart bool

// main is the entry point of kube-rollout
func main() {
	if err := rolloutRootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
