package config

import (
	"os"
	"time"
)

func ValidDimensions() map[string]struct{} {
	return map[string]struct{}{
		"country": {},
		"app":     {},
		"os":      {},
	}
}

func MongoURI() string {
	if uri := os.Getenv("MONGO_URI"); uri != "" {
		return uri
	}
	return "mongodb://localhost:27017"
}

func Port() string {
	if p := os.Getenv("PORT"); p != "" {
		return ":" + p
	}
	return ":8080"
}

// sync/poller interval
const SyncInterval = 30 * time.Second
