package platform

import (
	"context"
	"fmt"
	"sort"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Incident struct {
	ID                string    `json:"incidentId"`
	Type              string    `json:"type"`
	Severity          string    `json:"severity"`
	Resource          string    `json:"resource"`
	Namespace         string    `json:"namespace,omitempty"`
	Status            string    `json:"status"`
	Summary           string    `json:"summary"`
	Evidence          []string  `json:"evidence"`
	RecommendedAction string    `json:"recommendedAction"`
	DetectedAt        time.Time `json:"detectedAt"`
}

func GetIncidents(clientset *kubernetes.Clientset) ([]Incident, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	now := time.Now().UTC()
	var incidents []Incident
	index := make(map[string]int)

	addIncident := func(incident Incident) {
		key := incident.Namespace + "/" + incident.Resource

		if i, exists := index[key]; exists {
			current := &incidents[i]
			current.Evidence = append(current.Evidence, incident.Evidence...)

			if incident.Status == "ACTIVE" {
				current.Status = "ACTIVE"
			}

			priority := map[string]int{
				"CRITICAL": 1,
				"HIGH":     2,
				"MEDIUM":   3,
				"LOW":      4,
			}

			if priority[incident.Severity] < priority[current.Severity] {
				current.Severity = incident.Severity
			}
			return
		}

		index[key] = len(incidents)
		incidents = append(incidents, incident)
	}

	// 1. Detect unhealthy nodes.
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("collect nodes: %w", err)
	}

	for _, node := range nodes.Items {
		for _, condition := range node.Status.Conditions {
			if condition.Type != corev1.NodeReady ||
				condition.Status == corev1.ConditionTrue {
				continue
			}

			addIncident(Incident{
				ID:       "node:" + node.Name,
				Type:     "NODE_NOT_READY",
				Severity: "CRITICAL",
				Resource: node.Name,
				Status:   "ACTIVE",
				Summary:  "Kubernetes node is not ready",
				Evidence: []string{
					"Ready condition: " + string(condition.Status),
					"Reason: " + condition.Reason,
					"Message: " + condition.Message,
				},
				RecommendedAction: "Inspect kubelet, node conditions and container runtime",
				DetectedAt:        now,
			})
			break
		}
	}

	// 2. Detect pod and container failures.
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("collect pods: %w", err)
	}

	for _, pod := range pods.Items {
		base := Incident{
			Resource:   pod.Name,
			Namespace:  pod.Namespace,
			DetectedAt: now,
		}

		if pod.Status.Phase == corev1.PodFailed {
			incident := base
			incident.ID = "pod:" + pod.Namespace + "/" + pod.Name
			incident.Type = "POD_FAILED"
			incident.Severity = "HIGH"
			incident.Status = "ACTIVE"
			incident.Summary = "Pod has entered Failed phase"
			incident.Evidence = []string{"Pod phase: Failed"}
			incident.RecommendedAction = "Inspect pod events and container logs"
			addIncident(incident)
		}

		if pod.Status.Phase == corev1.PodPending &&
			now.Sub(pod.CreationTimestamp.Time) > 5*time.Minute {
			incident := base
			incident.ID = "pending:" + pod.Namespace + "/" + pod.Name
			incident.Type = "POD_PENDING"
			incident.Severity = "MEDIUM"
			incident.Status = "ACTIVE"
			incident.Summary = "Pod has remained pending for more than five minutes"
			incident.Evidence = []string{
				"Pod phase: Pending",
				"Pod age exceeds five minutes",
			}
			incident.RecommendedAction = "Inspect scheduling events, resource availability and PVC binding"
			addIncident(incident)
		}

		for _, container := range pod.Status.ContainerStatuses {
			// CrashLoopBackOff: current waiting state.
			if container.State.Waiting != nil &&
				container.State.Waiting.Reason == "CrashLoopBackOff" {
				incident := base
				incident.ID = "crashloop:" + pod.Namespace + "/" + pod.Name
				incident.Type = "CRASH_LOOP_BACK_OFF"
				incident.Severity = "HIGH"
				incident.Status = "ACTIVE"
				incident.Summary = "Container is repeatedly crashing"
				incident.Evidence = []string{
					"Container: " + container.Name,
					"Waiting reason: CrashLoopBackOff",
					"Restart count: " + fmt.Sprint(container.RestartCount),
					container.State.Waiting.Message,
				}
				incident.RecommendedAction = "Inspect current and previous container logs"
				addIncident(incident)
			}

			// OOMKilled can be in current OR previous termination state.
			var oomReason string
			var oomMessage string
			oomActive := false

			if container.State.Terminated != nil &&
				container.State.Terminated.Reason == "OOMKilled" {
				oomReason = container.State.Terminated.Reason
				oomMessage = container.State.Terminated.Message
				oomActive = true
			} else if container.LastTerminationState.Terminated != nil &&
				container.LastTerminationState.Terminated.Reason == "OOMKilled" {
				oomReason = container.LastTerminationState.Terminated.Reason
				oomMessage = container.LastTerminationState.Terminated.Message
			}

			if oomReason == "OOMKilled" {
				status := "OBSERVED"
				if oomActive {
					status = "ACTIVE"
				}

				incident := base
				incident.ID = "oom:" + pod.Namespace + "/" + pod.Name
				incident.Type = "OOM_KILLED"
				incident.Severity = "HIGH"
				incident.Status = status
				incident.Summary = "Container experienced memory exhaustion"
				incident.Evidence = []string{
					"Container: " + container.Name,
					"Termination reason: " + oomReason,
					"Restart count: " + fmt.Sprint(container.RestartCount),
					oomMessage,
				}
				incident.RecommendedAction = "Inspect memory usage, limits and application memory consumption"
				addIncident(incident)
			}

			// Recent restart burst.
			last := container.LastTerminationState.Terminated

			if container.RestartCount >= 3 &&
				last != nil &&
				!last.FinishedAt.IsZero() &&
				now.Sub(last.FinishedAt.Time) >= 0 &&
				now.Sub(last.FinishedAt.Time) <= 15*time.Minute {
				incident := base
				incident.ID = "restarts:" + pod.Namespace + "/" + pod.Name
				incident.Type = "HIGH_CONTAINER_RESTARTS"
				incident.Severity = "MEDIUM"
				incident.Status = "OBSERVED"
				incident.Summary = "Container experienced repeated recent restarts"
				incident.Evidence = []string{
					"Container: " + container.Name,
					"Restart count: " + fmt.Sprint(container.RestartCount),
					"Last termination: " + last.Reason,
				}
				incident.RecommendedAction = "Inspect container logs and termination history"
				addIncident(incident)
			}
		}
	}

	// 3. Detect unavailable deployments.
	deployments, err := clientset.AppsV1().Deployments("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("collect deployments: %w", err)
	}

	for _, deployment := range deployments.Items {
		desired := int32(1)
		if deployment.Spec.Replicas != nil {
			desired = *deployment.Spec.Replicas
		}

		if deployment.Status.AvailableReplicas >= desired ||
			now.Sub(deployment.CreationTimestamp.Time) < 5*time.Minute {
			continue
		}

		incident := Incident{
			ID:        "deployment:" + deployment.Namespace + "/" + deployment.Name,
			Type:      "DEPLOYMENT_UNAVAILABLE",
			Severity:  "HIGH",
			Resource:  deployment.Name,
			Namespace: deployment.Namespace,
			Status:    "ACTIVE",
			Summary:   "Deployment has fewer available replicas than desired",
			Evidence: []string{
				fmt.Sprintf("Desired replicas: %d", desired),
				fmt.Sprintf("Available replicas: %d", deployment.Status.AvailableReplicas),
				fmt.Sprintf("Ready replicas: %d", deployment.Status.ReadyReplicas),
			},
			RecommendedAction: "Inspect deployment conditions, ReplicaSets and associated pods",
			DetectedAt:        now,
		}

		for _, condition := range deployment.Status.Conditions {
			if condition.Type == appsv1.DeploymentProgressing {
				incident.Evidence = append(
					incident.Evidence,
					"Progressing reason: "+condition.Reason,
				)
			}
		}

		addIncident(incident)
	}

	// 4. Collect recent warning events across namespaces.
	events, err := clientset.CoreV1().Events("").List(ctx, metav1.ListOptions{
		FieldSelector: "type=Warning",
	})
	if err != nil {
		return nil, fmt.Errorf("collect warning events: %w", err)
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

		resource := event.InvolvedObject.Name
		if resource == "" {
			continue
		}

		addIncident(Incident{
			ID:        "event:" + event.Namespace + "/" + resource + "/" + event.Reason,
			Type:      "KUBERNETES_WARNING",
			Severity:  "MEDIUM",
			Resource:  resource,
			Namespace: event.Namespace,
			Status:    "OBSERVED",
			Summary:   "Kubernetes warning event: " + event.Reason,
			Evidence: []string{
				event.Message,
				"Event count: " + fmt.Sprint(event.Count),
			},
			RecommendedAction: "Inspect the affected resource and related events",
			DetectedAt:        lastSeen,
		})
	}

	// 5. Sort incidents by severity.
	priority := map[string]int{
		"CRITICAL": 1,
		"HIGH":     2,
		"MEDIUM":   3,
		"LOW":      4,
	}

	sort.Slice(incidents, func(i, j int) bool {
		return priority[incidents[i].Severity] <
			priority[incidents[j].Severity]
	})

	return incidents, nil
}
