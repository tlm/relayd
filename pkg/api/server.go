package api

import (
	"net"

	"github.com/tlmiller/relayd/pkg/api/relays"

	"google.golang.org/grpc"
)

type Server struct {
	GRPCServer *grpc.Server
	Listener   net.Listener
}

func NewServer() (*Server, error) {
	listener, err := net.Listen("tcp", ":5632")
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()
	relays.NewService().Register(grpcServer)

	return &Server{
		GRPCServer: grpcServer,
		Listener:   listener,
	}, nil
}

func (s *Server) Serve() error {
	return s.GRPCServer.Serve(s.Listener)
}
