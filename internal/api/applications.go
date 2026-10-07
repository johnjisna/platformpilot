package api

import (
	"encoding/json"
	"net/http"
)

type Application struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Status      string `json:"status"`
	Replicas    int    `json:"replicas"`
}

func Applications(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	applications := []Application{
		{
			Name:        "payment-api",
			Environment: "dev",
			Status:      "healthy",
			Replicas:    2,
		},
		{
			Name:        "frontend",
			Environment: "dev",
			Status:      "healthy",
			Replicas:    3,
		},
	}

	json.NewEncoder(w).Encode(applications)
}
