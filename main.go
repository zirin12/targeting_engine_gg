package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// campaigns global variable for now for quick & dirty solution
var campaignsInMemory []campaign
var validDimensions map[string]struct{}

type dimension struct {
	name        string
	isInclusion bool
	values      map[string]struct{}
}

type campaign struct {
	CampaignID string      `json:"cid"`
	Name       string      `json:"name"`
	Image      string      `json:"image"`
	Cta        string      `json:"cta"`
	IsActive   bool        `json:"-"`
	Dimensions []dimension `json:"-"`
}

func isMatchingCampaign(c campaign, querydimensionMap map[string]string) bool {
	match := true
	for qdim, qval := range querydimensionMap {
		found := false
		log.Printf("input query param %v and value %v\n", qdim, qval)
		for _, d := range c.Dimensions {
			if d.name == qdim {
				found = true
				if _, ok := d.values[qval]; ok {
					if !d.isInclusion {
						log.Printf("match:false")
						return false
					}
				} else {
					if d.isInclusion {
						log.Printf("match:false")
						return false
					}
				}
				match = match && true
				break
			}
		}
		if !found {
			match = match && true
		}
	}
	log.Printf("match: %v", match)
	return match
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func deliveryHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Only GET method allowed", http.StatusMethodNotAllowed)
		return
	}

	dimensionMap := make(map[string]string)
	// Extract query parameters and put them in map
	for d := range validDimensions {
		dval := r.URL.Query().Get(d)
		if dval == "" {
			http.Error(w, "missing required query param:"+d, http.StatusBadRequest)
			return
		}
		dimensionMap[d] = dval
	}

	campaignsRes := make([]campaign, 0) // make a 0 length slice not just var as the json will return nil in the latter case
	for _, c := range campaignsInMemory {
		if c.IsActive {
			isMatch := isMatchingCampaign(c, dimensionMap)
			log.Println("campaign id: ", c.CampaignID)
			if isMatch {
				campaignsRes = append(campaignsRes, c)
			}
		}
	}
	log.Println("matched campaigns: ", campaignsRes)
	if len(campaignsRes) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(campaignsRes)
}

func main() {
	campaignsInMemory = loadMockCampaigns()
	validDimensions = LoadMockValidDimensions()
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/v1/delivery", deliveryHandler)

	port := ":8080"
	log.Printf("Starting delivery service on port %s ...", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// To be loaded from DB
func loadMockCampaigns() []campaign {
	return []campaign{
		{
			CampaignID: "duolingo",
			Name:       "Duolingo XP Boost",
			Image:      "https://cdn.duo/img1.png",
			Cta:        "Install Now",
			IsActive:   true,
			Dimensions: []dimension{
				{
					name:        "country",
					isInclusion: true,
					values: map[string]struct{}{
						"IN": {},
						"US": {},
					},
				},
				{
					name:        "os",
					isInclusion: false,
					values: map[string]struct{}{
						"iOS": {},
					},
				},
			},
		},
		{
			CampaignID: "flipkart",
			Name:       "Flipkart Fest",
			Image:      "https://cdn.flipkart/img2.png",
			Cta:        "Shop Now",
			IsActive:   true,
			Dimensions: []dimension{
				{
					name:        "app",
					isInclusion: true,
					values: map[string]struct{}{
						"com.flipkart.android": {},
					},
				},
			},
		},
	}
}

// To be loaded from a config or something
func LoadMockValidDimensions() map[string]struct{} {
	return map[string]struct{}{
		"country": {},
		"app":     {},
		"os":      {},
	}
}
