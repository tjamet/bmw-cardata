package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"time"

	bmwcardata "github.com/tjamet/bmw-cardata"
	"github.com/tjamet/bmw-cardata/cardataapi"
)

// Config configures a proxy Server.
type Config struct {
	// Authenticator authenticates to BMW. It is used both to authorise the REST
	// upstream requests and, unless Source is provided, to build the streaming
	// client. Required unless Source and an explicit upstream editor are supplied.
	Authenticator bmwcardata.AuthenticatorInterface

	// Source is the bmwcardata client holding the single BMW streaming
	// connection. When nil, one is built from Authenticator.
	Source *bmwcardata.Client

	// UpstreamREST is the BMW REST base URL. Defaults to the official server.
	UpstreamREST string

	// HTTPAddress is the listen address for the REST reverse proxy, e.g. ":8080".
	HTTPAddress string
	// HTTPTLS, when set, serves the REST proxy over HTTPS.
	HTTPTLS *tls.Config

	// MQTTAddress is the listen address for the MQTT broker, e.g. ":1883".
	MQTTAddress string
	// MQTTTLS, when set, serves the MQTT broker over TLS.
	MQTTTLS *tls.Config
	// MQTTUsers enables login/password auth on the broker. Empty means allow-all.
	MQTTUsers map[string]string

	// RESTAuth authorises incoming REST requests. Defaults to AllowAll.
	RESTAuth Authorizer
	// CacheTTL is the freshness window for cached REST responses. Zero disables
	// expiry.
	CacheTTL time.Duration

	// VINs to subscribe to on the BMW stream. Empty subscribes to all VINs.
	VINs []string
	// TopicPrefix is prepended to the VIN when republishing to the broker. It
	// must match downstream StaticStreamingProfile clients.
	TopicPrefix string
}

// Server bundles the caching REST reverse proxy, the self-hosted MQTT broker and
// the BMW-to-broker bridge into a single runnable unit.
type Server struct {
	REST   *RESTProxy
	Broker *Broker
	Bridge *Bridge

	http *http.Server
	cfg  Config
}

// NewServer builds a proxy Server from cfg.
func NewServer(cfg Config) (*Server, error) {
	if cfg.Authenticator == nil && cfg.Source == nil {
		return nil, errors.New("proxy: Authenticator or Source is required")
	}

	upstream := cfg.UpstreamREST
	if upstream == "" {
		upstream = cardataapi.CarDataAPIServer
	}

	source := cfg.Source
	if source == nil {
		var err error
		source, err = bmwcardata.NewClient(bmwcardata.WithAuthenticator(cfg.Authenticator))
		if err != nil {
			return nil, err
		}
	}

	authorizer := cfg.RESTAuth
	if authorizer == nil {
		authorizer = AllowAll()
	}

	restOpts := []RESTProxyOption{
		WithAuthorizer(authorizer),
		WithCache(NewMemoryCache(cfg.CacheTTL)),
	}
	if cfg.Authenticator != nil {
		restOpts = append(restOpts, WithUpstreamEditor(BearerAuthEditor(cfg.Authenticator)))
	}
	rest, err := NewRESTProxy(upstream, restOpts...)
	if err != nil {
		return nil, err
	}

	broker, err := NewBroker(BrokerOptions{
		Address:   cfg.MQTTAddress,
		TLSConfig: cfg.MQTTTLS,
		Users:     cfg.MQTTUsers,
	})
	if err != nil {
		return nil, err
	}

	bridge := &Bridge{
		Source:      source,
		Publisher:   broker,
		TopicPrefix: cfg.TopicPrefix,
		OnMessage:   rest.ApplyMQTT,
	}

	return &Server{
		REST:   rest,
		Broker: broker,
		Bridge: bridge,
		http:   &http.Server{Addr: cfg.HTTPAddress, Handler: rest, TLSConfig: cfg.HTTPTLS},
		cfg:    cfg,
	}, nil
}

// Run starts the broker, the BMW bridge and the REST reverse proxy, blocking
// until ctx is cancelled, at which point everything is shut down.
func (s *Server) Run(ctx context.Context) error {
	brokerErr := make(chan error, 1)
	go func() { brokerErr <- s.Broker.Serve() }()

	if err := s.Bridge.Start(ctx, s.cfg.VINs...); err != nil {
		_ = s.Broker.Close()
		return err
	}

	httpErr := make(chan error, 1)
	go func() {
		if s.http.TLSConfig != nil {
			httpErr <- s.http.ListenAndServeTLS("", "")
		} else {
			httpErr <- s.http.ListenAndServe()
		}
	}()

	var runErr error
	select {
	case <-ctx.Done():
	case err := <-brokerErr:
		runErr = err
	case err := <-httpErr:
		if !errors.Is(err, http.ErrServerClosed) {
			runErr = err
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = s.http.Shutdown(shutdownCtx)
	_ = s.Bridge.Stop()
	_ = s.Broker.Close()
	return runErr
}
