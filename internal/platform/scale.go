package platform

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ScaleDeployment(
	clientset *kubernetes.Clientset,
	name string,
	replicas int32,
) error {
	deployment, err := clientset.AppsV1().
		Deployments("platformpilot").
		Get(context.Background(), name, metav1.GetOptions{})

	if err != nil {
		return err
	}

	deployment.Spec.Replicas = &replicas

	_, err = clientset.AppsV1().
		Deployments("platformpilot").
		Update(context.Background(), deployment, metav1.UpdateOptions{})

	return err
}
