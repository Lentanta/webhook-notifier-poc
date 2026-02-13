package models

// QMessage is the message format for RabbitMQ queue
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

// DBEvent represents an event row from the database
type DBEvent struct {
	ID        int64  `db:"id"`
	EventName string `db:"event_name"`
	EventTime string `db:"event_time"`
	Payload   string `db:"payload"`
	WebhookID string `db:"webhook_id"`
	CreatedAt string `db:"created_at"`
}
