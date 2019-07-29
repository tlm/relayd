package relay

import (
	"testing"
)

func TestHandlerChangeForRelays(t *testing.T) {
	opened, closed, opening, closing, stopped := State("opened"),
		State("closed"),
		State("opening"),
		State("closing"),
		State("stopped")
	transtable := TransitionTable{
		opened:  []State{closing, stopped},
		closed:  []State{opening, stopped},
		opening: []State{closing, stopped},
		closing: []State{opening, stopped},
		stopped: []State{opening, closing},
	}
	relays := []Relay{
		NewDummy("dummy1", opened, transtable),
		NewDummy("dummy2", closed, transtable),
	}

	changeCh, watcher := NewWatcher(relays...)

	stopCh := make(chan struct{})
	go watcher.Run(stopCh)

	for _, relay := range relays {
		from := relay.State()
		to := transtable[from][0]
		if err := relay.Transition(to); err != nil {
			t.Fatalf("unexpected error transitioning relay state from %s to %s",
				from, to)
		}

		ev, ok := <-changeCh
		if !ok {
			t.Fatal("relay event handler testing channel unexpectedly closed")
		}
		if ev.Previous != from {
			t.Fatalf("relay handle event from '%s' does not match expected '%s'",
				ev.Previous, from)
		}
		if ev.Current != to {
			t.Fatalf("relay handle event to '%s' does not match expected '%s'",
				ev.Current, to)
		}
	}

	close(stopCh)
}
