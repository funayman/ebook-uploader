package mid

import (
	"context"
	"net/http"

	"github.com/funayman/simple-file-uploader/web"
)

func LimitBodySize(maxSize int64) web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
			if r.ContentLength > maxSize {
				return &http.MaxBytesError{Limit: maxSize}
			}

			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			return next(ctx, w, r)
		}
	}
}
