package models

import "time"

type Dimension struct {
	Name        string
	IsInclusion bool
	Values      map[string]struct{}
}

type Campaign struct {
	CampaignID string      `json:"cid"`
	Name       string      `json:"name"`
	Image      string      `json:"image"`
	Cta        string      `json:"cta"`
	IsActive   bool        `json:"-"`
	Dimensions []Dimension `json:"-"`
}

type RawDimension struct {
	Name        string   `bson:"name"`
	IsInclusion bool     `bson:"is_inclusion"`
	Values      []string `bson:"values"`
}

type RawCampaign struct {
	CampaignID string         `bson:"campaign_id"`
	Name       string         `bson:"name"`
	Image      string         `bson:"image"`
	CTA        string         `bson:"cta"`
	IsActive   bool           `bson:"is_active"`
	CreatedAt  time.Time      `bson:"created_at"`
	UpdatedAt  time.Time      `bson:"updated_at"`
	Dimensions []RawDimension `bson:"dimensions"`
}
