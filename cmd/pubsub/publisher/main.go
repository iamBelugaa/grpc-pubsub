package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

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

	fmt.Print("Enter your username: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	username := scanner.Text()

	for {
		fmt.Print("Type messages (press Enter to send, 'quit' to exit): ")
		scanner.Scan()
		text := scanner.Text()
		println()

		if text == "quit" {
			break
		}

		message := fmt.Sprintf("%s: %s", username, text)
		publisher.Publish(context.Background(), "group:chat", []byte(message))
	}
}
