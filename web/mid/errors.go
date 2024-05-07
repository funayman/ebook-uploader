package mid

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"go.uber.org/zap"

	"github.com/funayman/simple-file-uploader/web"
	"github.com/inhies/go-bytesize"
)

func Errors(log *zap.SugaredLogger) web.Middleware {
	return func(next web.Handler) web.Handler {
		return func(ctx context.Context, w http.ResponseWriter, r *http.Request) (err error) {
			if err := next(ctx, w, r); err != nil {
				log.Errorw("web error", "error", err)

				var status int
				var output any

				var mbe *http.MaxBytesError
				switch {
				case errors.As(err, &mbe):
					status = 400
					output = fmt.Sprintf("max upload size is %s", bytesize.ByteSize(mbe.Limit).String())
				// case validate.IsFieldErrors(err):
				// 	status = http.StatusBadRequest

				// 	errs := validate.GetFieldErrors(err)
				// 	// quick and dirty
				// 	output = map[string]any{
				// 		"error":  "data validation error",
				// 		"fields": errs.Fields(),
				// 	}
				// case errors.Is(err, event.ErrNotFound):
				// 	status = http.StatusNotFound
				// 	output = map[string]any{
				// 		"error": "not found",
				// 	}
				case errors.Is(err, ErrBasicUnauthorized):
					w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
					status = http.StatusUnauthorized
					output = err.Error()
				default:
					status = http.StatusInternalServerError
					output = "internal server error"
				}

				if err := web.RespondJSON(ctx, w, output, status); err != nil {
					return err
				}

				// TODO check for shutdown errors
			}

			return nil
		}
	}
}
