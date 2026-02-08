package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"
	"webhook-notifier/internal/models"
	"webhook-notifier/internal/webhook"
)

// Config for retry sending webhook
const (
	maxRetries = 5
	baseDelay  = 1 * time.Second
	maxDelay   = 30 * time.Second
)

func ProcessSendWebhook(we models.WebhookEvent) error {

	jsonData, err := json.Marshal(we)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	webhooks := webhook.GetWebhook(we.EventName)
	for _, wh := range webhooks {
		err := handleRetrySendWebhook(wh.PostURL, jsonData)
		if err != nil {
			return fmt.Errorf("Failed to send and retry")
		}
	}

	return nil
}

func handleRetrySendWebhook(postUrl string, jsonData []byte) error {
	for attempt := 0; attempt <= maxRetries; attempt++ {
		err, isRetry := sendWebhook(postUrl, jsonData)

		if err != nil && !isRetry {
			return fmt.Errorf("Failed to send webhook")
		}

		if attempt == maxRetries {
			return fmt.Errorf("Failed to send webhook, max retry")
		}

		if err == nil {
			break
		}

		delay := calculateDelay(attempt)
		fmt.Println("Retry send webhook")
		time.Sleep(delay)
	}

	return nil
}

func sendWebhook(postUrl string, jsonData []byte) (error, bool) {
	contentType := "application/json"
	resp, err := http.Post(
		postUrl,
		contentType,
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		return fmt.Errorf("Send http error"), false
	}
	defer resp.Body.Close()

	fmt.Println("Status Code: ", resp.StatusCode)
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil, false
	}

	// Need to retry send again
	return fmt.Errorf("Status Code: %v", resp.StatusCode), true
}

func calculateDelay(attempt int) time.Duration {
	// Exponential backoff: 1s, 2s, 4s, 8s... and max is 30
	delayTimeInSec := time.Duration(math.Pow(2, float64(attempt))) * baseDelay

	// Cap at max delay
	return min(delayTimeInSec, maxDelay)
}
