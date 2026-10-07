package platform

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Resource struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Replicas  int32  `json:"replicas,omitempty"`
	Available int32  `json:"available,omitempty"`
	Image     string `json:"image,omitempty"`
}

type Resources struct {
	Deployments []Resource `json:"deployments"`
	Pods        []Resource `json:"pods"`
	Services    []Resource `json:"services"`
}

func GetResources(clientset *kubernetes.Clientset) (*Resources, error) {
	ctx := context.Background()
	namespace := "platformpilot"

	result := &Resources{}

	deployments, err := clientset.AppsV1().
		Deployments(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, deployment := range deployments.Items {
		status := "Progressing"

		if deployment.Status.ReadyReplicas == deployment.Status.Replicas &&
			deployment.Status.AvailableReplicas == deployment.Status.Replicas {
			status = "Running"
		}

		var image string

		if len(deployment.Spec.Template.Spec.Containers) > 0 {
			image = deployment.Spec.Template.Spec.Containers[0].Image
		}

		result.Deployments = append(result.Deployments, Resource{
			Name:      deployment.Name,
			Type:      "Deployment",
			Namespace: namespace,
			Status:    status,
			Replicas:  deployment.Status.Replicas,
			Available: deployment.Status.AvailableReplicas,
			Image:     image,
		})
	}

	pods, err := clientset.CoreV1().
		Pods(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		result.Pods = append(result.Pods, Resource{
			Name:      pod.Name,
			Type:      "Pod",
			Namespace: namespace,
			Status:    string(pod.Status.Phase),
		})
	}

	services, err := clientset.CoreV1().
		Services(namespace).
		List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, service := range services.Items {
		result.Services = append(result.Services, Resource{
			Name:      service.Name,
			Type:      "Service",
			Namespace: namespace,
			Status:    "Active",
		})
	}

	return result, nil
}
