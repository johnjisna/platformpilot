package api

import (
	"encoding/json"
	"net/http"
	"strings"

	platformk8s "github.com/johnjisna/platformpilot/internal/platform"
	"k8s.io/client-go/kubernetes"
)

type ScaleRequest struct {
	Replicas int32 `json:"replicas"`
}

func ScaleApplication(clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		name := strings.TrimSuffix(
			strings.TrimPrefix(r.URL.Path, "/api/v1/applications/"),
			"/scale",
		)

		if name == "" {
			http.Error(w, "application name is required", http.StatusBadRequest)
			return
		}

		var request ScaleRequest

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if request.Replicas < 1 {
			http.Error(w, "replicas must be at least 1", http.StatusBadRequest)
			return
		}

		if err := platformk8s.ScaleDeployment(
			clientset,
			name,
			request.Replicas,
		); err != nil {
			http.Error(w, "failed to scale application: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "application scaled",
			"name":     name,
			"replicas": request.Replicas,
		})
	}
}
