package api

import (
	"net/http"
	"strings"

	platformk8s "github.com/johnjisna/platformpilot/internal/platform"
	"k8s.io/client-go/kubernetes"
)

func DeleteApplication(clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		name := strings.TrimPrefix(
			r.URL.Path,
			"/api/v1/applications/",
		)

		if name == "" {
			http.Error(w, "application name is required", http.StatusBadRequest)
			return
		}

		if err := platformk8s.DeleteDeployment(clientset, name); err != nil {
			http.Error(
				w,
				"failed to delete application: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)

		w.Write([]byte(`{"message":"application deleted"}`))
	}
}
