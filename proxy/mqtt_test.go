package proxy

import (
	"context"
	"encoding/json"
	"net"
	"net/url"
	"sync"
	"testing"
	"time"

	bmwcardata "github.com/tjamet/bmw-cardata"
)

type fakePublisher struct {
	mu       sync.Mutex
	topics   []string
	payloads [][]byte
}

func (f *fakePublisher) Publish(topic string, payload []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.topics = append(f.topics, topic)
	f.payloads = append(f.payloads, payload)
	return nil
}

func TestBridgeHandleRepublishes(t *testing.T) {
	pub := &fakePublisher{}
	var seen bmwcardata.StreamedMessage
	br := &Bridge{
		Publisher:   pub,
		TopicPrefix: "cardata/",
		OnMessage:   func(m bmwcardata.StreamedMessage) { seen = m },
	}

	msg := bmwcardata.StreamedMessage{VIN: "WBA9", EntityID: "e1", Topic: "t"}
	br.handle(msg)

	if seen.VIN != "WBA9" {
		t.Fatalf("OnMessage not invoked with the message, got %+v", seen)
	}
	if len(pub.topics) != 1 || pub.topics[0] != "cardata/WBA9" {
		t.Fatalf("unexpected republish topics: %v", pub.topics)
	}
	var decoded bmwcardata.StreamedMessage
	if err := json.Unmarshal(pub.payloads[0], &decoded); err != nil {
		t.Fatalf("payload is not a StreamedMessage: %s", err)
	}
	if decoded.VIN != "WBA9" || decoded.EntityID != "e1" {
		t.Fatalf("unexpected republished payload: %+v", decoded)
	}
}

// freeAddr returns a currently-free loopback TCP address.
func freeAddr(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("reserve port: %s", err)
	}
	addr := l.Addr().String()
	_ = l.Close()
	return addr
}

func TestBrokerEndToEnd(t *testing.T) {
	cases := []struct {
		name    string
		users   map[string]string
		profile bmwcardata.StaticStreamingProfile
	}{
		{name: "allow-all", profile: bmwcardata.StaticStreamingProfile{}},
		{
			name:    "user-password",
			users:   map[string]string{"local": "secret"},
			profile: bmwcardata.StaticStreamingProfile{Username: "local", Password: "secret"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			addr := freeAddr(t)
			broker, err := NewBroker(BrokerOptions{Address: addr, Users: tc.users})
			if err != nil {
				t.Fatalf("NewBroker: %s", err)
			}
			serveErr := make(chan error, 1)
			go func() { serveErr <- broker.Serve() }()
			t.Cleanup(func() { _ = broker.Close() })

			streamURL, err := url.Parse("mqtt://" + addr)
			if err != nil {
				t.Fatal(err)
			}
			client, err := bmwcardata.NewClient(
				bmwcardata.WithCarDataServer("http://rest.invalid"), // avoids requiring an authenticator
				bmwcardata.WithStreamingURL(streamURL),
				bmwcardata.WithStreamingProfile(tc.profile),
			)
			if err != nil {
				t.Fatalf("NewClient: %s", err)
			}
			if err := client.StartEventStream(); err != nil {
				t.Fatalf("StartEventStream: %s", err)
			}
			t.Cleanup(func() { _ = client.StopEventStream() })

			received := make(chan bmwcardata.StreamedMessage, 1)
			if _, err := client.Subscribe(context.Background(), bmwcardata.AllVINs, func(m bmwcardata.StreamedMessage) {
				select {
				case received <- m:
				default:
				}
			}); err != nil {
				t.Fatalf("Subscribe: %s", err)
			}

			payload, _ := json.Marshal(bmwcardata.StreamedMessage{VIN: "WBATEST", EntityID: "42"})

			deadline := time.After(10 * time.Second)
			ticker := time.NewTicker(150 * time.Millisecond)
			defer ticker.Stop()
			for {
				if err := broker.Publish("WBATEST", payload); err != nil {
					t.Fatalf("Publish: %s", err)
				}
				select {
				case msg := <-received:
					if msg.VIN != "WBATEST" || msg.EntityID != "42" {
						t.Fatalf("unexpected message: %+v", msg)
					}
					return
				case <-ticker.C:
					// Re-publish; the subscription may have raced the first publish.
				case <-deadline:
					t.Fatal("timed out waiting for bridged message")
				}
			}
		})
	}
}
