package bmwcardata

import (
	"context"
	"testing"
	"time"

	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
)

type staticAuthenticator struct {
	session *AuthenticatedSession
	err     error
}

func (s staticAuthenticator) GetSession(ctx context.Context) (*AuthenticatedSession, error) {
	return s.session, s.err
}

func TestStaticStreamingProfile(t *testing.T) {
	ctx := context.Background()

	t.Run("no credentials", func(t *testing.T) {
		p := StaticStreamingProfile{TopicPrefix: "cardata/"}
		connect := &paho.Connect{}
		if err := p.ConfigureConnect(ctx, connect); err != nil {
			t.Fatalf("ConfigureConnect: %s", err)
		}
		if connect.UsernameFlag || connect.PasswordFlag {
			t.Fatal("expected no credentials for an allow-all broker")
		}
		if connect.Properties == nil || connect.Properties.SessionExpiryInterval == nil {
			t.Fatal("expected a session expiry interval")
		}
		if *connect.Properties.SessionExpiryInterval != 60 {
			t.Fatalf("default session expiry = %d, want 60", *connect.Properties.SessionExpiryInterval)
		}
		topic, err := p.Topic(ctx, "WBA1")
		if err != nil {
			t.Fatalf("Topic: %s", err)
		}
		if topic != "cardata/WBA1" {
			t.Fatalf("topic = %q, want cardata/WBA1", topic)
		}
	})

	t.Run("with credentials", func(t *testing.T) {
		p := StaticStreamingProfile{Username: "u", Password: "pw", SessionExpiryInterval: 120}
		connect := &paho.Connect{}
		if err := p.ConfigureConnect(ctx, connect); err != nil {
			t.Fatalf("ConfigureConnect: %s", err)
		}
		if !connect.UsernameFlag || connect.Username != "u" {
			t.Fatalf("username = %q (flag %v), want u", connect.Username, connect.UsernameFlag)
		}
		if !connect.PasswordFlag || string(connect.Password) != "pw" {
			t.Fatalf("password = %q (flag %v), want pw", connect.Password, connect.PasswordFlag)
		}
		if *connect.Properties.SessionExpiryInterval != 120 {
			t.Fatalf("session expiry = %d, want 120", *connect.Properties.SessionExpiryInterval)
		}
	})
}

func TestBMWStreamingProfile(t *testing.T) {
	ctx := context.Background()
	idToken := "id-token"
	profile := bmwStreamingProfile{authenticator: staticAuthenticator{session: &AuthenticatedSession{
		Gcid:      "gcid-1",
		IdToken:   &idToken,
		ExpiresAt: time.Now().Add(time.Hour),
	}}}

	connect := &paho.Connect{}
	if err := profile.ConfigureConnect(ctx, connect); err != nil {
		t.Fatalf("ConfigureConnect: %s", err)
	}
	if connect.Username != "gcid-1" || string(connect.Password) != "id-token" {
		t.Fatalf("BMW connect = (%q,%q), want (gcid-1,id-token)", connect.Username, connect.Password)
	}

	topic, err := profile.Topic(ctx, "WBA1")
	if err != nil {
		t.Fatalf("Topic: %s", err)
	}
	if topic != "gcid-1/WBA1" {
		t.Fatalf("topic = %q, want gcid-1/WBA1", topic)
	}
}

func TestResolveStreamingProfileDefault(t *testing.T) {
	auth := staticAuthenticator{session: &AuthenticatedSession{ClientID: uuid.New()}}
	client, err := NewClient(
		WithCarDataServer("http://rest.invalid"),
		WithAuthenticator(auth),
	)
	if err != nil {
		t.Fatalf("NewClient: %s", err)
	}
	if _, ok := client.resolveStreamingProfile().(bmwStreamingProfile); !ok {
		t.Fatal("expected default profile to be the BMW profile")
	}

	custom := StaticStreamingProfile{TopicPrefix: "x/"}
	client.StreamingProfile = custom
	if got, ok := client.resolveStreamingProfile().(StaticStreamingProfile); !ok || got.TopicPrefix != "x/" {
		t.Fatal("expected the configured static profile to be returned")
	}
}
