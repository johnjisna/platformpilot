package kubernetes

import (
	"fmt"
	"os"
	"path/filepath"

	coreclient "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

// loadConfig loads Kubernetes configuration for both clients.
func loadConfig() (*rest.Config, error) {
	// Running inside Kubernetes
	config, err := rest.InClusterConfig()
	if err == nil {
		return config, nil
	}

	// Running locally
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("get user home directory: %w", err)
	}

	kubeconfig := filepath.Join(home, ".kube", "config")

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("load kubeconfig: %w", err)
	}

	return config, nil
}

// NewClients initializes both Kubernetes and Metrics API clients.
func NewClients() (*coreclient.Clientset, *metricsclient.Clientset, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, nil, err
	}

	k8sClient, err := coreclient.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("create Kubernetes client: %w", err)
	}

	metricsClient, err := metricsclient.NewForConfig(config)
	if err != nil {
		return nil, nil, fmt.Errorf("create Metrics API client: %w", err)
	}

	return k8sClient, metricsClient, nil
}

// NewClient preserves the existing single-client initialization function.
func NewClient() (*coreclient.Clientset, error) {
	config, err := loadConfig()
	if err != nil {
		return nil, err
	}

	return coreclient.NewForConfig(config)
}

func HealthCheck(clientset *coreclient.Clientset) error {
	_, err := clientset.Discovery().ServerVersion()
	return err
}
