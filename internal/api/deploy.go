package api

import (
	"encoding/json"
	"net/http"

	platformk8s "github.com/johnjisna/platformpilot/internal/platform"
	"k8s.io/client-go/kubernetes"
)

func DeployApplication(clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var app platformk8s.ApplicationRequest

		if err := json.NewDecoder(r.Body).Decode(&app); err != nil {
			http.Error(w, "invalid request body", http.StatusBadRequest)
			return
		}

		if app.Name == "" || app.Image == "" {
			http.Error(w, "name and image are required", http.StatusBadRequest)
			return
		}

		if app.Replicas <= 0 {
			app.Replicas = 1
		}

		if err := platformk8s.CreateDeployment(clientset, app); err != nil {
			http.Error(w, "failed to create deployment: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := platformk8s.CreateService(clientset, app); err != nil {
			http.Error(w, "failed to create service: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)

		json.NewEncoder(w).Encode(map[string]interface{}{
			"message":  "application deployed",
			"name":     app.Name,
			"image":    app.Image,
			"replicas": app.Replicas,
		})
	}
}
