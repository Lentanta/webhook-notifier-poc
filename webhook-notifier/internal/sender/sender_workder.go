package sender

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"webhook-notifier/internal/models"

	amqp "github.com/rabbitmq/amqp091-go"
)

func failOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func SenderWorker(
	id int,
	qConn *amqp.Connection,
	prefetchCount int,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	// Create a channel for each worker
	ch, err := qConn.Channel()
	errMsg := fmt.Sprintf("Worker %v: Failed to open channel", id)
	failOnError(err, errMsg)
	defer ch.Close()

	// Prefetch — only take 3 messages at a time
	err = ch.Qos(prefetchCount, 0, false)
	errMsg = fmt.Sprintf("Worker %v: Failed to set Prefetch", id)
	failOnError(err, errMsg)

	// Declare the queue with name in mock server
	queue, err := ch.QueueDeclare(
		"webhook_queue", // name
		false,           // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	errMsg = fmt.Sprintf("Worker %v: Failed to declare a queue", id)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		queue.Name,
		"",
		false, // autoAck = false
		false, false, false, nil,
	)
	errMsg = fmt.Sprintf("Worker %d: failed to start consumer", id)
	failOnError(err, errMsg)

	fmt.Println("===== Worker ", id, " started =====")

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

		err := ProcessSendWebhook(qMessage.Content)
		if err != nil {
			fmt.Print(err)
			msg.Nack(false, false)
		} else {
			log.Printf("===== Worker %d send succesfully =====", id)
			msg.Ack(false)
		}
	}

	log.Printf("Worker %d: channel closed, shutting down", id)
}
