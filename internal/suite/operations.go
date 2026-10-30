package suite

import "github.com/ghassan-ai-projects/streams-simulator/internal/suite/internal/app"

func Publish(options PublishOptions) (Published, error) {
	return app.Publish(options)
}

func PublishCampaign(options CampaignOptions) (PublishedCampaign, error) {
	return app.PublishCampaign(options)
}
