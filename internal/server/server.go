package server

import (
	"abf/internal/service"
	"abf/internal/service/pb"
	"context"
	"errors"
	"log"
	"net"
	"net/http"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	service Service
	server  *grpc.Server
	config  Config
}

type Service interface {
	Try(ctx context.Context, request *pb.CheckRequest) (*pb.Status, error)
}

type Config interface {
	GetGRPCAddr() string
}

func NewServer(service *service.Service, config Config) *Server {
	grpcServer := grpc.NewServer()
	pb.RegisterAntiBruteForceServer(grpcServer, service)

	return &Server{
		server:  grpcServer,
		service: service,
		config:  config,
	}
}

func (s *Server) Start(ctx context.Context) error {
	log.Println("starting server on ", s.config.GetGRPCAddr())

	lsn, err := net.Listen("tcp", s.config.GetGRPCAddr())
	if err != nil {
		log.Printf("fail start gprc server: %e", err)
	}

	reflection.Register(s.server)
	if err := s.server.Serve(lsn); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Printf("listen: %e", err)
	}

	<-ctx.Done()

	return nil
}

func (s *Server) Stop() {
	s.server.GracefulStop()
}
