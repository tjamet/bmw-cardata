// Command bmw-cardata-proxy runs a self-hosted front-end for the BMW CarData
// API. It keeps the single BMW streaming connection and the REST access, then
// exposes:
//
//   - a self-hosted MQTT broker that fans the BMW event stream out to any number
//     of local subscribers, and
//   - a caching REST reverse proxy that reduces pressure on the BMW rate limits.
//
// Point the bmwcardata client at this proxy with WithCarDataServer (REST) and a
// StaticStreamingProfile (MQTT).
package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	bmwcardata "github.com/tjamet/bmw-cardata"
	"github.com/tjamet/bmw-cardata/proxy"
)

func main() {
	defaultSessionPath, err := bmwcardata.DefaultSessionPath()
	if err != nil {
		log.Fatal(err)
	}

	sessionPath := flag.String("session-path", defaultSessionPath, "Path to the session file")
	clientID := flag.String("client-id", "", "BMW client ID (UUID)")

	httpAddr := flag.String("http-address", ":8080", "Listen address for the caching REST proxy")
	mqttAddr := flag.String("mqtt-address", ":1883", "Listen address for the MQTT broker")

	cacheTTL := flag.Duration("cache-ttl", 5*time.Minute, "Freshness window for cached REST responses (0 disables expiry)")

	vins := flag.String("vins", "", "Comma-separated VINs to stream (empty streams all VINs)")
	topicPrefix := flag.String("topic-prefix", "", "Prefix prepended to the VIN when republishing to the broker")

	mqttUser := flag.String("mqtt-user", "", "Username required to connect to the broker (empty allows all)")
	mqttPassword := flag.String("mqtt-password", "", "Password required to connect to the broker")

	restUser := flag.String("rest-user", "", "Username required for the REST proxy via HTTP Basic (empty allows all)")
	restPassword := flag.String("rest-password", "", "Password required for the REST proxy via HTTP Basic")

	tlsCert := flag.String("tls-cert", "", "Certificate file enabling TLS on both the REST proxy and the MQTT broker")
	tlsKey := flag.String("tls-key", "", "Private key file matching -tls-cert")

	flag.Parse()

	if *clientID == "" {
		log.Fatal("-client-id is required")
	}

	authenticator := bmwcardata.Must(bmwcardata.NewAuthenticator(
		bmwcardata.WithSessionStore(&bmwcardata.FileSessionStore{Path: *sessionPath}),
		bmwcardata.WithClientID(*clientID),
		bmwcardata.WithPromptURI(func(uri, code string) {
			log.Println("Open the following URL in your browser:")
			log.Println(uri)
			log.Println("And enter the code:")
			log.Println(code)
		}),
	))

	var tlsConfig *tls.Config
	if *tlsCert != "" || *tlsKey != "" {
		cert, err := tls.LoadX509KeyPair(*tlsCert, *tlsKey)
		if err != nil {
			log.Fatalf("failed to load TLS key pair: %s", err)
		}
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}
	}

	cfg := proxy.Config{
		Authenticator: authenticator,
		HTTPAddress:   *httpAddr,
		HTTPTLS:       tlsConfig,
		MQTTAddress:   *mqttAddr,
		MQTTTLS:       tlsConfig,
		CacheTTL:      *cacheTTL,
		TopicPrefix:   *topicPrefix,
		VINs:          splitVINs(*vins),
	}
	if *mqttUser != "" {
		cfg.MQTTUsers = map[string]string{*mqttUser: *mqttPassword}
	}
	if *restUser != "" {
		cfg.RESTAuth = proxy.BasicAuth(map[string]string{*restUser: *restPassword})
	}

	server, err := proxy.NewServer(cfg)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	log.Printf("REST proxy listening on %s, MQTT broker listening on %s", *httpAddr, *mqttAddr)
	if err := server.Run(ctx); err != nil {
		log.Fatal(err)
	}
}

func splitVINs(s string) []string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	vins := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			vins = append(vins, v)
		}
	}
	return vins
}
