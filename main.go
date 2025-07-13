package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func deliveryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract query parameters
	app := r.URL.Query().Get("app")
	country := r.URL.Query().Get("country")
	os := r.URL.Query().Get("os")

	// Check if any of the params are empty
	if app == "" || country == "" || os == "" {
		http.Error(w, "missing required query params", http.StatusBadRequest)
		return
	}

	// TODO: All campaigns need to be in a slice of structs in memory fetched from the DB
	// Mocked response - just to check the response
	campaigns := []map[string]string{
		{
			"cid": "duolingo",
			"img": "https://somelink2",
			"cta": "Install",
		},
	}

	// TODO: need to match the campaign rules against the input params to know which should be considered

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(campaigns)
}

func main() {
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/v1/delivery", deliveryHandler)

	port := ":8080"
	log.Printf("Starting delivery service on port %s ...", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
