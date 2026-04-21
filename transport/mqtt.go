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
// when the connection is restored.
type MQTTTransport struct {
	// Configuration fields, set via functional options.
	brokers                []string
	clientID               string
	autoReconnect          bool
	connectRetry           bool
	keepAlive              time.Duration
	pingTimeout            time.Duration
	maxReconnectInterval   time.Duration
	cleanSession           bool
	defaultQoS             byte
	onConnect              func()
	onConnectionLost       func(err error)

	client       pahomqtt.Client
	mu           sync.RWMutex
	subs         map[string]MessageHandler // topic → handler
	globalHandler MessageHandler           // unified dispatch function
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
			handler(msg.Topic(), msg.Payload())
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

// rebuildGlobalHandler creates a single dispatch function that routes messages
// by topic to the correct per-subscription handler. Must be called with t.mu held.
func (t *MQTTTransport) rebuildGlobalHandler() {
	snapshot := make(map[string]MessageHandler, len(t.subs))
	for k, v := range t.subs {
		snapshot[k] = v
	}
	t.globalHandler = func(topic string, payload []byte) {
		if h, ok := snapshot[topic]; ok {
			h(topic, payload)
		}
	}
}

// resubscribeAll re-establishes all subscriptions after a reconnect.
// It copies the topic list under a read lock, then subscribes without holding the lock.
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
