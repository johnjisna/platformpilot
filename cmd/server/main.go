package main

import (
	"log"
	"net/http"

	"github.com/johnjisna/platformpilot/internal/api"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", api.Health)

	log.Println("PlatformPilot running on :8080")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
