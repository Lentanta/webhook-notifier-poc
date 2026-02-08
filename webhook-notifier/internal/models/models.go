package models

// For now, every models I will just put in this file

type QMessage struct {
	Content   WebhookEvent `json:"content"`
	Timestamp string       `json:"timestamp"`
}

type WebhookEvent struct {
	EventName  string     `json:"event_name"`
	EventTime  string     `json:"event_time"`
	Subscriber Subscriber `json:"subscriber"`
	Segment    *Segment   `json:"segment,omitempty"`
	WebhookID  string     `json:"webhook_id"`
}

type Subscriber struct {
	ID             string            `json:"id"`
	Status         string            `json:"status"`
	Email          string            `json:"email"`
	Source         string            `json:"source"`
	FirstName      string            `json:"first_name"`
	LastName       string            `json:"last_name"`
	Segments       []Segment         `json:"segments"`
	CustomFields   map[string]string `json:"custom_fields"`
	OptinIP        string            `json:"optin_ip"`
	OptinTimestamp string            `json:"optin_timestamp"`
	CreatedAt      string            `json:"created_at"`
}

type Segment struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// --- Mock schema for the webhook
type Webhook struct {
	ID      string   `json:"id"`
	PostURL string   `json:"post_url"`
	Events  []string `json:"events"`
}
