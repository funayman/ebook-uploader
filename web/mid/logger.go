package mid

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/funayman/ebook-uploader/web"
)

func Logger(log *zap.SugaredLogger) web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			v := web.GetValues(ctx)

			path := r.URL.Path
			if r.URL.RawQuery != "" {
				path = fmt.Sprintf("%s?%s", path, r.URL.RawQuery)
			}

			log.Infow("request started",
				"trace_id", v.TraceID,
				"method", r.Method,
				"host", r.Host,
				"path", path,
				"remote_addr", r.RemoteAddr,
			)

			defer func() {
				log.Infow("request completed",
					"trace_id", v.TraceID,
					"method", r.Method,
					"host", r.Host,
					"path", path,
					"remote_addr", r.RemoteAddr,
					"status_code", v.StatusCode,
					"since", time.Since(v.Now).String(),
				)
			}()

			return next(ctx, w, r)
		}
	}
}
