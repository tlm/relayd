package relays

import (
	"log"

	"github.com/tlmiller/relayd/pkg/proto"
	"github.com/tlmiller/relayd/pkg/relay"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type WatchHandler interface {
	Watch(*proto.WatchRelaysRequest, proto.Relays_WatchServer) error
}

type defaultWatchHandler struct {
	relays []relay.Relay
}

func NewWatchHandler(r []relay.Relay) WatchHandler {
	return &defaultWatchHandler{
		relays: r,
	}
}

func (h *defaultWatchHandler) Watch(_ *proto.WatchRelaysRequest, s proto.Relays_WatchServer) error {
	changeCh, watcher := relay.NewWatcher(h.relays...)
	stopCh := make(chan struct{})
	defer close(stopCh)
	go watcher.Run(stopCh)

	for {
		select {
		case <-s.Context().Done():
			return nil
		case change := <-changeCh:
			err := s.Send(&proto.WatchRelaysResponse{
				Relay: relay.RelayToProto(change.R),
			})
			if err != nil {
				log.Printf("[info] sending relay watch change to client: %v", err)
				return status.Error(codes.Internal, "sending relay watch response")
			}
		}
	}
}
