package broker

import (
	"context"
	"slices"
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

// Subscribe adds the subscriber to the topic's list and streams payloads until disconnected.
func (s *Service) Subscribe(req *pubsubpb.SubscribeRequest, stream grpc.ServerStreamingServer[pubsubpb.PayloadStream]) error {
	s.log.Infow("subscribe request received", "subscriberId", req.SubscriberId, "topic", req.Topic)

	// Check if subscriber already exists.
	s.mu.RLock()
	exists := slices.ContainsFunc(s.subscribers[req.Topic], func(e *Subscriber) bool {
		return e.id == req.SubscriberId
	})
	s.mu.RUnlock()

	if exists {
		s.log.Infow("already subscribed to topic", "subscriberId", req.SubscriberId, "topic", req.Topic)
		return nil
	}

	// Register new subscriber.
	s.mu.Lock()
	subscriber := &Subscriber{id: req.SubscriberId, stream: stream}
	s.subscribers[req.Topic] = append(s.subscribers[req.Topic], subscriber)
	s.mu.Unlock()

	s.log.Infow("subscriber add to list", "subscriberId", req.SubscriberId, "topic", req.Topic)

	// Wait for disconnection or shutdown.
	for {
		select {
		case <-stream.Context().Done():
			s.log.Infow("stream ended", "subscriberId", req.SubscriberId, "topic", req.Topic)
			return stream.Context().Err()

		case <-s.context.Done():
			s.log.Infow("broker closed", "subscriberId", req.SubscriberId, "topic", req.Topic)
			return s.context.Err()
		}
	}
}

// Unsubscribe removes a subscriber from the topic.
func (s *Service) Unsubscribe(ctx context.Context, req *pubsubpb.UnsubscribeRequest) (*pubsubpb.UnsubscribeResponse, error) {
	s.log.Infow("unsubscribe request received", "subscriberId", req.SubscriberId, "topic", req.Topic)

	s.mu.RLock()
	subscribers := s.subscribers[req.Topic]
	index := slices.IndexFunc(subscribers, func(e *Subscriber) bool {
		return e.id == req.SubscriberId
	})
	s.mu.RUnlock()

	if index == -1 {
		s.log.Infow("subscriber not found", "subscriberId", req.SubscriberId, "topic", req.Topic)
		return nil, status.Errorf(codes.NotFound, "subscriber with id %s doesn't exists", req.SubscriberId)
	}

	// Remove subscriber.
	s.mu.Lock()
	s.subscribers[req.Topic] = slices.Delete(subscribers, index, index+1)
	s.mu.Unlock()

	s.log.Infow("unsubscribed from topic", "subscriberId", req.SubscriberId, "topic", req.Topic)
	return &pubsubpb.UnsubscribeResponse{Status: ToString(pubsubpb.ResponseStatus_OK)}, nil
}
