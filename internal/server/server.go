package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	pubsubpb "github.com/iamBelugaa/grpc-pubsub/internal/generated/__proto__"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Service struct {
	addr     string
	listener net.Listener
	server   *grpc.Server
	context  context.Context
	cancel   context.CancelFunc
	log      *zap.SugaredLogger
}

func NewService(addr string, logger *zap.SugaredLogger) (context.Context, *Service) {
	context, cancel := context.WithCancel(context.Background())
	return context, &Service{addr: addr, cancel: cancel, context: context}
}

func (s *Service) ListenAndServe(pubsubService pubsubpb.PubSubServiceServer) error {
	s.log.Infow("creating grpc server", "addr", s.addr)

	server := grpc.NewServer()
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.server = server
	s.listener = listener

	s.log.Infow("grpc server created successfully", "addr", s.addr)

	pubsubpb.RegisterPubSubServiceServer(server, pubsubService)
	s.log.Infow("pubsub service registered")

	serverErrors := make(chan error, 1)
	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		s.log.Infow("starting grpc server", "addr", s.addr)
		if err := s.server.Serve(s.listener); err != nil {
			serverErrors <- err
		}
	}()

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

func (s *Service) Stop() error {
	s.cancel()
	s.log.Infow("shutting down server", "addr", s.addr)

	if err := s.listener.Close(); err != nil {
		return fmt.Errorf("could not stop listener gracefully: %w", err)
	}
	s.server.GracefulStop()

	s.log.Infow("shutdown complete", "addr", s.addr)
	return nil
}
