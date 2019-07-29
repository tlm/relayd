package relays

import (
	"context"
	"log"

	"github.com/tlmiller/relayd/pkg/proto"
	"github.com/tlmiller/relayd/pkg/relay"

	"google.golang.org/genproto/googleapis/rpc/errdetails"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type TransitionHandler interface {
	Transition(context.Context, *proto.TransitionRelayRequest) (*proto.TransitionRelayResponse, error)
}

type defaultTransitionHandler struct {
	relays map[string]relay.Relay
}

func NewTransitionHandler(relays []relay.Relay) TransitionHandler {
	defHandler := &defaultTransitionHandler{
		relays: make(map[string]relay.Relay, len(relays)),
	}
	for _, r := range relays {
		defHandler.relays[r.Id()] = r
	}
	return defHandler
}

func (h *defaultTransitionHandler) Transition(_ context.Context, req *proto.TransitionRelayRequest) (
	*proto.TransitionRelayResponse, error) {
	invalidFields := make([]*errdetails.BadRequest_FieldViolation, 0)
	if req.GetToState() == "" {
		v := &errdetails.BadRequest_FieldViolation{
			Field:       "to_state",
			Description: "cannot be empty",
		}
		invalidFields = append(invalidFields, v)
	}

	if req.GetRelayId() == "" {
		v := &errdetails.BadRequest_FieldViolation{
			Field:       "relay_id",
			Description: "cannot be empty",
		}
		invalidFields = append(invalidFields, v)
	}

	res := &proto.TransitionRelayResponse{}

	if len(invalidFields) != 0 {
		st := status.New(codes.InvalidArgument, "invalid arguments")
		br := &errdetails.BadRequest{}
		br.FieldViolations = append(br.FieldViolations, invalidFields...)
		st, err := st.WithDetails(br)
		if err != nil {
			log.Printf("[error] constructing transition invalid fields status")
			return res, status.Error(codes.Internal, "sending transition relay response")
		}
		return res, st.Err()
	}

	reqRelay, found := h.relays[req.GetRelayId()]
	if !found {
		return res, status.Errorf(codes.NotFound, "relayd for id %s not found", req.GetRelayId())
	}

	toState := relay.State(req.GetToState())
	if err := reqRelay.Transition(toState); err != nil {
		return res, status.Errorf(codes.FailedPrecondition,
			"relay with id %s unable to transition to %s: %v", req.GetRelayId(),
			toState, err)
	}
	return res, nil
}
