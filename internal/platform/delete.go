package platform

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func DeleteDeployment(
	clientset *kubernetes.Clientset,
	name string,
) error {
	return clientset.AppsV1().
		Deployments("platformpilot").
		Delete(context.Background(), name, metav1.DeleteOptions{})
}
