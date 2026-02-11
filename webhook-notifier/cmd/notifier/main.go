package main

import (
	"log"
	"os"
	"sync"
	"time"
	"webhook-notifier/internal/sender"

	amqp "github.com/rabbitmq/amqp091-go"
)

const numWorkers = 3

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

	var wg sync.WaitGroup
	for i := range numWorkers {
		wg.Add(1)
		id := i + 1 // For beautifull log :)
		go sender.SenderWorker(id, conn, 2, &wg)
	}

	quit := make(chan os.Signal, 1)
	log.Printf("Service is running")
	<-quit

	// shutting down gracefully
	time.Sleep(3 * time.Second)
	conn.Close()
	log.Printf("Everything closed")
}
