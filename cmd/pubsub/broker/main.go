package main

import (
	"github.com/iamBelugaa/grpc-pubsub/internal/broker"
	"github.com/iamBelugaa/grpc-pubsub/internal/server"
	"github.com/iamBelugaa/grpc-pubsub/pkg/logger"
)

const (
	// Address for the gRPC server.
	serverURL string = "localhost:8080"
)

func main() {
	// Initialize logger for broker.
	logger := logger.New("grpc-pubsub:broker")

	// Create a new gRPC server.
	context, server := server.New(serverURL, logger)

	// Create a new broker.
	broker := broker.New(context, logger)

	// Start the server.
	if err := server.ListenAndServe(broker); err != nil {
		logger.Infow("listen error", "error", err)
	}
}
