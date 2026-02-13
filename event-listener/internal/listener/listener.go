package listener

import (
	"context"
	"encoding/json"
	"log"

	"event-listener/internal/models"
	"event-listener/internal/publisher"

	"github.com/jackc/pgx/v5"
)

type DBListener struct {
	dbConn    *pgx.Conn
	publisher *publisher.Publisher
}

func NewDBListener(dbURL string, pub *publisher.Publisher) (*DBListener, error) {
	dbConn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		return nil, err
	}

	return &DBListener{
		dbConn:    dbConn,
		publisher: pub,
	}, nil
}

func (l *DBListener) Start() error {
	// Start listening for new event notifications
	_, err := l.dbConn.Exec(context.Background(), "LISTEN new_event")
	if err != nil {
		return err
	}

	log.Println("Listening for new events...")

	for {
		notification, err := l.dbConn.WaitForNotification(context.Background())
		if err != nil {
			log.Printf("Error waiting for notification: %v", err)
			continue
		}

		// When notification received, fetch the event and push to queue
		l.pushEventToQueue(notification.Payload)
	}
}

func (l *DBListener) pushEventToQueue(eventID string) {
	// Fetch the event from database
	var event models.DBEvent
	err := l.dbConn.QueryRow(context.Background(), `
		SELECT id, event_name, event_time::text, payload::text, webhook_id
		FROM events
		WHERE id = $1
	`, eventID).Scan(&event.ID, &event.EventName, &event.EventTime, &event.Payload, &event.WebhookID)

	if err != nil {
		log.Printf("Error fetching event %s: %v", eventID, err)
		return
	}

	// Parse the payload
	var payload struct {
		Subscriber models.Subscriber `json:"subscriber"`
		Segment    *models.Segment   `json:"segment,omitempty"`
	}

	if err := json.Unmarshal([]byte(event.Payload), &payload); err != nil {
		log.Printf("Error unmarshaling payload for event %s: %v", eventID, err)
		return
	}

	// Create webhook event
	webhookEvent := models.WebhookEvent{
		EventName:  event.EventName,
		EventTime:  event.EventTime,
		Subscriber: payload.Subscriber,
		Segment:    payload.Segment,
		WebhookID:  event.WebhookID,
	}

	// Publish to queue
	if err := l.publisher.Publish(webhookEvent); err != nil {
		log.Printf("Error publishing event %s: %v", eventID, err)
		return
	}

	log.Printf("Event %s pushed to queue", eventID)
}

func (l *DBListener) Close() {
	if l.dbConn != nil {
		l.dbConn.Close(context.Background())
	}
}
