package main

import (
	"log"
	"net/http"
	"strings"

	"github.com/johnjisna/platformpilot/internal/api"
	platformk8s "github.com/johnjisna/platformpilot/internal/kubernetes"
)

func main() {
	k8sClient, err := platformk8s.NewClient()
	if err != nil {
		log.Fatal("failed to create Kubernetes client:", err)
	}

	if err := platformk8s.HealthCheck(k8sClient); err != nil {
		log.Fatal("failed to connect to Kubernetes:", err)
	}

	log.Println("Connected to Kubernetes")

	mux := http.NewServeMux()

	// Frontend
	mux.HandleFunc("/", api.UI)

	mux.Handle(
		"/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.Dir("web/static")),
		),
	)

	// Existing API
	mux.HandleFunc("/health", api.Health)
	mux.HandleFunc("/applications", api.Applications)

	// Deploy
	mux.Handle(
		"/api/v1/applications",
		api.DeployApplication(k8sClient),
	)

	// Status, Scale and Delete
	mux.HandleFunc("/api/v1/applications/", func(w http.ResponseWriter, r *http.Request) {

		if strings.HasSuffix(r.URL.Path, "/scale") {
			api.ScaleApplication(k8sClient)(w, r)
			return
		}

		if r.Method == http.MethodDelete {
			api.DeleteApplication(k8sClient)(w, r)
			return
		}

		api.ApplicationStatus(k8sClient)(w, r)
	})

	mux.Handle(
		"/api/v1/resources",
		api.Resources(k8sClient),
	)

	mux.HandleFunc(
		"/api/v1/resources/",
		api.DeleteResource(k8sClient),
	)

	log.Println("PlatformPilot running on :8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
