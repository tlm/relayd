package relay

import (
	"github.com/tlmiller/relayd/pkg/proto"
)

type Relay interface {
	Id() string
	State() State
	Transitions() TransitionTable
	Transition(State) error
}

func RelayToProto(r Relay) *proto.Relay {
	trans := []*proto.Transition{}
	for k, v := range r.Transitions() {
		trans = append(trans, &proto.Transition{
			FromState: string(k),
			ToStates:  StateSliceToStrings(v),
		})
	}
	return &proto.Relay{
		Id:          r.Id(),
		State:       string(r.State()),
		Transitions: trans,
	}
}
