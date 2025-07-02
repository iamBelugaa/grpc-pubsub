package main

import (
	"github.com/iamBelugaa/grpc-pubsub/internal/broker"
	"github.com/iamBelugaa/grpc-pubsub/internal/server"
	"github.com/iamBelugaa/grpc-pubsub/pkg/logger"
)

const (
	baseURL string = "localhost:8080"
)

func main() {
	logger := logger.New("grpc-pubsub")

	context, server := server.NewService(baseURL, logger)
	broker := broker.NewService(context, logger)

	if err := server.ListenAndServe(broker); err != nil {
		logger.Infow("listen error", "error", err)
	}
}
