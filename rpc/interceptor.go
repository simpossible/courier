package rpc

// Interceptor wraps a HandlerFunc to add cross-cutting behavior
// (logging, metrics, auth, recovery, etc.).
//
// Interceptors are applied in the order provided: the first interceptor
// in the slice is the outermost wrapper.
type Interceptor func(next HandlerFunc) HandlerFunc

// ChainInterceptors combines multiple interceptors into one.
// The first interceptor in the list is called first (outermost).
func ChainInterceptors(interceptors ...Interceptor) Interceptor {
	switch len(interceptors) {
	case 0:
		return func(next HandlerFunc) HandlerFunc { return next }
	case 1:
		return interceptors[0]
	default:
		return func(next HandlerFunc) HandlerFunc {
			for i := len(interceptors) - 1; i >= 0; i-- {
				next = interceptors[i](next)
			}
			return next
		}
	}
}
