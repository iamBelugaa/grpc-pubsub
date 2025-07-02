package consumer

import (
	"context"
	"fmt"
	"io"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/iamBelugaa/grpc-pubsub/internal/broker"
	pubsubpb "github.com/iamBelugaa/grpc-pubsub/internal/generated/__proto__"
)

// consumer represents a consumer that subscribes to topics via a gRPC broker.
type consumer struct {
	id     string
	log    *zap.SugaredLogger
	conn   *grpc.ClientConn
	client pubsubpb.PubSubServiceClient
}

// New creates a new consumer service and establishes a gRPC connection to the broker.
func New(logger *zap.SugaredLogger, brokerAddr string) (*consumer, error) {
	logger.Infow("creating consumer client", "brokerAddr", brokerAddr)

	// Create a gRPC connection using insecure credentials (non-TLS).
	conn, err := grpc.NewClient(brokerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("create consumer client error : %w", err)
	}

	logger.Infow("consumer client connection established", "brokerAddr", brokerAddr)

	// Create a new PubSub client using the connection.
	client := pubsubpb.NewPubSubServiceClient(conn)
	return &consumer{id: uuid.New().String(), log: logger, conn: conn, client: client}, nil
}

// Subscribe connects to a topic and starts reading messages from the stream.
func (s *consumer) Subscribe(context context.Context, topic string) {
	s.log.Infow("subscribe request initiated", "topic", topic, "subscriberId", s.id)

	stream, err := s.client.Subscribe(context, &pubsubpb.SubscribeRequest{
		SubscriberId: s.id,
		Topic:        topic,
	})

	// Handle subscription errors.
	if err != nil {
		if status, ok := status.FromError(err); ok {
			s.log.Infow(
				"failed to subscribe",
				"statusCode", status.Code(),
				"message", status.Message(),
				"topic", topic, "subscriberId", s.id,
			)
		} else {
			s.log.Infow("failed to subscribe", "error", err, "topic", topic, "subscriberId", s.id)
		}
		return
	}

	// Start reading from the stream.
	go s.readFromStream(context, stream)
}

// Unsubscribe disconnects from a topic.
func (s *consumer) Unsubscribe(context context.Context, topic string) {
	s.log.Infow("unsubscribe request initiated", "topic", topic, "subscriberId", s.id)

	response, err := s.client.Unsubscribe(context, &pubsubpb.UnsubscribeRequest{
		SubscriberId: s.id,
		Topic:        topic,
	})

	// Handle unsubscription errors.
	if err != nil {
		if status, ok := status.FromError(err); ok {
			s.log.Infow(
				"failed to unsubscribe",
				"statusCode", status.Code(),
				"message", status.Message(),
				"topic", topic, "subscriberId", s.id,
			)
		} else {
			s.log.Infow("failed to unsubscribe", "error", err, "topic", topic, "subscriberId", s.id)
		}
		return
	}

	if response.Status == broker.ToString(pubsubpb.ResponseStatus_ERROR) {
		s.log.Infow("failed to unsubscribe", "error", err, "topic", topic, "subscriberId", s.id)
		return
	}

	s.log.Infow("unsubscribed successfully", "topic", topic, "subscriberId", s.id)
}

// Close cleanly shuts down the gRPC connection.
func (s *consumer) Close() error {
	s.log.Infow("closing consumer service", "subscriberId", s.id)
	if err := s.conn.Close(); err != nil {
		return err
	}

	s.log.Infow("closed consumer service successfully", "subscriberId", s.id)
	return nil
}

// readFromStream continuously reads messages from the gRPC stream.
func (s *consumer) readFromStream(context context.Context, stream grpc.ServerStreamingClient[pubsubpb.PayloadStream]) {
	for {
		select {
		case <-context.Done():
			s.log.Infow("context cancelled - stopping stream read", "subscriberId", s.id)
			stream.CloseSend()
			s.log.Infow("stream closed after context cancellation", "subscriberId", s.id)

		default:
			// Receive the next message from the stream.
			msg, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					s.log.Infow("stream closed by server", "subscriberId", s.id)
					return
				} else {
					s.log.Infow("error receiving message from stream", "error", err, "subscriberId", s.id)
					continue
				}
			}

			s.log.Infow("message received", "topic", msg.Topic, "payload", string(msg.Payload), "subscriberId", s.id)
		}
	}
}
