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

// Client wraps Kubernetes client với các helper methods
type Client struct {
	Clientset *kubernetes.Clientset
	Config    *rest.Config
	Context   context.Context
}

// NewClient tạo một Kubernetes client mới
// Tự động phát hiện cấu hình từ kubeconfig hoặc in-cluster config
func NewClient(kubeconfig string, contextName string) (*Client, error) {
	var config *rest.Config
	var err error

	if kubeconfig == "" {
		// Thử tìm kubeconfig file ở vị trí mặc định
		if home := homedir.HomeDir(); home != "" {
			kubeconfig = filepath.Join(home, ".kube", "config")
		}
	}

	// Kiểm tra xem có đang chạy trong cluster không
	if _, err := os.Stat("/var/run/secrets/kubernetes.io/serviceaccount/token"); err == nil {
		// Sử dụng in-cluster config
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
		}
	} else {
		// Sử dụng kubeconfig file
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig: %w", err)
		}

		// Nếu có context name được chỉ định, load config với context đó
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

	// Tạo clientset
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

// SetNamespace thiết lập default namespace cho client
func (c *Client) SetNamespace(namespace string) {
	// Cập nhật context với namespace
	c.Context = context.WithValue(c.Context, "namespace", namespace)
}
