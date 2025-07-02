package publisher

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/iamBelugaa/grpc-pubsub/internal/broker"
	pubsubpb "github.com/iamBelugaa/grpc-pubsub/internal/generated/__proto__"
)

// publisher represents the publisher service that connects to the gRPC broker.
type publisher struct {
	log    *zap.SugaredLogger
	conn   *grpc.ClientConn
	client pubsubpb.PubSubServiceClient
}

// New creates and returns a new publisher service connected to the given broker address.
func New(logger *zap.SugaredLogger, brokerAddr string) (*publisher, error) {
	logger.Infow("creating publisher client", "brokerAddr", brokerAddr)

	// Establish a gRPC connection using insecure credentials (no TLS).
	conn, err := grpc.NewClient(brokerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("create client error : %w", err)
	}

	logger.Infow("publisher client connection established", "brokerAddr", brokerAddr)

	// Create a new PubSub client using the connection.
	client := pubsubpb.NewPubSubServiceClient(conn)
	return &publisher{log: logger, conn: conn, client: client}, nil
}

// Publish sends a message with the given topic and payload to the broker.
func (s *publisher) Publish(context context.Context, topic string, payload []byte) {
	s.log.Infow("publish request initiated", "topic", topic)

	response, err := s.client.Publish(context, &pubsubpb.PublishRequest{
		Topic:   topic,
		Payload: payload,
	})

	// Handle any error returned from the broker.
	if err != nil {
		if status, ok := status.FromError(err); ok {
			s.log.Infow(
				"failed to publish", "statusCode", status.Code(), "message", status.Message(), "topic", topic,
			)
		} else {
			s.log.Infow("failed to publish", "error", err, "topic", topic)
		}
		return
	}

	// Check for an application-level error response.
	if response.Status == broker.ToString(pubsubpb.ResponseStatus_ERROR) {
		s.log.Infow("failed to publish", "error", err, "topic", topic)
		return
	}

	s.log.Infow("published successfully", "topic", topic)
}

// Close cleanly shuts down the gRPC connection.
func (s *publisher) Close() error {
	s.log.Infow("closing publisher service")
	if err := s.conn.Close(); err != nil {
		return err
	}

	s.log.Infow("closed publisher service successfully")
	return nil
}
