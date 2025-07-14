package main

import (
	"context"
	"log"
	"net/http"

	"github.com/rahul/delivery-service/internal/config"
	"github.com/rahul/delivery-service/internal/httpapi"
	"github.com/rahul/delivery-service/internal/syncer"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(config.MongoURI()))
	if err != nil {
		log.Fatalf("Mongo failed to connect: %v", err)
	}
	// Sync job that runs every N seconds polling to update in memory campaigns from db
	go syncer.SyncCampaignsPoller(context.Background(), client)
	//campaignsInMemory = loadMockCampaigns()
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("OK")) })
	http.HandleFunc("/v1/delivery", httpapi.DeliveryHandler)

	log.Printf("Starting delivery service on port %s ...", config.Port())
	if err := http.ListenAndServe(config.Port(), nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
