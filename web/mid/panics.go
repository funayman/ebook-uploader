package mid

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/funayman/ebook-uploader/web"
)

func Panics() web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
			defer func() {
				if rec := recover(); rec != nil {
					trace := debug.Stack()
					err = fmt.Errorf("PANIC [%v] TRACE[%s]", rec, string(trace))
				}
			}()

			return next(ctx, w, r)
		}
	}
}
