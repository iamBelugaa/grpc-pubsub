package main

import (
	"context"
	"time"

	"github.com/iamBelugaa/grpc-pubsub/internal/publisher"
	"github.com/iamBelugaa/grpc-pubsub/pkg/logger"
)

const (
	// Address of the gRPC broker to connect to.
	brokerURL string = "localhost:8080"
)

func main() {
	// Initialize logger for publisher.
	logger := logger.New("grpc-pubsub:publisher")

	// Create a new publisher client.
	publisher, err := publisher.New(logger, brokerURL)
	if err != nil {
		logger.Fatalw("create publisher error", "error", err)
	}

	// Ensure the publisher is closed.
	defer func() {
		if err := publisher.Close(); err != nil {
			logger.Fatalw("publisher close error", "error", err)
		}
	}()

	time.Sleep(time.Second * 2)

	message := []byte("New release: Kubernetes 1.30 is out")
	publisher.Publish(context.Background(), "kubernetes", message)
	publisher.Publish(context.Background(), "kubernetes", message)
	publisher.Publish(context.Background(), "kubernetes", message)

	time.Sleep(time.Second * 4)
	publisher.Publish(context.Background(), "kubernetes", message)
	publisher.Publish(context.Background(), "kubernetes", message)
	publisher.Publish(context.Background(), "kubernetes", message)
}
