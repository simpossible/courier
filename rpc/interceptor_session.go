package rpc

import "time"

// timeNow is a variable for testing purposes.
var timeNow = time.Now

// SessionInterceptor returns an interceptor that looks up the device's session
// from the SessionStore and injects it into the Context.
// Commands listed in publicCmds skip the session lookup.
func SessionInterceptor(store SessionStore, publicCmds map[uint32]bool) Interceptor {
	return func(next HandlerFunc) HandlerFunc {
		return func(ctx *Context, raw []byte) ([]byte, error) {
			if publicCmds[ctx.Cmd] {
				return next(ctx, raw)
			}

			if ctx.DeviceID == "" {
				return nil, NewError(401, "missing device id")
			}

			sess, err := store.Get(ctx.DeviceID)
			if err != nil {
				return nil, NewError(500, "session store error")
			}
			if sess == nil {
				return nil, NewError(401, "not logged in")
			}

			ctx.Session = sess

			// Touch last active time.
			sess.LastActive = timeNow()
			_ = store.Set(ctx.DeviceID, sess)

			return next(ctx, raw)
		}
	}
}

// PublicCmds is a convenience constructor for the public command map.
func PublicCmds(cmds ...uint32) map[uint32]bool {
	m := make(map[uint32]bool, len(cmds))
	for _, c := range cmds {
		m[c] = true
	}
	return m
}
