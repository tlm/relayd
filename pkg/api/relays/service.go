package relays

import (
	"github.com/tlmiller/relayd/pkg/proto"
	"github.com/tlmiller/relayd/pkg/relay"

	"google.golang.org/grpc"
)

type Service struct {
	ListHandler
	TransitionHandler
	WatchHandler
}

func NewService(r []relay.Relay) *Service {
	return &Service{
		ListHandler:       NewListHandler(r),
		TransitionHandler: NewTransitionHandler(r),
		WatchHandler:      NewWatchHandler(r),
	}
}

func (s *Service) Register(server *grpc.Server) {
	proto.RegisterRelaysServer(server, s)
}
