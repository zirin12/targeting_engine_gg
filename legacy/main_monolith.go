// Legacy prototype — not wired into the modular system. Preserved for reference.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// campaigns global variable for now for quick & dirty solution
var campaignsInMemory []*safeCampaign
var validDimensions map[string]struct{}
var lastSynced time.Time = time.Time{}

type safeCampaign struct {
	sync.RWMutex
	data campaign
}

func (sc *safeCampaign) GetData() campaign {
	sc.RLock()
	defer sc.RUnlock()
	return sc.data
}

type dimension struct {
	Name        string
	IsInclusion bool
	Values      map[string]struct{}
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
	log.Println("campaign dimensions: ", c.Dimensions)
	for qdim, qval := range querydimensionMap {
		found := false
		log.Printf("input query param %v and value %v\n", qdim, qval)
		for _, d := range c.Dimensions {
			if d.Name == qdim {
				found = true
				if _, ok := d.Values[qval]; ok {
					if !d.IsInclusion {
						log.Printf("match:false")
						return false
					}
				} else {
					if d.IsInclusion {
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

func filterCampaigns(querydimensionMap map[string]string) []campaign {
	campaignsRes := make([]campaign, 0) // make a 0 length slice not just var as the json will return nil in the latter case
	for _, sc := range campaignsInMemory {
		c := sc.GetData()
		if c.IsActive {
			isMatch := isMatchingCampaign(c, querydimensionMap)
			log.Println("campaign id: ", c.CampaignID)
			if isMatch {
				campaignsRes = append(campaignsRes, c)
			}
		}
	}
	return campaignsRes
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

	campaignsRes := filterCampaigns(dimensionMap)
	log.Println("matched campaigns: ", campaignsRes)
	if len(campaignsRes) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(campaignsRes)
}

func main() {

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatalf("Mongo failed to connect: %v", err)
	}
	// Sync job that runs every 30s or poller to update in memory campaigns from db
	go syncCampaignsPoller(context.Background(), client)

	//campaignsInMemory = loadMockCampaigns()
	validDimensions = LoadMockValidDimensions()
	http.HandleFunc("/health", healthHandler)
	http.HandleFunc("/v1/delivery", deliveryHandler)

	port := ":8080"
	log.Printf("Starting delivery service on port %s ...", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

type rawDimension struct {
	Name        string   `bson:"name"`
	IsInclusion bool     `bson:"is_inclusion"`
	Values      []string `bson:"values"`
}

type rawCampaign struct {
	CampaignID string         `bson:"campaign_id"`
	Name       string         `bson:"name"`
	Image      string         `bson:"image"`
	CTA        string         `bson:"cta"`
	IsActive   bool           `bson:"is_active"`
	CreatedAt  time.Time      `bson:"created_at"`
	UpdatedAt  time.Time      `bson:"updated_at"`
	Dimensions []rawDimension `bson:"dimensions"`
}

// Transform mongo retreived campaigns to what we want
func transformRawCampaign(rc rawCampaign) campaign {
	campaign := campaign{}
	campaign.CampaignID = rc.CampaignID
	campaign.Cta = rc.CTA
	campaign.Image = rc.Image
	campaign.IsActive = rc.IsActive
	campaign.Name = rc.Name
	campaign.Dimensions = []dimension{}
	for _, rcDimension := range rc.Dimensions {
		values := map[string]struct{}{}
		for _, rd := range rcDimension.Values {
			values[rd] = struct{}{}
		}
		campaign.Dimensions = append(campaign.Dimensions, dimension{
			Name:        rcDimension.Name,
			IsInclusion: rcDimension.IsInclusion,
			Values:      values,
		})
	}
	return campaign
}

func mergeUpdatedCampaigns(updated []campaign) {
	for _, newC := range updated {
		updatedId := newC.CampaignID
		found := false

		for _, sc := range campaignsInMemory {
			if sc.GetData().CampaignID == updatedId {
				sc.Lock()
				sc.data = newC
				sc.Unlock()
				found = true
				break
			}
		}
		if !found {
			campaignsInMemory = append(campaignsInMemory, &safeCampaign{data: newC})
		}
	}
}

func syncCampaignsPoller(ctx context.Context, client *mongo.Client) {
	collection := client.Database("campaignDB").Collection("campaigns")
	for {
		// Fetch campaigns updated after last sync
		filter := bson.M{
			"updated_at": bson.M{
				"$gt": lastSynced,
			},
		}
		cursor, err := collection.Find(ctx, filter)
		if err != nil {
			log.Printf("SYNC failed to query campaigns: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		var rawCampaigns []rawCampaign
		if err := cursor.All(ctx, &rawCampaigns); err != nil {
			log.Printf("SYNC failed to decode campaigns: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		log.Printf("SYNC : %d updated campaigns fetched", len(rawCampaigns))

		// Transform raw to type campaigns
		newCampaigns := make([]campaign, 0)
		for _, rc := range rawCampaigns {
			c := transformRawCampaign(rc)
			newCampaigns = append(newCampaigns, c)
		}

		// Merge with the in memory campaigns
		mergeUpdatedCampaigns(newCampaigns)

		//update last synced time to now
		lastSynced = time.Now()

		// sync interval until the next run
		time.Sleep(60 * time.Second)
	}
}

// To be loaded from DB
func loadMockCampaigns() []*safeCampaign {
	return []*safeCampaign{
		{
			data: campaign{
				CampaignID: "duolingo",
				Name:       "Duolingo Boost",
				Image:      "https://cdn.duo/img1.png",
				Cta:        "Install",
				IsActive:   true,
				Dimensions: []dimension{
					{
						Name:        "country",
						IsInclusion: true,
						Values:      map[string]struct{}{"IN": {}, "US": {}},
					},
					{
						Name:        "os",
						IsInclusion: false,
						Values:      map[string]struct{}{"iOS": {}},
					},
				},
			},
		},
		{
			data: campaign{
				CampaignID: "amazon",
				Name:       "Amazon Prime Bonus",
				Image:      "https://cdn.amazon/img3.png",
				Cta:        "Claim",
				IsActive:   false,
				Dimensions: []dimension{
					{
						Name:        "country",
						IsInclusion: false,
						Values:      map[string]struct{}{"CN": {}, "RU": {}},
					},
				},
			},
		},
		{
			data: campaign{
				CampaignID: "nykaa",
				Name:       "Nykaa Beauty Drop",
				Image:      "https://cdn.nykaa/img4.png",
				Cta:        "Explore",
				IsActive:   true,
				Dimensions: []dimension{
					{
						Name:        "os",
						IsInclusion: true,
						Values:      map[string]struct{}{"Android": {}, "iOS": {}},
					},
					{
						Name:        "country",
						IsInclusion: true,
						Values:      map[string]struct{}{"IN": {}},
					},
					{
						Name:        "app",
						IsInclusion: false,
						Values:      map[string]struct{}{"duolingo": {}}, // exclude duolingo app
					},
				},
			},
		},
		{
			data: campaign{
				CampaignID: "cred",
				Name:       "CRED Cashback",
				Image:      "https://cdn.cred/img5.png",
				Cta:        "Pay Now",
				IsActive:   true,
				Dimensions: []dimension{
					{
						Name:        "app",
						IsInclusion: true,
						Values:      map[string]struct{}{"cred": {}, "amazon": {}},
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
