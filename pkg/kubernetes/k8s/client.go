package k8s

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

// Client wraps Kubernetes client with helper methods
type Client struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
	Context   context.Context
}

// NewClient creates a new Kubernetes client
// Automatically detects configuration from kubeconfig or in-cluster config
func NewClient(kubeconfig string, contextName string) (*Client, error) {
	var config *rest.Config
	var err error

	if kubeconfig == "" {
		// Try to find kubeconfig file at default location
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	// Check if running inside cluster
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		// Use in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
		}
	} else {
		// Use kubeconfig file
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
		}

		// If context name is specified, load config with that context
		if contextName != "" {
			loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
			loadingRules.ExplicitPath = kubeconfig

			configOverrides := &clientcmd.ConfigOverrides{
				CurrentContext: contextName,
			}

			clientConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
				loadingRules, configOverrides)

			config, err = clientConfig.ClientConfig()
			if err != nil {
				return nil, fmt.Errorf("failed to build config with context %s: %w", contextName, err)
			}
		}
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return &Client{
		Clientset: clientset,
		Config:    config,
		Context:   context.Background(),
	}, nil
}

// SetNamespace sets default namespace for client
func (c *Client) SetNamespace(namespace string) {
	// Update context with namespace
	c.Context = context.WithValue(c.Context, "namespace", namespace)
}

// GetCurrentNamespace returns current namespace from kubeconfig for specified context.
// If namespace is not found, returns "default".
func GetCurrentNamespace(contextName string) (string, error) {
	// Determine default kubeconfig path
	kubeconfigPath := ""
	if home := homedir.HomeDir(); home != "" {
		kubeconfigPath = filepath.Join(home, ".kube", "config")
	}

	// Load raw config from kubeconfig
	rawCfg, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return "", fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// Determine context to use
	activeContext := contextName
	if activeContext == "" {
		activeContext = rawCfg.CurrentContext
	}
	if activeContext == "" {
		// No context => fallback to default namespace
		return "default", nil
	}

	ctx, ok := rawCfg.Contexts[activeContext]
	if !ok || ctx == nil {
		return "default", nil
	}

	if ctx.Namespace != "" {
		return ctx.Namespace, nil
	}

	return "default", nil
}
