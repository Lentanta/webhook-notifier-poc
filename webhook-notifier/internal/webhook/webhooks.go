package webhook

import (
	"slices"
	"time"
	"webhook-notifier/internal/models"
)

// --- Mock the get webhook API
func GetWebhookList() []models.Webhook {
	return []models.Webhook{
		{ID: "WH01", Events: []string{"subscriber.added_to_segment"},
			PostURL: "http://localhost:3000/customer-webhook"},
		{ID: "WH02", Events: []string{"subscriber.created"},
			PostURL: "http://localhost:3000/customer-webhook"},
		{ID: "WH03", Events: []string{"subscriber.unsubscribed"},
			PostURL: "http://localhost:3000/customer-webhook"},
	}
}

// --- Find the web hook with event name
func GetWebhook(eventName string) []models.Webhook {
	var result []models.Webhook
	webhooks := GetWebhookList()

	// Simulate quering DB or getting from cache
	time.Sleep(1 * time.Second)

	for _, whook := range webhooks {
		if slices.Contains(whook.Events, eventName) {
			result = append(result, whook)
		}
	}

	return result
}
