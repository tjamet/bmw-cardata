package proxy

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"log/slog"

	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	bmwcardata "github.com/tjamet/bmw-cardata"
)

// Publisher publishes bridged MQTT payloads to the local broker.
type Publisher interface {
	// Publish delivers payload to subscribers of topic.
	Publish(topic string, payload []byte) error
}

// BrokerOptions configures the self-hosted MQTT broker.
type BrokerOptions struct {
	// Address is the TCP listen address, e.g. ":1883".
	Address string
	// TLSConfig, when set, wraps the listener in TLS ("client TLS"). Leave nil to
	// serve plain TCP.
	TLSConfig *tls.Config
	// Users, when non-empty, enables login/password authentication with the given
	// username/password pairs. When empty, the broker allows all connections.
	Users map[string]string
	// Logger overrides the broker logger. When nil, logging is discarded.
	Logger *slog.Logger
}

// Broker is a self-hosted MQTT broker that fronts the BMW event stream. It
// satisfies Publisher so the Bridge can fan BMW messages out to local
// subscribers.
type Broker struct {
	server *mqtt.Server
}

var _ Publisher = (*Broker)(nil)

// NewBroker builds an MQTT broker according to opts. Call Serve to start
// accepting connections and Close to stop.
func NewBroker(opts BrokerOptions) (*Broker, error) {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewTextHandler(io.Discard, nil))
	}
	server := mqtt.New(&mqtt.Options{
		InlineClient: true, // required for server-side Publish
		Logger:       logger,
	})

	if len(opts.Users) == 0 {
		if err := server.AddHook(new(auth.AllowHook), nil); err != nil {
			return nil, err
		}
	} else {
		ledger := &auth.Ledger{
			Users: auth.Users{},
			// Allow authenticated users to read and write any topic.
			ACL: auth.ACLRules{{}},
		}
		for user, pass := range opts.Users {
			ledger.Users[user] = auth.UserRule{
				Username: auth.RString(user),
				Password: auth.RString(pass),
			}
		}
		if err := server.AddHook(new(auth.Hook), &auth.Options{Ledger: ledger}); err != nil {
			return nil, err
		}
	}

	tcp := listeners.NewTCP(listeners.Config{
		Type:      listeners.TypeTCP,
		ID:        "cardata-tcp",
		Address:   opts.Address,
		TLSConfig: opts.TLSConfig,
	})
	if err := server.AddListener(tcp); err != nil {
		return nil, err
	}

	return &Broker{server: server}, nil
}

// Serve starts the broker listeners. It blocks until the broker is closed.
func (b *Broker) Serve() error { return b.server.Serve() }

// Close stops the broker and closes all listeners.
func (b *Broker) Close() error { return b.server.Close() }

// Publish delivers payload to subscribers of topic with QoS 0 and no retention.
func (b *Broker) Publish(topic string, payload []byte) error {
	return b.server.Publish(topic, payload, false, 0)
}

// Bridge consumes the BMW event stream through a bmwcardata.Client and
// republishes every message to a local Publisher, keeping BMW's single-connection
// constraint satisfied while serving many local subscribers.
type Bridge struct {
	// Source is the client holding the single connection to BMW.
	Source *bmwcardata.Client
	// Publisher receives the republished messages (typically a Broker).
	Publisher Publisher
	// TopicPrefix is prepended to the VIN to build the republish topic. It must
	// match the TopicPrefix of downstream StaticStreamingProfile clients.
	TopicPrefix string
	// OnMessage, when set, is invoked for every received message before it is
	// republished. The proxy uses it to refresh the REST cache.
	OnMessage func(bmwcardata.StreamedMessage)
}

// Start opens the BMW event stream and subscribes to the given VINs, forwarding
// every message to the Publisher. Pass bmwcardata.AllVINs to receive all VINs.
func (br *Bridge) Start(ctx context.Context, vins ...string) error {
	if len(vins) == 0 {
		vins = []string{bmwcardata.AllVINs}
	}
	if err := br.Source.StartEventStream(); err != nil {
		return err
	}
	for _, vin := range vins {
		if _, err := br.Source.Subscribe(ctx, vin, br.handle); err != nil {
			return err
		}
	}
	return nil
}

// Stop closes the underlying BMW event stream.
func (br *Bridge) Stop() error { return br.Source.StopEventStream() }

func (br *Bridge) handle(msg bmwcardata.StreamedMessage) {
	if br.OnMessage != nil {
		br.OnMessage(msg)
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return
	}
	_ = br.Publisher.Publish(br.TopicPrefix+msg.VIN, payload)
}
