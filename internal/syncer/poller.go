package syncer

import (
	"context"
	"log"
	"time"

	"github.com/rahul/delivery-service/internal/config"
	"github.com/rahul/delivery-service/internal/models"
	"github.com/rahul/delivery-service/internal/store"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var lastSynced time.Time = time.Time{}

func SyncCampaignsPoller(ctx context.Context, client *mongo.Client) {
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

		var rawCampaigns []models.RawCampaign
		if err := cursor.All(ctx, &rawCampaigns); err != nil {
			log.Printf("SYNC failed to decode campaigns: %v", err)
			time.Sleep(10 * time.Second)
			continue
		}

		log.Printf("SYNC : %d updated campaigns fetched", len(rawCampaigns))

		// Transform raw to type campaigns
		newCampaigns := make([]models.Campaign, 0)
		for _, rc := range rawCampaigns {
			c := transformRawCampaign(rc)
			newCampaigns = append(newCampaigns, c)
		}

		// Merge with the in memory campaigns
		store.MergeUpdatedCampaigns(newCampaigns)

		//update last synced time to now
		lastSynced = time.Now()

		// sync interval until the next run
		time.Sleep(config.SyncInterval)
	}
}
