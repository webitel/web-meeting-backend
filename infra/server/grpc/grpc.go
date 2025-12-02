package server

import (
	"github.com/webitel/web-meeting-backend/infra/server/interceptor"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/webitel/web-meeting-backend/infra/auth"
)

type Server struct {
	*grpc.Server
}

// New provides a new gRPC server.
func New(auth auth.Manager) (*Server, error) {
	otelgrpc.NewServerHandler()
	s := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler(otelgrpc.WithMessageEvents(otelgrpc.SentEvents, otelgrpc.ReceivedEvents))),
		grpc.ChainUnaryInterceptor(
			interceptor.ErrUnaryServerInterceptor(),
			interceptor.AuthUnaryServerInterceptor(auth),
		),
	)

	srv := &Server{s}

	// Register reflection service on gRPC server.
	reflection.Register(srv.Server)

	return srv, nil
}

func (s *Server) Shutdown() error {
	s.Server.GracefulStop()
	return nil
}
