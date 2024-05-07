// Package web is a simple web framework package
package web

import (
	"context"
	"errors"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
)

type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

type App struct {
	mux        *chi.Mux
	shutdownCh chan os.Signal
	mws        []Middleware
}

func (a *App) SignalShutdown() {
	a.shutdownCh <- SIGWEB
}

func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	a.mux.ServeHTTP(w, r)
}

func (a *App) Handle(method, path string, handler Handler, mws ...Middleware) {
	handler = wrapMiddleware(handler, mws...)
	handler = wrapMiddleware(handler, a.mws...)
	// handler = wrapMiddleware(handler, append(a.mws, mws...)...)

	a.mux.MethodFunc(method, path, func(w http.ResponseWriter, r *http.Request) {
		// TODO move to override function
		// check if traceID is provided by the web request
		traceID := r.Header.Get("x-trace-id")
		if traceID == "" {
			traceID = uuid.NewString()
		}

		ctx := SetValues(r.Context(), &Values{
			TraceID: traceID,
			Now:     time.Now().UTC(),
		})

		if err := handler(ctx, w, r); err != nil {
			if validateShutdown(err) {
				a.SignalShutdown()
				return
			}
		}
	})
}

func NewApp(ch chan os.Signal, corsOpts CORSOptions, mws ...Middleware) *App {
	mux := chi.NewMux()

	// mux.Use adds middleware specific to chi
	// adding more middleware with mux.Use will load middleware BEFORE any
	// middlware created to fit the foundation/web format.
	mux.Use(cors.Handler(cors.Options{
		AllowedOrigins: corsOpts.AllowedOrigins,
		AllowedMethods: corsOpts.AllowedMethods,
		AllowedHeaders: corsOpts.AllowedHeaders,
		MaxAge:         corsOpts.MaxAge,
	}))

	return &App{
		mux:        mux,
		shutdownCh: ch,
		mws:        mws,
	}
}

// validateShutdown validates the error for special conditions that do not
// warrant an actual shutdown by the system.
// https://github.com/ardanlabs/service5-video/blob/main/foundation/web/web.go#L88
func validateShutdown(err error) bool {

	// Ignore syscall.EPIPE and syscall.ECONNRESET errors which occurs
	// when a write operation happens on the http.ResponseWriter that
	// has simultaneously been disconnected by the client (TCP
	// connections is broken). For instance, when large amounts of
	// data is being written or streamed to the client.
	// https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
	// https://gosamples.dev/broken-pipe/
	// https://gosamples.dev/connection-reset-by-peer/

	switch {
	case errors.Is(err, syscall.EPIPE):

		// Usually, you get the broken pipe error when you write to the connection after the
		// RST (TCP RST Flag) is sent.
		// The broken pipe is a TCP/IP error occurring when you write to a stream where the
		// other end (the peer) has closed the underlying connection. The first write to the
		// closed connection causes the peer to reply with an RST packet indicating that the
		// connection should be terminated immediately. The second write to the socket that
		// has already received the RST causes the broken pipe error.
		return false

	case errors.Is(err, syscall.ECONNRESET):

		// Usually, you get connection reset by peer error when you read from the
		// connection after the RST (TCP RST Flag) is sent.
		// The connection reset by peer is a TCP/IP error that occurs when the other end (peer)
		// has unexpectedly closed the connection. It happens when you send a packet from your
		// end, but the other end crashes and forcibly closes the connection with the RST
		// packet instead of the TCP FIN, which is used to close a connection under normal
		// circumstances.
		return false
	}

	return true
}
