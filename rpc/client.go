package rpc

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/simpossible/courier/codec"
	"github.com/simpossible/courier/transport"
)

// pendingCall tracks an in-flight RPC request awaiting a response.
type pendingCall struct {
	respChan chan []byte
	timer    *time.Timer
}

// Client sends RPC requests over MQTT and matches responses to pending calls.
type Client struct {
	tp           transport.Transport
	deviceID     string
	timeout      time.Duration
	retryCount   int
	retryInterval time.Duration
	retryBackoff float64
	interceptors []Interceptor

	mu      sync.RWMutex
	pending map[[16]byte]*pendingCall
}

// NewClient creates a new RPC client with the given options.
func NewClient(opts ...ClientOption) *Client {
	c := &Client{
		timeout:       10 * time.Second,
		retryCount:    0,
		retryInterval: 1 * time.Second,
		retryBackoff:  1.5,
		pending:       make(map[[16]byte]*pendingCall),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Connect establishes the MQTT connection and subscribes to the response topic.
func (c *Client) Connect() error {
	if c.tp == nil {
		return fmt.Errorf("courier/rpc: client has no transport")
	}
	if c.deviceID == "" {
		return fmt.Errorf("courier/rpc: client has no device ID")
	}

	respTopic := ResponseTopic(c.deviceID)
	if err := c.tp.Subscribe(respTopic, c.handleResponse); err != nil {
		return fmt.Errorf("courier/rpc: subscribe to response topic failed: %w", err)
	}

	log.Printf("[courier/rpc] client subscribed to %s", respTopic)
	return nil
}

// Close unsubscribes from the response topic and cleans up pending calls.
func (c *Client) Close() error {
	if c.tp == nil {
		return nil
	}

	respTopic := ResponseTopic(c.deviceID)
	_ = c.tp.Unsubscribe(respTopic)

	c.mu.Lock()
	for id, call := range c.pending {
		call.timer.Stop()
		select {
		case call.respChan <- nil:
		default:
		}
		delete(c.pending, id)
	}
	c.mu.Unlock()

	return c.tp.Close()
}

// Call sends an RPC request and waits for the response or timeout.
// serviceName identifies the target service, cmd is the numeric method ID,
// and payload is the serialized request body (typically protobuf).
func (c *Client) Call(ctx context.Context, serviceName string, cmd uint32, payload []byte) ([]byte, error) {
	requestID, err := newRequestID()
	if err != nil {
		return nil, fmt.Errorf("courier/rpc: generate request ID: %w", err)
	}

	call := &pendingCall{
		respChan: make(chan []byte, 1),
	}

	// Register pending call before publishing.
	c.mu.Lock()
	c.pending[requestID] = call
	c.mu.Unlock()

	// Ensure cleanup on any exit path.
	defer func() {
		c.mu.Lock()
		if pc, ok := c.pending[requestID]; ok {
			pc.timer.Stop()
			delete(c.pending, requestID)
		}
		c.mu.Unlock()
	}()

	// Set timeout timer.
	call.timer = time.AfterFunc(c.timeout, func() {
		c.mu.Lock()
		delete(c.pending, requestID)
		c.mu.Unlock()

		select {
		case call.respChan <- nil:
		default:
		}
	})

	// Encode and publish.
	reqBytes := codec.EncodeRequest(cmd, payload)
	reqTopic := RequestTopic(serviceName)

	if pubErr := c.tp.Publish(reqTopic, reqBytes); pubErr != nil {
		c.mu.Lock()
		delete(c.pending, requestID)
		c.mu.Unlock()
		call.timer.Stop()
		return nil, fmt.Errorf("courier/rpc: publish failed: %w", pubErr)
	}

	// Retry loop.
	interval := c.retryInterval
	for attempt := 0; attempt < c.retryCount; attempt++ {
		select {
		case resp := <-call.respChan:
			return c.handleCallResult(resp)
		case <-ctx.Done():
			return nil, ErrCanceled
		case <-time.After(interval):
			// Retry: re-publish the same request.
			_ = c.tp.Publish(reqTopic, reqBytes)
			interval = time.Duration(float64(interval) * c.retryBackoff)
		}
	}

	// Final wait for response (no more retries).
	select {
	case resp := <-call.respChan:
		return c.handleCallResult(resp)
	case <-ctx.Done():
		return nil, ErrCanceled
	}
}

func (c *Client) handleCallResult(resp []byte) ([]byte, error) {
	if resp == nil {
		return nil, ErrTimeout
	}
	return resp, nil
}

// handleResponse is the transport message handler for the response topic.
func (c *Client) handleResponse(topic string, payload []byte) {
	frame, err := codec.DecodeResponse(payload)
	if err != nil {
		log.Printf("[courier/rpc] failed to decode response: %v", err)
		return
	}

	c.mu.Lock()
	call, ok := c.pending[frame.RequestID]
	if ok {
		delete(c.pending, frame.RequestID)
		call.timer.Stop()
	}
	c.mu.Unlock()

	if !ok {
		return
	}

	select {
	case call.respChan <- frame.Payload:
	default:
		log.Printf("[courier/rpc] response channel full for request %x", frame.RequestID)
	}
}

// newRequestID generates a random 16-byte request identifier.
func newRequestID() ([16]byte, error) {
	var id [16]byte
	_, err := rand.Read(id[:])
	return id, err
}
