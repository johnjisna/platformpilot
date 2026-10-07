package platform

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DeleteResource(clientset *kubernetes.Clientset, resourceType string, name string) error {
	ctx := context.Background()
	namespace := "platformpilot"

	switch resourceType {

	case "Deployment":
		return clientset.AppsV1().
			Deployments(namespace).
			Delete(ctx, name, metav1.DeleteOptions{})

	case "Pod":
		if strings.HasPrefix(name, "platformpilot-") {
			return fmt.Errorf("platformpilot pod cannot be deleted")
		}

		return clientset.CoreV1().
			Pods(namespace).
			Delete(ctx, name, metav1.DeleteOptions{})

	case "Service":
		if name == "platformpilot" {
			return fmt.Errorf("platformpilot service cannot be deleted")
		}

		return clientset.CoreV1().
			Services(namespace).
			Delete(ctx, name, metav1.DeleteOptions{})

	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}
