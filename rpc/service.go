package rpc

// MethodInfo describes a single RPC method within a service.
type MethodInfo struct {
	// Cmd is the numeric command identifier used for routing.
	Cmd uint32

	// Name is the human-readable method name.
	Name string

	// Handle is the function that processes requests for this method.
	Handle HandlerFunc
}

// ServiceInfo describes an RPC service and all its methods.
type ServiceInfo struct {
	// ServiceName is used as the MQTT topic segment and $share group name.
	ServiceName string

	// Methods lists all RPC methods provided by this service.
	Methods []MethodInfo
}
