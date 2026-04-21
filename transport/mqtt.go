package transport

import (
	"fmt"
	"log"
	"sync"
	"time"

	pahomqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTTransport implements Transport using the paho MQTT client library.
//
// It handles automatic reconnection and resubscription of all topics
// when the connection is restored. For MQTT 5.0 brokers (e.g. EMQX),
// it extracts publisher metadata from message properties.
type MQTTTransport struct {
	brokers              []string
	clientID             string
	autoReconnect        bool
	connectRetry         bool
	keepAlive            time.Duration
	pingTimeout          time.Duration
	maxReconnectInterval time.Duration
	cleanSession         bool
	defaultQoS           byte
	onConnect            func()
	onConnectionLost     func(err error)

	client        pahomqtt.Client
	mu            sync.RWMutex
	subs          map[string]MessageHandler
	globalHandler func(topic string, payload []byte, props MessageProperties)
}

// NewMQTTTransport creates a new MQTT transport with the given options.
func NewMQTTTransport(opts ...MQTTTransportOption) *MQTTTransport {
	t := &MQTTTransport{
		autoReconnect:        true,
		connectRetry:         true,
		keepAlive:            60 * time.Second,
		pingTimeout:          30 * time.Second,
		maxReconnectInterval: 1 * time.Minute,
		cleanSession:         true,
		defaultQoS:           0,
		subs:                 make(map[string]MessageHandler),
	}
	for _, opt := range opts {
		opt(t)
	}
	return t
}

// Connect establishes a connection to the MQTT broker.
func (t *MQTTTransport) Connect() error {
	opts := pahomqtt.NewClientOptions()

	for _, broker := range t.brokers {
		opts.AddBroker(broker)
	}

	opts.SetClientID(resolvedClientID(t.clientID))
	opts.SetAutoReconnect(t.autoReconnect)
	opts.SetConnectRetry(t.connectRetry)
	opts.SetKeepAlive(t.keepAlive)
	opts.SetPingTimeout(t.pingTimeout)
	opts.SetMaxReconnectInterval(t.maxReconnectInterval)
	opts.SetCleanSession(t.cleanSession)

	opts.SetConnectionLostHandler(func(_ pahomqtt.Client, err error) {
		log.Printf("[courier/transport] connection lost: %v", err)
		if t.onConnectionLost != nil {
			t.onConnectionLost(err)
		}
	})

	opts.SetOnConnectHandler(func(_ pahomqtt.Client) {
		log.Println("[courier/transport] connected, resubscribing...")
		t.resubscribeAll()

		if t.onConnect != nil {
			t.onConnect()
		}
	})

	opts.SetDefaultPublishHandler(func(_ pahomqtt.Client, msg pahomqtt.Message) {
		t.mu.RLock()
		handler := t.globalHandler
		t.mu.RUnlock()

		if handler != nil {
			props := extractProperties(msg)
			handler(msg.Topic(), msg.Payload(), props)
		}
	})

	t.client = pahomqtt.NewClient(opts)
	if token := t.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("courier/transport: connect failed: %w", token.Error())
	}

	return nil
}

// Close disconnects from the MQTT broker.
func (t *MQTTTransport) Close() error {
	if t.client != nil && t.client.IsConnected() {
		t.client.Disconnect(250)
	}
	return nil
}

// Subscribe registers a handler for the given topic.
func (t *MQTTTransport) Subscribe(topic string, handler MessageHandler) error {
	t.mu.Lock()
	t.subs[topic] = handler
	t.rebuildGlobalHandler()
	t.mu.Unlock()

	if t.client != nil && t.client.IsConnected() {
		if token := t.client.Subscribe(topic, t.defaultQoS, nil); token.Wait() && token.Error() != nil {
			return fmt.Errorf("courier/transport: subscribe to %s failed: %w", topic, token.Error())
		}
	}
	return nil
}

// Unsubscribe removes the subscription for the given topic.
func (t *MQTTTransport) Unsubscribe(topic string) error {
	t.mu.Lock()
	delete(t.subs, topic)
	t.rebuildGlobalHandler()
	t.mu.Unlock()

	if t.client != nil && t.client.IsConnected() {
		if token := t.client.Unsubscribe(topic); token.Wait() && token.Error() != nil {
			return fmt.Errorf("courier/transport: unsubscribe from %s failed: %w", topic, token.Error())
		}
	}
	return nil
}

// Publish sends a message to the given topic.
func (t *MQTTTransport) Publish(topic string, payload []byte) error {
	if t.client == nil || !t.client.IsConnected() {
		return fmt.Errorf("courier/transport: not connected")
	}
	if token := t.client.Publish(topic, t.defaultQoS, false, payload); token.Wait() && token.Error() != nil {
		return fmt.Errorf("courier/transport: publish to %s failed: %w", topic, token.Error())
	}
	return nil
}

func (t *MQTTTransport) rebuildGlobalHandler() {
	snapshot := make(map[string]MessageHandler, len(t.subs))
	for k, v := range t.subs {
		snapshot[k] = v
	}
	t.globalHandler = func(topic string, payload []byte, props MessageProperties) {
		if h, ok := snapshot[topic]; ok {
			h(topic, payload, props)
		}
	}
}

func (t *MQTTTransport) resubscribeAll() {
	t.mu.RLock()
	topics := make([]string, 0, len(t.subs))
	for topic := range t.subs {
		topics = append(topics, topic)
	}
	t.mu.RUnlock()

	for _, topic := range topics {
		if token := t.client.Subscribe(topic, t.defaultQoS, nil); token.Wait() && token.Error() != nil {
			log.Printf("[courier/transport] resubscribe to %s failed: %v", topic, token.Error())
		} else {
			log.Printf("[courier/transport] resubscribed to %s", topic)
		}
	}
}

// extractProperties pulls publisher and user metadata from MQTT message properties.
// The current paho v1 library does not expose MQTT 5.0 properties on the Message interface.
// Properties will be populated when upgrading to paho v2 or using EMQX's rule engine
// to inject publisher metadata into the payload.
//
// EMQX configuration options to pass client_id:
//   - Rule engine: rewrite payload to include client_id
//   - Webhook plugin: inject client_id as a prefix in the payload
//   - Upgrade to paho.mqtt.golang v2 which supports MQTT 5.0 properties
func extractProperties(msg pahomqtt.Message) MessageProperties {
	return make(MessageProperties)
}
