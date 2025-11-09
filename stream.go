package bmwcardata

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
)

const (
	AllVINs   = "+"
	AllTopics = "#"
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

type streamingManager struct {
	Authenticator     AuthenticatorInterface
	connectionManager *autopaho.ConnectionManager
	subscriptions     map[string]map[string]func(message StreamedMessage)
	m                 sync.Mutex
	streamingURL      *url.URL
	stop              context.CancelFunc
	ctx               context.Context
}

type Subscription struct {
	ID  string
	VIN string
}

// Subscribe registers a callback for the provided VINs. The MQTT connection is shared across
// subscriptions and is managed by the client. The returned subscription ID can be used to
// unsubscribe later on.
func (c *Client) Subscribe(ctx context.Context, vin string, callback func(message StreamedMessage)) (*Subscription, error) {
	if callback == nil {
		return nil, fmt.Errorf("callback must not be nil")
	}
	subscription := Subscription{ID: uuid.New().String(), VIN: vin}
	c.registerCallback(&subscription, callback)

	err := c.streaming.Load().updateSubscriptions(ctx, c.subscriptions)
	if err != nil {
		return nil, err
	}
	return &subscription, nil
}

func (c *Client) Unsubscribe(ctx context.Context, subscription *Subscription) error {
	if subscription == nil {
		return fmt.Errorf("subscription must not be nil")
	}
	c.unregisterCallback(subscription)
	err := c.streaming.Load().updateSubscriptions(ctx, c.subscriptions)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) registerCallback(subscription *Subscription, callback func(message StreamedMessage)) {
	c.m.Lock()
	defer c.m.Unlock()
	if c.subscriptions == nil {
		c.subscriptions = make(map[string]map[string]func(message StreamedMessage))
	}
	if _, ok := c.subscriptions[subscription.VIN]; !ok {
		c.subscriptions[subscription.VIN] = make(map[string]func(message StreamedMessage))
	}
	c.subscriptions[subscription.VIN][subscription.ID] = callback
}

func (c *Client) unregisterCallback(subscription *Subscription) {
	c.m.Lock()
	defer c.m.Unlock()
	if _, ok := c.subscriptions[subscription.VIN]; !ok {
		return
	}
	delete(c.subscriptions[subscription.VIN], subscription.ID)
	if len(c.subscriptions[subscription.VIN]) == 0 {
		delete(c.subscriptions, subscription.VIN)
	}
}

func (c *Client) Done() <-chan struct{} {
	existing := c.streaming.Load()
	if existing == nil {
		return nil
	}
	return existing.ctx.Done()
}

func (c *Client) StartEventStream() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)

	candidate := &streamingManager{
		Authenticator: c.Authenticator,
		streamingURL:  c.StreamingURL,
		subscriptions: c.subscriptions,
		ctx:           ctx,
		stop:          stop,
	}

	if c.streaming.CompareAndSwap(nil, candidate) {
		// the new connection manager was successfully stored,
		// we can start it.
		// In case there is a concurrent call to `ensureStreamingManager`,
		// it may happen that the other call wins and our candidate is not the one
		// stored. In this case, we won't get here, but the other one will
		// start the connection.
		if err := candidate.connect(); err != nil {
			return err
		}
		return nil
	} else {
		candidate.stop()
	}
	return nil
}

func (c *Client) StopEventStream() error {
	// try to clean the streaming manager
	existing := c.streaming.Load()
	if existing == nil {
		return nil
	}
	if !c.streaming.CompareAndSwap(existing, nil) {
		// another call to `cleanStreamingManager` won the race and cleaned the manager
		return nil
	}

	existing.stop()
	// Wait for the context to be done, so we can be sure that the connection
	<-existing.ctx.Done()
	return nil
}

func (m *streamingManager) connect() error {

	cm, err := autopaho.NewConnection(m.ctx, m.autopahoConfig())
	if err != nil {
		return err
	}
	m.connectionManager = cm

	err = cm.AwaitConnection(m.ctx)
	if err != nil {
		return err
	}
	go func() {
		<-m.ctx.Done()
		cm.Disconnect(m.ctx)
	}()

	return nil
}

