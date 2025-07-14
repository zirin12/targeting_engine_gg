package syncer

import "github.com/rahul/delivery-service/internal/models"

// Transform mongo retreived campaigns to what we want
func transformRawCampaign(rc models.RawCampaign) models.Campaign {
	campaign := models.Campaign{}
	campaign.CampaignID = rc.CampaignID
	campaign.Cta = rc.CTA
	campaign.Image = rc.Image
	campaign.IsActive = rc.IsActive
	campaign.Name = rc.Name
	campaign.Dimensions = []models.Dimension{}
	for _, rcDimension := range rc.Dimensions {
		values := map[string]struct{}{}
		for _, rd := range rcDimension.Values {
			values[rd] = struct{}{}
		}
		campaign.Dimensions = append(campaign.Dimensions, models.Dimension{
			Name:        rcDimension.Name,
			IsInclusion: rcDimension.IsInclusion,
			Values:      values,
		})
	}
	return campaign
}
