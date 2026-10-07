package api

import (
	"encoding/json"
	"net/http"
	"strings"

	platformk8s "github.com/johnjisna/platformpilot/internal/platform"
	"k8s.io/client-go/kubernetes"
)

func ApplicationStatus(clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		name := strings.TrimPrefix(r.URL.Path, "/api/v1/applications/")

		if name == "" {
			http.Error(w, "application name is required", http.StatusBadRequest)
			return
		}

		status, err := platformk8s.GetDeploymentStatus(clientset, name)

		if err != nil {
			http.Error(w, "application not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(status)
	}
}
