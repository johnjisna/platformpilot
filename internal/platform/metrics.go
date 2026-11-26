package platform

import (
	"context"
	"fmt"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

type PodMetric struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	CPU       string `json:"cpu"`
	Memory    string `json:"memory"`
}

type PodMetricsResponse struct {
	Namespace   string      `json:"namespace"`
	CollectedAt time.Time   `json:"collectedAt"`
	Pods        []PodMetric `json:"pods"`
}

func GetPodMetrics(
	ctx context.Context,
	metricsClient *metricsclient.Clientset,
	namespace string,
) (*PodMetricsResponse, error) {

	podMetrics, err := metricsClient.MetricsV1beta1().
		PodMetricses(namespace).
		List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, fmt.Errorf("fetch pod metrics: %w", err)
	}

	response := &PodMetricsResponse{
		Namespace:   namespace,
		CollectedAt: time.Now().UTC(),
		Pods:        make([]PodMetric, 0, len(podMetrics.Items)),
	}

	for _, pod := range podMetrics.Items {
		var totalCPU int64
		var totalMemory int64

		for _, container := range pod.Containers {
			totalCPU += container.Usage.Cpu().MilliValue()
			totalMemory += container.Usage.Memory().Value()
		}

		memoryQuantity := resource.NewQuantity(
			totalMemory,
			resource.BinarySI,
		)

		response.Pods = append(response.Pods, PodMetric{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			CPU:       fmt.Sprintf("%dm", totalCPU),
			Memory:    memoryQuantity.String(),
		})
	}

	return response, nil
}
