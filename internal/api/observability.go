package api

import (
	"encoding/json"
	"net/http"

	platformk8s "github.com/johnjisna/platformpilot/internal/platform"
	"k8s.io/client-go/kubernetes"
)

func ClusterOverview(clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodGet {
			http.Error(
				w,
				"method not allowed",
				http.StatusMethodNotAllowed,
			)
			return
		}

		overview, err := platformk8s.GetClusterOverview(clientset)
		if err != nil {
			http.Error(
				w,
				"failed to collect cluster overview: "+err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(overview)
	}
}
