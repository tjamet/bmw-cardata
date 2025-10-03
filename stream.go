package bmwcardata

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
)

func p[T any](v T) *T {
	return &v
}

type ReasonCode int

func (c ReasonCode) String() string {
	return fmt.Sprintf("0x%02X", int(c))
}

func (c ReasonCode) MQTTError() *MQTTReasonCode {
	code, ok := mqttReasonCodes[c]
	if !ok {
		return nil
	}
	return &code
}

func (c ReasonCode) Name() string {
	mqttError := c.MQTTError()
	if mqttError == nil {
		return "Unknown"
	}
	return c.MQTTError().Name
}

func (c ReasonCode) Description() string {
	mqttError := c.MQTTError()
	if mqttError == nil {
		return "Unknown"
	}
	return mqttError.Details
}

func (c MQTTError) Error() string {
	return fmt.Sprintf("%d (%s): %s", int(c), ReasonCode(c).Name(), ReasonCode(c).Description())
}

type MQTTReasonCode struct {
	ReasonCode ReasonCode
	Name       string
	Packets    []string
	Details    string
}

type MQTTError ReasonCode

const (
	StreamingEndpoint = "mqtts://customer.streaming-cardata.bmwgroup.com:9000"
)

var (
	streamingURL = Must(url.Parse(StreamingEndpoint))
)

type StreamedMessage struct {
	VIN       string                         `json:"vin"`
	EntityID  string                         `json:"entityId"`
	Topic     string                         `json:"topic"`
	Timestamp string                         `json:"timestamp"`
	Data      map[string]StreamedDataDetails `json:"data"`
}

type StreamedDataValue struct {
	String *string
	Bool   *bool
	Int    *int64
	Float  *float64
}

func (v *StreamedDataValue) UnmarshalJSON(data []byte) error {
	var raw any
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	switch val := raw.(type) {
	case string:
		v.String = &val

	case bool:
		v.Bool = &val
	case int64:
		v.Int = &val
	case float64:
		v.Float = &val
	default:
		return fmt.Errorf("unsupported type: %T", val)
	}
	return nil
}

func (v StreamedDataValue) MarshalJSON() ([]byte, error) {
	if v.String != nil {
		return json.Marshal(v.String)
	}
	if v.Bool != nil {
		return json.Marshal(v.Bool)
	}
	if v.Int != nil {
		return json.Marshal(v.Int)
	}
	return json.Marshal(v.Float)
}

type StreamedDataDetails struct {
	Timestamp string            `json:"timestamp,omitempty"`
	Value     StreamedDataValue `json:"value"`
	Unit      string            `json:"unit,omitempty"`
}

// Subscribe allows to subscribe to vehicle changes rather than polling the BMW CarData API.
// At the time of writing, the BMW API rate limit is set to 50 requests per day.
//
// See: https://bmw-cardata.bmwgroup.com/customer/public/api-documentation/Id-CarData-API_Authentication
//
// EXP(tjamet): This function is still experimental and its interface may change in the future.
func (c *Client) Subscribe(ctx context.Context, vin string, callback func(message StreamedMessage)) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()

	cliCfg := autopaho.ClientConfig{
		ServerUrls: []*url.URL{streamingURL},
		TlsCfg: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		KeepAlive: 20,
		ReconnectBackoff: func(attempt int) time.Duration {
			return time.Duration(attempt) * 10 * time.Second
		},
		// ConnectUsername:               session.Gcid,
		// ConnectPassword:               []byte(*session.IdToken),
		CleanStartOnInitialConnection: false,
		SessionExpiryInterval:         60,
		OnConnectionDown: func() bool {
			return true
		},
		OnConnectError: func(err error) {
			if connackErr, ok := err.(*autopaho.ConnackError); ok {
				err = MQTTError(connackErr.ReasonCode)
			}
			fmt.Printf("error whilst attempting connection: %s\n", err)
		},
		ConnectPacketBuilder: func(connect *paho.Connect, url *url.URL) (*paho.Connect, error) {
			fmt.Println("building connect packet")
			session, err := c.Authenticator.GetSession(ctx)
			if err != nil {
				return nil, err
			}
			connect.UsernameFlag = true
			connect.PasswordFlag = true
			connect.Username = session.Gcid
			connect.Password = []byte(*session.IdToken)
			connect.Properties = &paho.ConnectProperties{
				SessionExpiryInterval: p(uint32(time.Until(session.ExpiresAt).Seconds())),
			}
			return connect, nil
		},
		ClientConfig: paho.ClientConfig{
			ClientID: ClientID,
			OnClientError: func(err error) {
				fmt.Printf("client error: %s\n", err)
			},
		},
	}
	cliCfg.OnServerDisconnect = func(d *paho.Disconnect) {

		if d.Properties != nil {
			fmt.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
		} else {
			fmt.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
		}
	}

	// OnPublishReceived is a slice of functions that will be called when a message is received.
	// You can write the function(s) yourself or use the supplied Router
	cliCfg.ClientConfig.OnPublishReceived = []func(paho.PublishReceived) (bool, error){
		func(pr paho.PublishReceived) (bool, error) {
			var msg StreamedMessage
			err := json.Unmarshal(pr.Packet.Payload, &msg)
			if err != nil {
				return true, fmt.Errorf("error unmarshaling message: %s\n", err)
			}
			callback(msg)
			return true, nil
		},
	}
	cliCfg.OnConnectionUp = func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
		session, err := c.Authenticator.GetSession(ctx)
		if err != nil {
			fmt.Printf("error getting session: %s\n", err)
			return
		}

		topics := []string{
			fmt.Sprintf("%s/%s", session.Gcid, vin),
			fmt.Sprintf("%s/%s/#", session.Gcid, vin),
			fmt.Sprintf("%s/%s/+", session.Gcid, vin),
		}

		subscriptions := []paho.SubscribeOptions{}
		for _, topic := range topics {
			subscriptions = append(subscriptions, paho.SubscribeOptions{Topic: topic, QoS: 1})
		}

		// Subscribing in the OnConnectionUp callback is recommended (ensures the subscription is reestablished if
		// the connection drops)
		if _, err := cm.Subscribe(context.Background(), &paho.Subscribe{
			Subscriptions: subscriptions,
		}); err != nil {
			fmt.Printf("failed to subscribe (%s). This is likely to mean no messages will be received.", err)
		}
		fmt.Println("mqtt subscription made to", strings.Join(topics, ", "))
	}

	connection, err := autopaho.NewConnection(ctx, cliCfg) // starts process; will reconnect until context cancelled
	if err != nil {
		panic(err)
	}

	// Wait for the connection to come up
	if err = connection.AwaitConnection(ctx); err != nil {
		panic(err)
	}

	<-ctx.Done() // Wait for clean shutdown (cancelling the context triggered the shutdown)
	fmt.Println("signal caught - exiting")
	return nil
}
