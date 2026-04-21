package transport

import (
	"fmt"
	"time"
)

// MQTTTransportOption configures an MQTTTransport.
type MQTTTransportOption func(*MQTTTransport)

// WithBrokers sets the MQTT broker addresses (e.g. "tcp://localhost:1883").
// At least one broker is required.
func WithBrokers(brokers ...string) MQTTTransportOption {
	return func(t *MQTTTransport) {
		t.brokers = brokers
	}
}

// WithClientID sets the MQTT client identifier.
// If empty, one is generated automatically.
func WithClientID(id string) MQTTTransportOption {
	return func(t *MQTTTransport) {
		t.clientID = id
	}
}

// WithAutoReconnect enables or disables automatic reconnection. Default: true.
func WithAutoReconnect(enabled bool) MQTTTransportOption {
	return func(t *MQTTTransport) {
		t.autoReconnect = enabled
	}
}

// WithConnectRetry enables or disables connection retry on initial connect. Default: true.
func WithConnectRetry(enabled bool) MQTTTransportOption {
	return func(t *MQTTTransport) {
		t.connectRetry = enabled
	}
}

// WithKeepAlive sets the MQTT keep-alive interval. Default: 60s.
func WithKeepAlive(d time.Duration) MQTTTransportOption {
	return func(t *MQTTTransport) {
		t.keepAlive = d
	}
}

// WithPingTimeout sets the timeout for MQTT ping responses. Default: 30s.
func WithPingTimeout(d time.Duration) MQTTTransportOption {
	return func(t *MQTTTransport) {
		t.pingTimeout = d
	}
}

// WithMaxReconnectInterval sets the maximum interval between reconnection attempts. Default: 1min.
func WithMaxReconnectInterval(d time.Duration) MQTTTransportOption {
	return func(t *MQTTTransport) {
		t.maxReconnectInterval = d
	}
}

// WithCleanSession sets the MQTT clean session flag. Default: true.
func WithCleanSession(clean bool) MQTTTransportOption {
	return func(t *MQTTTransport) {
		t.cleanSession = clean
	}
}

// WithOnConnect registers a callback invoked after a successful connection
// (including reconnections), in addition to the internal resubscribe logic.
func WithOnConnect(fn func()) MQTTTransportOption {
	return func(t *MQTTTransport) {
		t.onConnect = fn
	}
}

// WithOnConnectionLost registers a callback invoked when the connection is lost.
func WithOnConnectionLost(fn func(err error)) MQTTTransportOption {
	return func(t *MQTTTransport) {
		t.onConnectionLost = fn
	}
}

// WithDefaultQoS sets the QoS level for publish and subscribe operations. Default: 0.
func WithDefaultQoS(qos byte) MQTTTransportOption {
	return func(t *MQTTTransport) {
		t.defaultQoS = qos
	}
}

func resolvedClientID(id string) string {
	if id != "" {
		return id
	}
	return fmt.Sprintf("courier_%d", time.Now().UnixNano())
}
