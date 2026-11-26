package api

import (
	"encoding/json"
	"net/http"

	"github.com/johnjisna/platformpilot/internal/platform"
	metricsclient "k8s.io/metrics/pkg/client/clientset/versioned"
)

func PodMetrics(metricsClient *metricsclient.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		namespace := r.URL.Query().Get("namespace")
		if namespace == "" {
			namespace = "platformpilot"
		}

		result, err := platform.GetPodMetrics(
			r.Context(),
			metricsClient,
			namespace,
		)
		if err != nil {
			http.Error(
				w,
				"Failed to collect pod metrics: "+err.Error(),
				http.StatusBadGateway,
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if err := json.NewEncoder(w).Encode(result); err != nil {
			return
		}
	}
}
