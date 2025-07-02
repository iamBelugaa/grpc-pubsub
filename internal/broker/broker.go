package broker

import (
	"context"
	"sync"
	"sync/atomic"

	pubsubpb "github.com/iamBelugaa/grpc-pubsub/internal/generated/__proto__"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Subscriber represents a single subscriber with an ID and a gRPC stream.
type Subscriber struct {
	id     string
	stream grpc.ServerStreamingServer[pubsubpb.PayloadStream]
}

// Service implements the PubSub service with subscriber tracking.
type Service struct {
	mu          sync.RWMutex
	context     context.Context
	log         *zap.SugaredLogger
	subscribers map[string][]*Subscriber
	pubsubpb.UnimplementedPubSubServiceServer
}

// NewService initializes the broker Service.
func NewService(context context.Context, logger *zap.SugaredLogger) *Service {
	return &Service{
		log:         logger,
		context:     context,
		subscribers: make(map[string][]*Subscriber),
	}
}

// Publish sends the payload to all subscribers of the specified topic.
func (s *Service) Publish(ctx context.Context, req *pubsubpb.PublishRequest) (*pubsubpb.PublishResponse, error) {
	s.log.Infow("publish request received", "topic", req.Topic)

	s.mu.RLock()
	subscribers, ok := s.subscribers[req.Topic]
	s.mu.RUnlock()

	// No subscribers found for the topic.
	if !ok {
		return nil, status.Errorf(codes.NotFound, "subscribers with %s topic doesn't exists", req.Topic)
	}

	var errorCount int64
	for _, subscriber := range subscribers {
		s.log.Infow("sending payload stream", "topic", req.Topic, "subscriberId", subscriber.id)

		// Attempt to send payload to the subscriber.
		if err := subscriber.stream.Send(
			&pubsubpb.PayloadStream{Topic: req.Topic, Payload: req.Payload},
		); err != nil {
			atomic.AddInt64(&errorCount, 1)
			s.log.Infow("error send payload stream", "topic", req.Topic, "subscriberId", subscriber.id)
			continue
		}

		s.log.Infow("send payload stream successful", "topic", req.Topic, "subscriberId", subscriber.id)
	}

	if errorCount > 0 {
		return &pubsubpb.PublishResponse{
			Status: ToString(pubsubpb.ResponseStatus_ERROR),
		}, status.Errorf(codes.DataLoss, "failed to send payload to %d streams", errorCount)
	}

	return &pubsubpb.PublishResponse{Status: ToString(pubsubpb.ResponseStatus_OK)}, nil
}
