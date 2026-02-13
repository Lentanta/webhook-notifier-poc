package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Configuration and setup for prometheus

var (
	MessagesReceived = promauto.NewCounter(prometheus.CounterOpts{
		Name: "webhook_messages_received_total",
		Help: "Total number of messages received from RabbitMQ queue",
	})

	MessagesProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "webhook_messages_processed_total",
		Help: "Total number of messages processed successfully",
	})

	MessagesFailed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "webhook_messages_failed_total",
		Help: "Total number of messages that failed processing",
	})
)
