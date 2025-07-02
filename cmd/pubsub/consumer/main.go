package main

import (
	"context"
	"time"

	"github.com/iamBelugaa/grpc-pubsub/internal/consumer"
	"github.com/iamBelugaa/grpc-pubsub/pkg/logger"
)

const (
	// Address of the gRPC broker to connect to.
	brokerURL string = "localhost:8080"
)

func main() {
	// Initialize logger for consumer.
	logger := logger.New("grpc-pubsub:consumer")

	// Create a new consumer client.
	consumer, err := consumer.New(logger, brokerURL)
	if err != nil {
		logger.Fatalw("create consumer error", "error", err)
	}

	// Ensure the consumer closes properly.
	defer func() {
		if err := consumer.Close(); err != nil {
			logger.Fatalw("consumer close error", "error", err)
		}
	}()

	// Subscribe to topics.
	consumer.Subscribe(context.Background(), "kubernetes")
	time.Sleep(time.Second * 2)

	// Unsubscribe from a topic.
	consumer.Unsubscribe(context.Background(), "kubernetes")
	time.Sleep(time.Second * 3)
}
