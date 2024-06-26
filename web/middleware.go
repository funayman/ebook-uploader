package web

// Middleware is a function designed to run some code before and/or after
// another Handler. It is designed to remove boilerplate or other concerns not
// direct to any given Handler.
type Middleware func(Handler) Handler

// wrapMiddleware creates a new handler by wrapping middleware around a final
// handler. The middlewares' Handlers will be executed by requests in the order
// they are provided.
func wrapMiddleware(h Handler, mws ...Middleware) Handler {
	// loop backwards through the middleware invoking each one. replace the
	// handler with the new wrapped handler. looping backwards ensures that the
	// first middleware of the slice is the first to be executed by requests.
	for i := len(mws) - 1; i >= 0; i-- {
		if mw := mws[i]; mw != nil {
			h = mw(h)
		}
	}
	return h
}
