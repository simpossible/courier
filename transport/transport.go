// Package transport abstracts the underlying messaging transport used by courier.
//
// The Transport interface decouples the RPC layer from MQTT specifics,
// allowing alternative transports (WebSocket, QUIC, etc.) in the future.
package transport

// MessageHandler is called when a message is received on a subscribed topic.
type MessageHandler func(topic string, payload []byte)

// Transport represents a message transport capable of pub/sub messaging.
type Transport interface {
	// Connect establishes a connection to the message broker.
	Connect() error

	// Close disconnects from the broker and releases resources.
	Close() error

	// Subscribe registers a handler for the given topic.
	// If the transport is connected, the subscription takes effect immediately.
	// If not yet connected, subscriptions are recorded and applied on connect.
	Subscribe(topic string, handler MessageHandler) error

	// Unsubscribe removes the subscription for the given topic.
	Unsubscribe(topic string) error

	// Publish sends a message to the given topic.
	Publish(topic string, payload []byte) error
}
