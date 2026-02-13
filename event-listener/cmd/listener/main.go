package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"event-listener/internal/listener"
	"event-listener/internal/publisher"
)

// Configuration
const (
	amqpURL = "amqp://guest:guest@localhost:5672/"
	dbURL   = "postgres://postgres:postgres@localhost:5432/events_db?sslmode=disable"
)

func main() {
	pub, err := publisher.NewPublisher(amqpURL)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer pub.Close()

	dbListener, err := listener.NewDBListener(dbURL, pub)
	if err != nil {
		log.Fatalf("Failed to create listener: %v", err)
	}
	defer dbListener.Close()

	log.Println("Event listener service started")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := dbListener.Start(); err != nil {
			log.Printf("Listener error: %v", err)
		}
	}()

	<-quit
	log.Println("Service stopped")
}
