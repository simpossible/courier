// Package transport abstracts the underlying messaging transport used by courier.
//
// The Transport interface decouples the RPC layer from MQTT specifics,
// allowing alternative transports (WebSocket, QUIC, etc.) in the future.
package transport

// MessageProperties carries metadata extracted from the transport layer.
// For MQTT 5.0 with EMQX, this includes the publisher's ClientID and user properties.
type MessageProperties map[string]string

// MessageHandler is called when a message is received on a subscribed topic.
type MessageHandler func(topic string, payload []byte, props MessageProperties)

// Transport represents a message transport capable of pub/sub messaging.
type Transport interface {
	// Connect establishes a connection to the message broker.
	Connect() error

	// Close disconnects from the broker and releases resources.
	Close() error

	// Subscribe registers a handler for the given topic.
	Subscribe(topic string, handler MessageHandler) error

	// Unsubscribe removes the subscription for the given topic.
	Unsubscribe(topic string) error

	// Publish sends a message to the given topic.
	Publish(topic string, payload []byte) error
}
