package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	pubsubpb "github.com/iamBelugaa/grpc-pubsub/internal/generated/__proto__"
)

// server defines the structure for the gRPC server instance.
type server struct {
	addr     string
	listener net.Listener
	server   *grpc.Server
	context  context.Context
	cancel   context.CancelFunc
	log      *zap.SugaredLogger
}

// New initializes and returns a new Service instance with context.
func New(addr string, logger *zap.SugaredLogger) (context.Context, *server) {
	context, cancel := context.WithCancel(context.Background())
	return context, &server{addr: addr, cancel: cancel, context: context, log: logger}
}

// ListenAndServe starts the gRPC server and handles graceful shutdown.
func (s *server) ListenAndServe(pubsubService pubsubpb.PubSubServiceServer) error {
	s.log.Infow("creating grpc server", "addr", s.addr)

	// Create new gRPC server.
	server := grpc.NewServer()

	// Start listening on the configured address.
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.server = server
	s.listener = listener
	s.log.Infow("grpc server created successfully", "addr", s.addr)

	// Register the PubSub service with the gRPC server
	pubsubpb.RegisterPubSubServiceServer(server, pubsubService)
	s.log.Infow("pubsub service registered")

	// Channels to handle server errors and OS shutdown signals.
	serverErrors := make(chan error, 1)
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	// Run the gRPC server.
	go func() {
		s.log.Infow("starting grpc server", "addr", s.addr)
		if err := s.server.Serve(s.listener); err != nil {
			serverErrors <- err
		}
	}()

	// Wait for server error or shutdown signal.
	select {
	case err := <-serverErrors:
		s.cancel()
		s.log.Infow("received server error", "error", err)
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdownCh:
		s.cancel()
		s.log.Infow("shutting down server signal received", "signal", sig)

		if err := s.listener.Close(); err != nil {
			return fmt.Errorf("could not stop listener gracefully: %w", err)
		}
		s.server.GracefulStop()

		s.log.Infow("shutdown complete with signal", "signal", sig)
	}

	s.cancel()
	return nil
}

// Stop gracefully shuts down the server and cleans up resources.
func (s *server) Stop() error {
	s.cancel()
	s.log.Infow("shutting down server", "addr", s.addr)

	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("could not stop listener gracefully: %w", err)
	}
	s.server.GracefulStop()

	s.log.Infow("shutdown complete", "addr", s.addr)
	return nil
}
