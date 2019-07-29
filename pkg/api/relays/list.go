package relays

import (
	"context"

	"github.com/tlmiller/relayd/pkg/proto"
	"github.com/tlmiller/relayd/pkg/relay"
)

type ListHandler interface {
	List(context.Context, *proto.ListRelaysRequest) (*proto.ListRelaysResponse, error)
}

type defaultListHandler struct {
	relays []relay.Relay
}

func NewListHandler(r []relay.Relay) ListHandler {
	return &defaultListHandler{
		relays: r,
	}
}

func (r *defaultListHandler) List(_ context.Context, _ *proto.ListRelaysRequest) (
	*proto.ListRelaysResponse, error) {
	rResp := make([]*proto.Relay, len(r.relays))
	for i, r := range r.relays {
		rResp[i] = relay.RelayToProto(r)
	}
	return &proto.ListRelaysResponse{Relays: rResp}, nil
}
