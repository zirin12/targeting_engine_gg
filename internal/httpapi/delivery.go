package httpapi

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/rahul/delivery-service/internal/config"
	"github.com/rahul/delivery-service/internal/matcher"
	"github.com/rahul/delivery-service/internal/store"
)

var ValidDimensions map[string]struct{}

func DeliveryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	dimensionMap := make(map[string]string)
	// Extract query parameters and put them in map
	for d := range config.ValidDimensions() {
		dval := r.URL.Query().Get(d)
		if dval == "" {
			http.Error(w, "missing required query param:"+d, http.StatusBadRequest)
			return
		}
		dimensionMap[d] = dval
	}

	campaignsRes := matcher.FilterCampaigns(store.GetSnapShot(), dimensionMap)
	log.Println("matched campaigns: ", campaignsRes)
	if len(campaignsRes) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(campaignsRes)
}
