package api

import (
	"encoding/json"
	"net/http"
	"strings"

	platformk8s "github.com/johnjisna/platformpilot/internal/platform"
	"k8s.io/client-go/kubernetes"
)

func DeleteResource(clientset *kubernetes.Clientset) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/api/v1/resources/")

		parts := strings.SplitN(path, "/", 2)

		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			http.Error(
				w,
				"resource type and name are required",
				http.StatusBadRequest,
			)
			return
		}

		resourceType := parts[0]
		name := parts[1]

		err := platformk8s.DeleteResource(
			clientset,
			resourceType,
			name,
		)

		if err != nil {
			http.Error(
				w,
				err.Error(),
				http.StatusInternalServerError,
			)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "resource deleted",
			"type":    resourceType,
			"name":    name,
		})
	}
}
