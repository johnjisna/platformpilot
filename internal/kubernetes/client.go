package kubernetes

import (
	"os"
	"path/filepath"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func NewClient() (*kubernetes.Clientset, error) {

	// Running inside Kubernetes
	config, err := rest.InClusterConfig()

	if err == nil {
		return kubernetes.NewForConfig(config)
	}

	// Running locally
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	kubeconfig := filepath.Join(home, ".kube", "config")

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}

func HealthCheck(clientset *kubernetes.Clientset) error {
	_, err := clientset.Discovery().ServerVersion()
	return err
}
