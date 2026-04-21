package rpc

import "fmt"

// dispatcher routes incoming requests to registered handlers by command ID.
type dispatcher struct {
	handlers map[uint32]handlerEntry
}

type handlerEntry struct {
	serviceName string
	methodName  string
	handle      HandlerFunc
}

func newDispatcher() *dispatcher {
	return &dispatcher{
		handlers: make(map[uint32]handlerEntry),
	}
}

func (d *dispatcher) register(info ServiceInfo) {
	for _, m := range info.Methods {
		if _, exists := d.handlers[m.Cmd]; exists {
			panic(fmt.Sprintf("courier/rpc: duplicate command %d", m.Cmd))
		}
		d.handlers[m.Cmd] = handlerEntry{
			serviceName: info.ServiceName,
			methodName:  m.Name,
			handle:      m.Handle,
		}
	}
}

// dispatch looks up the handler for cmd, creates a Context, and calls the handler.
// Returns the response payload bytes or an error.
func (d *dispatcher) dispatch(cmd uint32, ctx *Context, payload []byte) ([]byte, error) {
	entry, ok := d.handlers[cmd]
	if !ok {
		return nil, ErrNotImplemented
	}

	ctx.ServiceName = entry.serviceName
	ctx.MethodName = entry.methodName

	return entry.handle(ctx, payload)
}
