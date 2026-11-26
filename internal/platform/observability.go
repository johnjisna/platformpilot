package platform

import (
	"context"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const ObservabilityNamespace = "platformpilot"

type ClusterOverview struct {
	CollectedAt   time.Time      `json:"collectedAt"`
	Nodes         []NodeHealth   `json:"nodes"`
	Pods          []PodHealth    `json:"pods"`
	WarningEvents []WarningEvent `json:"warningEvents"`
}

type NodeHealth struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}

type PodHealth struct {
	Name                  string `json:"name"`
	Namespace             string `json:"namespace"`
	Phase                 string `json:"phase"`
	Ready                 bool   `json:"ready"`
	RestartCount          int32  `json:"restartCount"`
	FailureReason         string `json:"failureReason,omitempty"`
	LastTerminationReason string `json:"lastTerminationReason,omitempty"`
}

type WarningEvent struct {
	Namespace string    `json:"namespace"`
	Object    string    `json:"object"`
	Reason    string    `json:"reason"`
	Message   string    `json:"message"`
	Count     int32     `json:"count"`
	LastSeen  time.Time `json:"lastSeen"`
}

func GetClusterOverview(
	clientset *kubernetes.Clientset,
) (*ClusterOverview, error) {

	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)
	defer cancel()

	now := time.Now().UTC()

	result := &ClusterOverview{
		CollectedAt:   now,
		Nodes:         []NodeHealth{},
		Pods:          []PodHealth{},
		WarningEvents: []WarningEvent{},
	}

	// 1. Collect node health.
	nodes, err := clientset.CoreV1().Nodes().List(
		ctx,
		metav1.ListOptions{},
	)
	if err != nil {
		return nil, err
	}

	for _, node := range nodes.Items {
		status := "NotReady"

		for _, condition := range node.Status.Conditions {
			if condition.Type == corev1.NodeReady &&
				condition.Status == corev1.ConditionTrue {
				status = "Ready"
				break
			}
		}

		result.Nodes = append(result.Nodes, NodeHealth{
			Name:   node.Name,
			Status: status,
		})
	}

	// 2. Collect PlatformPilot application pod health.
	pods, err := clientset.CoreV1().
		Pods(ObservabilityNamespace).
		List(ctx, metav1.ListOptions{})

	if err != nil {
		return nil, err
	}

	for _, pod := range pods.Items {
		podHealth := PodHealth{
			Name:      pod.Name,
			Namespace: pod.Namespace,
			Phase:     string(pod.Status.Phase),
		}

		// Check pod readiness.
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady {
				podHealth.Ready =
					condition.Status == corev1.ConditionTrue
				break
			}
		}

		// A completed or pending pod is not currently ready.
		if pod.Status.Phase != corev1.PodRunning {
			podHealth.Ready = false
		}

		// Inspect container statuses.
		for _, container := range pod.Status.ContainerStatuses {
			podHealth.RestartCount += container.RestartCount

			// Current waiting state.
			if container.State.Waiting != nil &&
				container.State.Waiting.Reason != "" {
				podHealth.FailureReason =
					container.State.Waiting.Reason
			}

			// Current terminated state.
			if container.State.Terminated != nil &&
				container.State.Terminated.Reason != "" {
				podHealth.FailureReason =
					container.State.Terminated.Reason
			}

			// Previous termination is historical evidence,
			// not necessarily a current container failure.
			if container.LastTerminationState.Terminated != nil {
				podHealth.LastTerminationReason =
					container.LastTerminationState.Terminated.Reason
			}
		}

		result.Pods = append(result.Pods, podHealth)
	}

	// 3. Collect warning events across ALL namespaces.
	events, err := clientset.CoreV1().
		Events("").
		List(ctx, metav1.ListOptions{
			FieldSelector: "type=Warning",
		})

	if err != nil {
		return nil, err
	}

	cutoff := now.Add(-24 * time.Hour)

	for _, event := range events.Items {
		lastSeen := event.LastTimestamp.Time

		if lastSeen.IsZero() {
			lastSeen = event.EventTime.Time
		}

		if lastSeen.IsZero() {
			lastSeen = event.CreationTimestamp.Time
		}

		if lastSeen.Before(cutoff) {
			continue
		}

		result.WarningEvents = append(
			result.WarningEvents,
			WarningEvent{
				Namespace: event.Namespace,
				Object:    event.InvolvedObject.Name,
				Reason:    event.Reason,
				Message:   event.Message,
				Count:     event.Count,
				LastSeen:  lastSeen,
			},
		)
	}

	// Sort warning events: newest first.
	sort.Slice(result.WarningEvents, func(i, j int) bool {
		return result.WarningEvents[i].LastSeen.After(
			result.WarningEvents[j].LastSeen,
		)
	})

	return result, nil
}
