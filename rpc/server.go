package rpc

import (
	"fmt"
	"log"

	"github.com/simpossible/courier/codec"
	"github.com/simpossible/courier/transport"
)

// Server receives RPC requests over MQTT, dispatches them to registered handlers,
// and sends responses back to the requesting client.
type Server struct {
	tp              transport.Transport
	serviceName     string
	sharedSubscribe bool
	interceptors    []Interceptor

	dispatcher *dispatcher
}

// NewServer creates a new RPC server with the given options.
func NewServer(opts ...ServerOption) *Server {
	s := &Server{
		sharedSubscribe: true,
		dispatcher:      newDispatcher(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// Register adds a service's methods to the server's dispatch table.
// Must be called before Start.
func (s *Server) Register(info ServiceInfo) {
	chain := ChainInterceptors(s.interceptors...)

	for i, m := range info.Methods {
		wrapped := m.Handle
		if len(s.interceptors) > 0 {
			wrapped = chain(m.Handle)
		}
		info.Methods[i].Handle = wrapped
	}

	s.dispatcher.register(info)
}

// Start connects the transport and subscribes to the request topic.
func (s *Server) Start() error {
	if s.tp == nil {
		return fmt.Errorf("courier/rpc: server has no transport")
	}
	if s.serviceName == "" {
		return fmt.Errorf("courier/rpc: server has no service name")
	}

	topic := s.requestTopic()
	respHandler := s.makeMessageHandler()

	if err := s.tp.Subscribe(topic, respHandler); err != nil {
		return fmt.Errorf("courier/rpc: subscribe to %s failed: %w", topic, err)
	}

	log.Printf("[courier/rpc] server subscribed to %s", topic)
	return nil
}

// Stop unsubscribes and closes the transport.
func (s *Server) Stop() error {
	if s.tp == nil {
		return nil
	}
	topic := s.requestTopic()
	_ = s.tp.Unsubscribe(topic)
	return s.tp.Close()
}

func (s *Server) requestTopic() string {
	if s.sharedSubscribe {
		return SharedRequestTopic(s.serviceName)
	}
	return RequestTopic(s.serviceName)
}

func (s *Server) makeMessageHandler() transport.MessageHandler {
	return func(topic string, payload []byte) {
		frame, err := codec.DecodeRequest(payload)
		if err != nil {
			log.Printf("[courier/rpc] failed to decode request: %v", err)
			return
		}

		ctx := &Context{}

		respPayload, dispatchErr := s.dispatcher.dispatch(frame.Cmd, ctx, frame.Payload)

		// Build and send response even on error — the client is waiting.
		var respBytes []byte
		if dispatchErr != nil {
			respBytes = errorToBytes(dispatchErr)
		} else {
			respBytes = respPayload
		}

		respFrame := codec.EncodeResponse(ctx.RequestID, respBytes)
		respTopic := ResponseTopic(ctx.DeviceID)

		if pubErr := s.tp.Publish(respTopic, respFrame); pubErr != nil {
			log.Printf("[courier/rpc] failed to publish response to %s: %v", respTopic, pubErr)
		}
	}
}

// errorToBytes serializes an error into a simple byte representation.
// The protoc-gen-courier plugin generates adapters that handle proper
// protobuf error encoding. For raw handler errors, we provide a basic fallback.
func errorToBytes(err error) []byte {
	if rpcErr, ok := err.(*Error); ok {
		return []byte(fmt.Sprintf(`{"code":%d,"msg":"%s"}`, rpcErr.Code, rpcErr.Message))
	}
	return []byte(fmt.Sprintf(`{"code":-2,"msg":"%s"}`, err.Error()))
}
