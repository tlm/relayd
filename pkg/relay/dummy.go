package relay

import (
	"sync"
)

type Dummy struct {
	lock       sync.Mutex
	id         string
	state      State
	transTable TransitionTable
}

func NewDummy(id string, state State, table TransitionTable) *Dummy {
	return &Dummy{
		id:         id,
		state:      state,
		transTable: table,
	}
}

func (d *Dummy) Id() string {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.id
}

func (d *Dummy) State() State {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.state
}

func (d *Dummy) Transitions() TransitionTable {
	return d.transTable
}

func (d *Dummy) Transition(s State) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	states, exists := d.transTable[d.state]
	if !exists {
		return ErrUnknownState(d.state)
	}
	if !StateInSlice(s, states) {
		return ErrNoTransition{
			Available: states,
			From:      d.state,
			To:        s,
		}
	}
	d.state = s
	return nil
}
