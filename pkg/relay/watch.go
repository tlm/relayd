package relay

import (
	"time"
)

type StateChange struct {
	Current  State
	Previous State
	R        Relay
}

type Watcher interface {
	Run(<-chan struct{})
}

type WatchFn func(stopCh <-chan struct{})

func (w WatchFn) Run(s <-chan struct{}) {
	w(s)
}

func NewWatcher(relays ...Relay) (<-chan StateChange, Watcher) {
	changeCh := make(chan StateChange)
	knownStates := make([]State, len(relays))
	for i, r := range relays {
		knownStates[i] = r.State()
	}

	return changeCh, WatchFn(func(stopCh <-chan struct{}) {
		for {
			for i, relay := range relays {
				if curr := relay.State(); curr != knownStates[i] {
					// State of this relay has changed
					changeCh <- StateChange{
						Current:  curr,
						Previous: knownStates[i],
						R:        relay,
					}
					knownStates[i] = curr
				}
				select {
				case <-stopCh:
					close(changeCh)
					goto finish
				default:
				}
			}
			time.Sleep(1 * time.Second)
		}
	finish:
	})
}
