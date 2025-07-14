package matcher

import (
	"log"

	"github.com/rahul/delivery-service/internal/models"
)

func IsMatchingCampaign(c models.Campaign, querydimensionMap map[string]string) bool {
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

func FilterCampaigns(campaigns []models.Campaign, querydimensionMap map[string]string) []models.Campaign {
	campaignsRes := make([]models.Campaign, 0) // make a 0 length slice not just var as the json will return nil in the latter case
	for _, c := range campaigns {
		if c.IsActive {
			isMatch := IsMatchingCampaign(c, querydimensionMap)
			log.Println("campaign id: ", c.CampaignID)
			if isMatch {
				campaignsRes = append(campaignsRes, c)
			}
		}
	}
	return campaignsRes
}
