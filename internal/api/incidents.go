package api

import (
	"encoding/json"
	"net/http"

	platformk8s "github.com/johnjisna/platformpilot/internal/platform"
	"k8s.io/client-go/kubernetes"
)

func Incidents(clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
			return
		}

		incidents, err := platformk8s.GetIncidents(clientset)
		if err != nil {
			http.Error(
				w,
				"failed to collect incidents: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if incidents == nil {
			incidents = []platformk8s.Incident{}
		}

		if err := json.NewEncoder(w).Encode(incidents); err != nil {
			http.Error(
				w,
				"failed to encode incidents",
				http.StatusInternalServerError,
			)
		}
	}
}
