package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
	"webhook-notifier/internal/models"
	"webhook-notifier/internal/sender"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	// --- Connect to the queue
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	// --- Connect to the channel
	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	// --- declare the queue with name in mock server
	q, err := ch.QueueDeclare(
		"webhook_queue", // name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	failOnError(err, "Failed to declare a queue")

	// --- register a consumer
	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	quit := make(chan os.Signal, 1)
	go func() {
		for msg := range msgs {
			log.Printf("Received a message: %s", msg.Body)

			// Get data from message
			jsonData := []byte(msg.Body)
			var qMessage models.QMessage
			if err := json.Unmarshal(jsonData, &qMessage); err != nil {
				log.Printf("Cannot unmarshal message data")
				msg.Nack(false, false) // Don't requeue
				continue
			}

			err := sender.ProcessSendWebhook(qMessage.Content)
			if err != nil {
				fmt.Print(err)
				msg.Nack(false, false) // Don't requeue
			} else {
				msg.Ack(false)
			}

		}
	}()

	log.Printf("Service is running")
	<-quit

	// shutting down gracefully
	time.Sleep(3 * time.Second)

	ch.Close()
	conn.Close()
	log.Printf("Everything closed")
}
