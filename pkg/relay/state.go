package relay

import (
	"fmt"
)

type ErrNoTransition struct {
	Available []State
	From      State
	To        State
}

type ErrUnknownState State

type State string

type TransitionTable map[State][]State

var (
	StateNone = State("")
)

func StateInSlice(state State, sli []State) bool {
	for _, s := range sli {
		if s == state {
			return true
		}
	}
	return false
}

func (e ErrNoTransition) Error() string {
	return fmt.Sprintf("cannot tranisition from %s to %s, acceptable transitions %v",
		string(e.From), string(e.To), e.Available)
}

func (e ErrUnknownState) Error() string {
	return fmt.Sprintf("unknown state %s", string(e))
}

func StateSliceToStrings(s []State) []string {
	r := make([]string, len(s))
	for i, t := range s {
		r[i] = string(t)
	}
	return r
}