func (m *streamingManager) autopahoConfig() autopaho.ClientConfig {
	return autopaho.ClientConfig{
		ServerUrls: []*url.URL{m.streamingURL},
		TlsCfg: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
		KeepAlive:                     20,
		ReconnectBackoff:              m.handlePahoReconnectBackoff,
		CleanStartOnInitialConnection: false,
		SessionExpiryInterval:         60,
		OnConnectionDown:              m.handlePahoConnectionDown,
		OnConnectionUp:                m.handlePahoConnectionUp,
		OnConnectError:                m.handlePahoConnectError,
		ConnectPacketBuilder:          m.buildPahoConnectPacket,
		ClientConfig: paho.ClientConfig{
			ClientID:      ClientID,
			OnClientError: m.onPahoClientError,
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				m.handlePahoPublishReceived,
			},
			OnServerDisconnect: m.handlePahoServerDisconnect,
		},
	}
}

func (m *streamingManager) handlePahoPublishReceived(pr paho.PublishReceived) (bool, error) {
	var msg StreamedMessage
	if err := json.Unmarshal(pr.Packet.Payload, &msg); err != nil {
		return true, fmt.Errorf("error unmarshaling message: %w", err)
	}
	for _, callback := range m.getCallbacks(msg.VIN) {
		go callback(msg)
	}
	return true, nil
}

func (m *streamingManager) handlePahoServerDisconnect(d *paho.Disconnect) {
	if d.Properties != nil {
		fmt.Printf("server requested disconnect: %s\n", d.Properties.ReasonString)
	} else {
		fmt.Printf("server requested disconnect; reason code: %d\n", d.ReasonCode)
	}
}

func (m *streamingManager) handlePahoConnectError(err error) {
	if connackErr, ok := err.(*autopaho.ConnackError); ok {
		err = MQTTError(connackErr.ReasonCode)
	}
	fmt.Printf("error whilst attempting connection: %s\n", err)
}

func (m *streamingManager) handlePahoReconnectBackoff(attempt int) time.Duration {
	return time.Duration(attempt) * 10 * time.Second
}

func (m *streamingManager) onPahoClientError(err error) {
	fmt.Printf("client error: %s\n", err)
}

func (m *streamingManager) handlePahoConnectionDown() bool {
	return true
}

func (m *streamingManager) listSubscribedVINs() []string {
	m.m.Lock()
	defer m.m.Unlock()
	vins := []string{}
	for vin := range m.subscriptions {
		vins = append(vins, vin)
	}
	return vins
}

func (m *streamingManager) getCallbacks(vin string) []func(message StreamedMessage) {
	m.m.Lock()
	defer m.m.Unlock()
	callbacks := []func(message StreamedMessage){}
	for _, callback := range m.subscriptions[vin] {
		callbacks = append(callbacks, callback)
	}
	for _, callback := range m.subscriptions[AllVINs] {
		callbacks = append(callbacks, callback)
	}
	for _, callback := range m.subscriptions[AllTopics] {
		callbacks = append(callbacks, callback)
	}
	return callbacks
}

func (m *streamingManager) updateSubscriptions(ctx context.Context, newSubscriptions map[string]map[string]func(message StreamedMessage)) error {
	if m == nil {
		return nil
	}
	m.m.Lock()
	defer m.m.Unlock()
	if m.connectionManager != nil {
		unsubscribe := &paho.Unsubscribe{}
		session, err := m.Authenticator.GetSession(m.ctx)
		if err != nil {
			fmt.Printf("error getting session: %s\n", err)
			return err
		}
		for vin := range m.subscriptions {
			if _, ok := newSubscriptions[vin]; !ok {
				unsubscribe.Topics = append(unsubscribe.Topics, fmt.Sprintf("%s/%s", session.Gcid, vin))
			}
		}
		if unsubscribe.Topics != nil {
			if _, err := m.connectionManager.Unsubscribe(m.ctx, unsubscribe); err != nil {
				fmt.Printf("failed to unsubscribe from topics: %s\n", err)
				return err
			}
		}
	}
	m.subscriptions = newSubscriptions
	return nil
}

func (m *streamingManager) handlePahoConnectionUp(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
	session, err := m.Authenticator.GetSession(m.ctx)
	if err != nil {
		fmt.Printf("error getting session: %s\n", err)
		return
	}

	subscribe := &paho.Subscribe{}
	for _, vin := range m.listSubscribedVINs() {
		subscribe.Subscriptions = append(subscribe.Subscriptions, paho.SubscribeOptions{Topic: fmt.Sprintf("%s/%s", session.Gcid, vin), QoS: 1})
	}
	if subscribe.Subscriptions != nil {
		if _, err := cm.Subscribe(m.ctx, subscribe); err != nil {
			fmt.Printf("failed to subscribe to topics: %s\n", err)
		}
	}
}

func (m *streamingManager) buildPahoConnectPacket(connect *paho.Connect, url *url.URL) (*paho.Connect, error) {
	session, err := m.Authenticator.GetSession(m.ctx)
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
}
