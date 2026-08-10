package bmwcardata

import (
	"context"
	"time"

	"github.com/eclipse/paho.golang/paho"
)

// StreamingProfile controls how the client authenticates to, and subscribes on,
// the MQTT streaming endpoint.
//
// The default profile ([bmwStreamingProfile]) targets BMW's official broker: it
// authenticates with the GCID/id-token pair from the session and subscribes to
// "<gcid>/<vin>" topics. A custom profile (see [StaticStreamingProfile]) allows
// the same client to target a self-hosted broker, for example the one exposed by
// the proxy package, which fans out the single BMW connection to many local
// subscribers.
type StreamingProfile interface {
	// ConfigureConnect mutates the MQTT CONNECT packet to carry the credentials
	// (and any connection properties) expected by the broker.
	ConfigureConnect(ctx context.Context, connect *paho.Connect) error
	// Topic returns the subscription topic filter for the given VIN.
	Topic(ctx context.Context, vin string) (string, error)
}

// bmwStreamingProfile is the default StreamingProfile. It authenticates against
// BMW's broker using the session's GCID as username and the id-token as
// password, and subscribes to "<gcid>/<vin>" topics.
type bmwStreamingProfile struct {
	authenticator AuthenticatorInterface
}

func (b bmwStreamingProfile) ConfigureConnect(ctx context.Context, connect *paho.Connect) error {
	session, err := b.authenticator.GetSession(ctx)
	if err != nil {
		return err
	}
	connect.UsernameFlag = true
	connect.PasswordFlag = true
	connect.Username = session.Gcid
	if session.IdToken != nil {
		connect.Password = []byte(*session.IdToken)
	}
	connect.Properties = &paho.ConnectProperties{
		SessionExpiryInterval: p(uint32(time.Until(session.ExpiresAt).Seconds())),
	}
	return nil
}

func (b bmwStreamingProfile) Topic(ctx context.Context, vin string) (string, error) {
	session, err := b.authenticator.GetSession(ctx)
	if err != nil {
		return "", err
	}
	return session.Gcid + "/" + vin, nil
}

// StaticStreamingProfile is a StreamingProfile for self-hosted brokers that do
// not implement BMW's authentication scheme. It sends fixed (optionally empty)
// credentials and builds subscription topics as "<TopicPrefix><vin>".
//
// It pairs with the broker exposed by the proxy package: leave Username and
// Password empty for an allow-all broker, or set them to match the broker's
// login/password. TopicPrefix must match the prefix the proxy publishes under
// (empty by default, so the topic is exactly the VIN).
type StaticStreamingProfile struct {
	// Username and Password are the MQTT credentials. When both are empty, the
	// CONNECT packet carries no credentials, which suits an allow-all broker.
	Username string
	Password string
	// TopicPrefix is prepended to the VIN to build the subscription topic, e.g.
	// "cardata/" yields "cardata/<vin>". Empty means the topic is exactly the VIN.
	TopicPrefix string
	// SessionExpiryInterval is advertised to the broker in seconds. When zero, a
	// default of 60 seconds is used.
	SessionExpiryInterval uint32
}

var _ StreamingProfile = StaticStreamingProfile{}

func (s StaticStreamingProfile) ConfigureConnect(ctx context.Context, connect *paho.Connect) error {
	if s.Username != "" {
		connect.UsernameFlag = true
		connect.Username = s.Username
	}
	if s.Password != "" {
		connect.PasswordFlag = true
		connect.Password = []byte(s.Password)
	}
	expiry := s.SessionExpiryInterval
	if expiry == 0 {
		expiry = 60
	}
	connect.Properties = &paho.ConnectProperties{
		SessionExpiryInterval: p(expiry),
	}
	return nil
}

func (s StaticStreamingProfile) Topic(ctx context.Context, vin string) (string, error) {
	return s.TopicPrefix + vin, nil
}
