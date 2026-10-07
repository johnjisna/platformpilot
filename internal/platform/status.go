package platform

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ApplicationStatus struct {
	Name              string `json:"name"`
	DesiredReplicas   int32  `json:"desiredReplicas"`
	ReadyReplicas     int32  `json:"readyReplicas"`
	AvailableReplicas int32  `json:"availableReplicas"`
	Status            string `json:"status"`
}

func GetDeploymentStatus(
	clientset *kubernetes.Clientset,
	name string,
) (*ApplicationStatus, error) {

	deployment, err := clientset.AppsV1().
		Deployments("platformpilot").
		Get(context.Background(), name, metav1.GetOptions{})

	if err != nil {
		return nil, err
	}

	status := "Progressing"

	if deployment.Status.ReadyReplicas == deployment.Status.Replicas &&
		deployment.Status.AvailableReplicas == deployment.Status.Replicas {
		status = "Running"
	}

	return &ApplicationStatus{
		Name:              deployment.Name,
		DesiredReplicas:   replicas(deployment),
		ReadyReplicas:     deployment.Status.ReadyReplicas,
		AvailableReplicas: deployment.Status.AvailableReplicas,
		Status:            status,
	}, nil
}

func replicas(deployment *appsv1.Deployment) int32 {
	if deployment.Spec.Replicas == nil {
		return 0
	}

	return *deployment.Spec.Replicas
}
