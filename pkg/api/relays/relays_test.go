package relays

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/tlmiller/relayd/pkg/proto"
	"github.com/tlmiller/relayd/pkg/relay"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

var lis *bufconn.Listener

func bufDialer(string, time.Duration) (net.Conn, error) {
	return lis.Dial()
}

func init() {
	lis = bufconn.Listen(1024 * 1024)
}

func TestListReturnsAllRelays(t *testing.T) {
	opened, closed, opening, closing, stopped := relay.State("opened"),
		relay.State("closed"),
		relay.State("opening"),
		relay.State("closing"),
		relay.State("stopped")
	transtable := relay.TransitionTable{
		opened:  []relay.State{closing, stopped},
		closed:  []relay.State{opening, stopped},
		opening: []relay.State{closing, stopped},
		closing: []relay.State{opening, stopped},
		stopped: []relay.State{opening, closing},
	}
	relays := []relay.Relay{
		relay.NewDummy("dummy1", opened, transtable),
		relay.NewDummy("dummy2", closed, transtable),
	}

	s := grpc.NewServer()
	proto.RegisterRelaysServer(s, NewService(relays))
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Fatalf("serving grpc test server: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("getting grpc test client: %v", err)
	}
	defer conn.Close()
	client := proto.NewRelaysClient(conn)
	listResp, err := client.List(ctx, &proto.ListRelaysRequest{})
	if err != nil {
		t.Fatalf("listing relays: %v", err)
	}

	if len(listResp.GetRelays()) != len(relays) {
		t.Fatalf("expect list relays to return %d relays and not %d",
			len(relays), len(listResp.GetRelays()))
	}

	matchCount := 0
	for _, lr := range listResp.GetRelays() {
		for _, r := range relays {
			if r.Id() == lr.GetId() {
				matchCount++
			}
		}
	}

	if matchCount != len(relays) {
		t.Fatalf("list relays did not match all supplied relays, only %d matched",
			matchCount)
	}
}

func TestTransitionRelaySucceeds(t *testing.T) {
	opened, closed, opening, closing, stopped := relay.State("opened"),
		relay.State("closed"),
		relay.State("opening"),
		relay.State("closing"),
		relay.State("stopped")
	transtable := relay.TransitionTable{
		opened:  []relay.State{closing, stopped},
		closed:  []relay.State{opening, stopped},
		opening: []relay.State{closing, stopped},
		closing: []relay.State{opening, stopped},
		stopped: []relay.State{opening, closing},
	}
	relays := []relay.Relay{
		relay.NewDummy("dummy1", opened, transtable),
	}

	s := grpc.NewServer()
	proto.RegisterRelaysServer(s, NewService(relays))
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Fatalf("serving grpc test server: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("getting grpc test client: %v", err)
	}
	defer conn.Close()
	client := proto.NewRelaysClient(conn)

	for _, r := range relays {
		toState := transtable[r.State()][0]
		_, err := client.Transition(ctx, &proto.TransitionRelayRequest{
			RelayId: r.Id(),
			ToState: string(toState),
		})

		if err != nil {
			t.Errorf("failed transition relay %s to %s: %v", r.Id(), toState, err)
		}

		if curr := r.State(); curr != toState {
			t.Errorf("unexpected relay state after transition, expected %s got %s",
				toState, curr)
		}
	}
}

func TestTransitionRelayFailsForInvalidArguments(t *testing.T) {
	tests := []proto.TransitionRelayRequest{
		// Invalid state
		proto.TransitionRelayRequest{
			RelayId: "4231",
			ToState: "",
		},
		proto.TransitionRelayRequest{
			RelayId: "",
			ToState: "madeUpState",
		},
		proto.TransitionRelayRequest{
			RelayId: "",
			ToState: "",
		},
	}

	s := grpc.NewServer()
	proto.RegisterRelaysServer(s, NewService([]relay.Relay{}))
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Fatalf("serving grpc test server: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("getting grpc test client: %v", err)
	}
	defer conn.Close()
	client := proto.NewRelaysClient(conn)

	for i, test := range tests {
		_, err := client.Transition(ctx, &test)
		if err == nil {
			t.Fatalf("expected error for bad transition arguments, test %d", i)
		}

		if status.Code(err) != codes.InvalidArgument {
			t.Errorf("unexpected error code %s for test %d, watned InvalidArgument",
				status.Code(err), i)
		}
	}
}

func TestTransitionNotFoundRelayId(t *testing.T) {
	s := grpc.NewServer()
	proto.RegisterRelaysServer(s, NewService([]relay.Relay{}))
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Fatalf("serving grpc test server: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("getting grpc test client: %v", err)
	}
	defer conn.Close()
	client := proto.NewRelaysClient(conn)

	_, err = client.Transition(ctx, &proto.TransitionRelayRequest{
		RelayId: "dummy",
		ToState: "madeUpState",
	})

	if err == nil {
		t.Fatal("expected error for bad transition relay id not found")
	}

	if status.Code(err) != codes.NotFound {
		t.Errorf("unexpected error code %s, wanted NotFound", status.Code(err))
	}
}

func TestWatchReturnsRelayTransitions(t *testing.T) {
	opened, closed, opening, closing, stopped := relay.State("opened"),
		relay.State("closed"),
		relay.State("opening"),
		relay.State("closing"),
		relay.State("stopped")
	transtable := relay.TransitionTable{
		opened:  []relay.State{closing, stopped},
		closed:  []relay.State{opening, stopped},
		opening: []relay.State{closing, stopped},
		closing: []relay.State{opening, stopped},
		stopped: []relay.State{opening, closing},
	}
	relays := []relay.Relay{
		relay.NewDummy("dummy1", opened, transtable),
	}

	s := grpc.NewServer()
	proto.RegisterRelaysServer(s, NewService(relays))
	go func() {
		if err := s.Serve(lis); err != nil {
			t.Fatalf("serving grpc test server: %v", err)
		}
	}()

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("getting grpc test client: %v", err)
	}
	defer conn.Close()
	client := proto.NewRelaysClient(conn)

	watchClient, err := client.Watch(ctx, &proto.WatchRelaysRequest{})
	if err != nil {
		t.Fatalf("unexpected error when getting watch relays client: %v", err)
	}

	// Transition the state to it's next available one
	relays[0].Transition(transtable[relays[0].State()][0])

	msg, err := watchClient.Recv()
	if err != nil {
		t.Fatalf("unxepected error when recieving watch relay updates: %v", err)
	}

	if msg.GetRelay().GetId() != relays[0].Id() {
		t.Fatalf("watched relay response id did not match that of changed, got %s expected %s",
			msg.GetRelay().GetId(), relays[0].Id())
	}

	if msg.GetRelay().GetState() != string(relays[0].State()) {
		t.Fatalf("watched relay response state did not match that of changed, got %s expected %s",
			msg.GetRelay().GetState(), relays[0].State())
	}
}
