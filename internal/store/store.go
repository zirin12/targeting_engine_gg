package store

import (
	"sync"

	"github.com/rahul/delivery-service/internal/models"
)

var campaignsInMemory []*SafeCampaign

type SafeCampaign struct {
	sync.RWMutex
	Data models.Campaign
}

func GetSnapShot() []models.Campaign {
	result := make([]models.Campaign, 0, len(campaignsInMemory))
	for _, sc := range campaignsInMemory {
		sc.RLock()
		result = append(result, sc.Data)
		sc.RUnlock()
	}
	return result
}

/*
func (sc *SafeCampaign) GetData() models.Campaign {
	sc.RLock()
	defer sc.RUnlock()
	return sc.Data
}
*/

func MergeUpdatedCampaigns(updated []models.Campaign) {
	for _, newC := range updated {
		updatedId := newC.CampaignID
		found := false

		for _, sc := range campaignsInMemory {
			if sc.Data.CampaignID == updatedId {
				sc.Lock()
				sc.Data = newC
				sc.Unlock()
				found = true
				break
			}
		}
		if !found {
			campaignsInMemory = append(campaignsInMemory, &SafeCampaign{Data: newC})
		}
	}
}
